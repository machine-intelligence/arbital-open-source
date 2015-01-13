// Package prober provides simple black-box monitoring of rewards.xelaie.com.
//
// It goes through the sequence of sign-in with Twitter pages like a
// user would, and logs successes/failures, as well as optionally
// alerting via email.
package prober

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hkjn/pages"
	"github.com/hkjn/timeutils"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"gopkg.in/yaml.v2"
)

var (
	// To: addresses for alert emails.
	alertRecipient = "xelaie+alerts@googlegroups.com"
	// CC: addresses for alert emails.
	alertCCs = []string{
		"me@hkjn.me",
	}
	maxAlertFrequency = time.Minute * 15
	// From: address for alert emails.
	alertSender = "alerts-noreply@xelaie.com"
	// Message to send out in alerts.
	alertMsg       = "The probe <a href=\"%s\">%s</a> failed enough that this alert fired, as the arbitrary metric of 'badness' is %d, which we can all agree is a big number.<br/>The description of the probe is: &ldquo;%s&rdquo;<br/>Failure details follow:<br/>"
	logDir         = os.TempDir()          // default logging directory, can be overridden with -log_dir flag, defined by glog
	logName        = "prober.outcomes.log" // name of logging file
	alertThreshold = flag.Int("alert_threshold", 100, "level of 'badness' before alerting")
	alertsDisabled = flag.Bool("no_alerts", false, "disables alerts when probes fail too often")
	// Note: The Twitter API has a rate limiting of 15 requests / 15
	// minute period, per user context, so setting pause to < 1 min may
	// cause requests to be denied.
	probeInterval = flag.Duration("probe_interval", time.Second*61, "duration to pause between prober runs")
	minBadness    = 0 // minimum allowed value for `badness`
	onceOpen      sync.Once
	logFile       *os.File
	bufferSize    = 200 // maximum number of results per prober to keep
	sgClient      = sendgrid.NewSendGridClient(xc.Monitoring.Sendgrid.User, xc.Monitoring.Sendgrid.Password)
)

type disabledProbesT map[string]bool

var disabledProbes disabledProbesT = make(disabledProbesT)

// String returns the flag's value.
func (d *disabledProbesT) String() string {
	s := ""
	i := 0
	for p, _ := range *d {
		if i > 0 {
			s += ","
		}
		s += p
	}
	return s
}

// Get is part of the flag.Getter interface. It always returns nil for
// this flag type since the struct is not exported.
func (d *disabledProbesT) Get() interface{} {
	return nil
}

// Syntax: -disable_probes=FooProbe,BarProbe
func (d *disabledProbesT) Set(value string) error {
	vals := strings.Split(value, ",")
	m := *d
	for _, p := range vals {
		m[p] = true
	}
	return nil
}

// Prober is a mechanism that can probe some target.
type Prober interface {
	Probe() error // probe target once
}

// Probe is a stateful representation of repeated probe runs.
type Probe struct {
	Prober            // underlying prober mechanism
	Name, Desc string // name, description of the probe
	// If badness reaches alert threshold, an alert email is sent and
	// alertThreshold resets.
	Badness   int
	Alerting  bool      // whether this probe is currently alerting
	LastAlert time.Time // time of last alert sent, if any
	Disabled  bool      // whether this probe is disabled
	Records   Records   // records of probe runs
}

// State describes values that a probe may use internally to pass
// state around within a single Probe() execution, e.g. log-in
// cookies.
type PageState struct {
	vals      pages.Values
	cookieJar http.CookieJar
}

// new returns a new probe from given prober implementation.
func new(p Prober, name, desc string) *Probe {
	return &Probe{p, name, desc, minBadness, false, time.Time{}, false, Records{}}
}

// Run repeatedly runs the probe, blocking forever.
func (p *Probe) Run() {
	glog.Infof("[%s] Starting..\n", p.Name)

	for {
		if _, ok := disabledProbes[p.Name]; ok {
			p.Disabled = true
			glog.Infof("[%s] is disabled, will now exit", p.Name)
			return
		} else {
			p.runProbe()
		}
	}
}

// runProbe runs the probe once.
func (p *Probe) runProbe() {
	c := make(chan error, 1)
	start := time.Now().UTC()
	go func() {
		glog.Infof("[%s] Probing..\n", p.Name)
		c <- p.Probe()
	}()
	select {
	case err := <-c:
		p.handleResult(err)
		wait := *probeInterval - time.Since(start)
		glog.V(2).Infof("[%s] needs to sleep %v more here\n", p.Name, wait)
		time.Sleep(wait)
	case <-time.After(*probeInterval):
		// Probe didn't finish in time for us to run the next one, report as failure.
		glog.Errorf("[%s] Timed out\n", p.Name)
		p.handleResult(fmt.Errorf("%s timed out (with -probe_interval %1.1f sec)", p.Name, probeInterval.Seconds()))
	}
}

// fakeProbe pretends to run the probe once, flipping a coin for the results.
func (p *Probe) fakeProbe() {
	glog.Infof("[%s] Pretending to probe..", p.Name)
	var err error
	if rand.Intn(3) == 0 {
		err = fmt.Errorf("pretending that fake probe for %s failed", p.Name)
	}
	p.handleResult(err)

}

// add appends the record to the buffer for the probe, keeping it within bufferSize.
func (p *Probe) addRecord(r Record) {
	p.Records = append(p.Records, r)
	if len(p.Records) >= bufferSize {
		over := len(p.Records) - bufferSize
		glog.V(2).Infof("[%s] buffer is over %d, reslicing it\n", p.Name, bufferSize)
		p.Records = p.Records[over:]
	}
	glog.V(2).Infof("[%s] buffer is now %d elements\n", p.Name, len(p.Records))
}

// Records is a grouping of probe records that implements sort.Interface.
type Records []Record

func (pr Records) Len() int           { return len(pr) }
func (pr Records) Swap(i, j int)      { pr[i], pr[j] = pr[j], pr[i] }
func (pr Records) Less(i, j int) bool { return pr[i].Timestamp.Before(pr[j].Timestamp) }

// RecentFailures returns only recent probe failures among the records.
func (pr Records) RecentFailures() Records {
	failures := make(Records, 0)
	for _, r := range pr {
		if !r.Passed && !r.Timestamp.Before(time.Now().Add(-time.Hour)) {
			failures = append(failures, r)
		}
	}
	sort.Sort(sort.Reverse(failures))
	return failures
}

// Record is the result of a single probe run.
type Record struct {
	Timestamp  time.Time `yaml: "-"`
	TimeMillis string    // same as Timestamp but makes it into the YAML logs
	Passed     bool
	Details    string `yaml: "omitempty"`
}

// Ago describes the duration since the record occured.
func (r Record) Ago() string {
	return timeutils.DescDuration(time.Since(r.Timestamp))
}

// newRecord returns a new record.
func newRecord(passed bool, line string) Record {
	now := time.Now().UTC()
	return Record{
		Timestamp:  now,
		TimeMillis: now.Format(time.StampMilli),
		Passed:     passed,
		Details:    line,
	}
}

// Marshal returns the record in YAML form.
func (r Record) marshal() []byte {
	b, err := yaml.Marshal(r)
	if err != nil {
		glog.Fatalf("failed to marshal record %+v: %v", r, err)
	}
	return b
}

// openLog opens the log file.
func openLog() {
	logPath := filepath.Join(logDir, logName)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		glog.Fatalf("failed to open %q: %v\n", logPath, err)
	}
	logFile = f
}

// handleResult handles a return value from a probe() run.
func (p *Probe) handleResult(err error) {
	if err != nil {
		p.Badness += 10
		glog.Errorf("[%s] Failed while probing, badness is now %d: %v\n", p.Name, p.Badness, err)
		p.logFail(err)
	} else {
		if p.Badness > minBadness {
			p.Badness -= 1
		}
		glog.Infof("[%s] Pass, badness is now %d.\n", p.Name, p.Badness)
		p.logPass()
	}

	if p.Badness < *alertThreshold {
		p.Alerting = false
		return
	}

	p.Alerting = true
	if *alertsDisabled {
		glog.Infof("[%s] would now be alerting, but alerts are supressed\n", p.Name)
		return
	}

	glog.Infof("[%s] is alerting\n", p.Name)
	if time.Since(p.LastAlert) < maxAlertFrequency {
		glog.V(1).Infof("[%s] will not alert, since last alert was sent %v back\n", p.Name, time.Since(p.LastAlert))
		return
	}
	go func() {
		// Send alert email in goroutine to not block further probing.
		err := p.alert()
		if err != nil {
			glog.Errorf("[%s] failed to alert: %v", p.Name, err)
			// Note: We don't reset badness here, so next cycle we'll keep
			// trying to send the alert.
		} else {
			glog.Infof("[%s] sent alert email, resetting badness to 0\n", p.Name)
			p.Badness = 0
		}
	}()
}

// logFail logs a failed probe run.
func (p *Probe) logFail(err error) {
	onceOpen.Do(openLog)
	r := newRecord(false, err.Error())
	p.addRecord(r)
	_, err = logFile.Write(r.marshal())
	if err != nil {
		glog.Fatalf("failed to write failure to log: %v", err)
	}
}

// logPass logs a successful probe run.
func (p *Probe) logPass() {
	onceOpen.Do(openLog)
	r := newRecord(true, "")
	p.addRecord(r)
	_, err := logFile.Write(r.marshal())
	if err != nil {
		glog.Fatalf("failed to write success to log: %v", err)
	}
}

// alert sends an email alert.
func (p *Probe) alert() error {
	dashLink := fmt.Sprintf("http://j.mp/xeldash#%s", p.Name)
	msg := fmt.Sprintf(alertMsg, dashLink, p.Name, p.Badness, p.Desc)
	for _, rec := range p.Records {
		if !rec.Passed {
			msg += fmt.Sprintf("<h2>%v (%s)</h2><p>%s</p>", rec.Timestamp, rec.Ago(), rec.Details)
		}
	}
	subject := fmt.Sprintf("%s failed (badness %d)", p.Name, p.Badness)
	m := sendgrid.NewMail()
	err := m.AddTo(alertRecipient)
	if err != nil {
		return fmt.Errorf("failed to add recipients: %v", err)
	}
	err = m.AddCcs(alertCCs)
	if err != nil {
		return fmt.Errorf("failed to add cc recipients: %v", err)
	}
	m.SetSubject(subject)
	m.SetHTML(msg)
	err = m.SetFrom(alertSender)
	if err != nil {
		return fmt.Errorf("failed to add sender %q: %v", alertSender, err)
	}

	err = sgClient.Send(m)
	if err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}
	p.LastAlert = time.Now().UTC()
	return nil
}

func init() {
	if xelaieCookie == "" {
		glog.Fatalf("FATAL: missing site.cookie value in config.yaml\n")
	}
	flag.Var(&disabledProbes, "disabled_probes", "comma-separated list of probes to disable")
	// log_dir is defined by glog.go.
	logDirFlag := flag.Lookup("log_dir").Value.String()
	if logDirFlag != "" {
		logDir = logDirFlag
	}
	fmt.Printf("Prober starting.. further logging to %s\n", logDir)
}
