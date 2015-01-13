// web.go is a simple monitoring dashboard for Xelaie.
//
// It shows:
//   * useful links
//   * prober results, if probing is enabled
package dash

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"zanaduu3/src/config"
	"zanaduu3/src/googleauth"
	"zanaduu3/src/monitoring/prober"

	"github.com/golang/glog"
	"github.com/hkjn/pages"
)

var (
	indexPage      pages.Page
	dashboardPage  pages.Page
	addr           = flag.String("addr", ":8080", "address to bind monitoring console to")
	proberDisabled = flag.Bool("no_prober", false, "disables prober")
	targetURL      = flag.String("target", "http://rewards.xelaie.com", "website to probe against")
	templatePath   = flag.String("template_path", "src/go/monitoring/dash/tmpl/", "path to look for templates under")
	authDisabled   = flag.Bool("no_auth", false, "disables authentication (use for testing only)")
	xc             = config.Load()
	logsURL        = "https://console.developers.google.com/project/exemplary-cycle-688/logs?service=App%20Engine&moduleId=default&versionId=001e&expandAll=false&logName=appengine.googleapis.com%2Frequest_log&lastVisibleOffset=543f91a000ff03811ce53ecb650001737e6578656d706c6172792d6379636c652d363838000130303165000100&minLogLevel=2"
	influxURL      = fmt.Sprintf("http://%s:8083/", xc.Vm.Monitoring.Address)
	probes         []*prober.Probe
	baseTmpls      = []string{"base.tmpl", "style.tmpl",
		"scripts.tmpl", "login.tmpl", "init_js.tmpl",
		"jquery.tmpl",
	}
)

// Run runs the dashboard.
func Run() {
	addPages()

	http.Handle(indexPage.URI, *requireLogin(&indexPage))
	http.Handle(dashboardPage.URI, *requireLogin(&dashboardPage))
	http.Handle("/connect", Handler(connect))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	if !*proberDisabled {
		glog.Infof("Starting prober against %q..\n", *targetURL)
		probes = []*prober.Probe{
			prober.NewRewardsProbe(*targetURL),
			prober.NewFailureRateProbe(),
			prober.NewIndexServedProbe(),
		}
		for _, p := range probes {
			go p.Run()
		}
	}

	glog.Infof("Monitoring dashboard is starting on %s..\n", *addr)
	glog.Fatal(http.ListenAndServe(*addr, nil))
}

// TODO: consider making this part of package pages.
// getTemplates returns the full template path from file names, also
// including the baseTmpls.
func getTemplates(tmpls ...string) []string {
	paths := []string{}
	for _, tmpl := range baseTmpls {
		paths = append(paths, filepath.Join(*templatePath, tmpl))
	}
	for _, tmpl := range tmpls {
		paths = append(paths, filepath.Join(*templatePath, tmpl))
	}
	return paths
}

// addPages initializes the pages we serve.
func addPages() {
	indexPage = pages.Add(
		"/",
		indexRenderer,
		getTemplates("index.tmpl", "links.tmpl", "prober.tmpl")...)
	indexPage = *requireLogin(&indexPage)
	dashboardPage = pages.Add(
		"/dash",
		dashboardRenderer,
		getTemplates("dash.tmpl")...)
}

func init() {
	flag.Parse()
}

// baseData is shared data across pages.
type baseData struct {
	LoginInfo *googleauth.LoginInfo
}

// requireLogin returns a new Page that first checks G+ login.
func requireLogin(p *pages.Page) *pages.Page {
	originRender := p.Render
	p.Render = func(w http.ResponseWriter, r *http.Request) pages.Result {
		l := pages.GetLogger(r)
		loggedIn := false
		var err error
		if *authDisabled {
			l.Infof("-disabled_auth, so not checking credentials")
			loggedIn = true
		} else {
			loggedIn, err = googleauth.IsLoggedIn(r)
			if err != nil {
				return pages.InternalErrorWith(fmt.Errorf("failed to get session: %v", err))
			}
		}
		if loggedIn {
			return originRender(w, r)
		}
		l.Debugf("not logged in, fetching state token\n")
		li, err := googleauth.CheckLogIn(w, r)
		if err != nil {
			l.Errorf("failed to get login info: %v\n", err)
			return pages.InternalErrorWith(err)
		}
		return pages.StatusOK(struct{ LoginInfo *googleauth.LoginInfo }{li})
	}
	return p
}

type linkInfo struct {
	Name, URL string
}

func indexRenderer(w http.ResponseWriter, r *http.Request) pages.Result {
	l := pages.GetLogger(r)
	l.Infof("rendering /..\n")

	data := struct {
		baseData
		ErrorMsg       string
		Links          []linkInfo
		Probes         []*prober.Probe
		ProberDisabled bool
	}{}
	data.Links = []linkInfo{
		linkInfo{"Dashboard", "/dash"},
		linkInfo{"App engine logs", logsURL},
		linkInfo{
			"InfluxDB (try e.g. 'select count(value) as \"events/min\" from /.*/ group by time(1m)')",
			influxURL,
		},
	}
	data.Probes = probes
	data.ProberDisabled = *proberDisabled
	return pages.StatusOK(data)
}

// connect exchanges the one-time authorization code for a token and stores the
// token in the session
func connect(w http.ResponseWriter, r *http.Request) *Error {
	err := googleauth.Connect(w, r)
	if err != nil {
		glog.Errorf("error connecting to googleauth: %v\n", err)
		return &Error{err, "Couldn't complete log in. Please email xelaie@googlegroups.com or join us in #xelaie @ irc.freenode.net.", http.StatusUnauthorized}
	}
	glog.Infof("current user already connected\n")
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	return nil
}
