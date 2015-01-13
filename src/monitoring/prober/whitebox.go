// whitebox.go: Whitebox monitoring probes, that fail when some
// monitoring metric go too bad.
package prober

import (
	"fmt"
	"math"
	"time"

	"github.com/golang/glog"
	"github.com/influxdb/influxdb/client"
)

const (
	// TODO: the 1m lookback + grouping in these queries should be the
	// same as the `probe_interval` flag, and should be constructed
	// dynamically if that can be done without doesn't obscuring things too much.
	indexServedQuery = "SELECT COUNT(value) FROM index_page_served_success WHERE time > now() - 1m GROUP BY time(1m)"
	failQuery        = "SELECT COUNT(value) FROM /.*_fail/ WHERE time > now() - 1m GROUP BY time(1m)"
)

// Expected columns in timeseries data.
var dataColumns = []string{"time", "count"}

// FailureRateProbe probes the whitebox monitoring and fails if the
// rate of metrics ending in "fail" is too high.
type FailureRateProbe struct {
	maxFailRate int64 // max allowed failures / min
}

// Probe probes monitoring for high rate of failures.
func (p FailureRateProbe) Probe() error {
	return queryInflux(failQuery, 0, p.maxFailRate)
}

// NewFailureRateProbe returns a new instance of the probe.
func NewFailureRateProbe() *Probe {
	maxFailRate := int64(1)
	return new(&FailureRateProbe{maxFailRate}, "FailureRateProbe",
		fmt.Sprintf("Fires if rate of counters ending in 'fail' is > %d / min", maxFailRate))
}

// IndexServedProbe fails if the rate of index_page_served_success
// metric is too low.
type IndexServedProbe struct {
	minIndexRate int64 // min allowed rate of successful index page renders / min
}

// NewIndexServedProbe returns a new instance of the probe.
func NewIndexServedProbe() *Probe {
	minIndexRate := int64(1)
	return new(&IndexServedProbe{minIndexRate}, "IndexServedProbe",
		fmt.Sprintf("Fires if rate of index page renders is < %d / min", minIndexRate))
}

// Probe probes monitoring for low rate of successes.
func (p IndexServedProbe) Probe() error {
	return queryInflux(indexServedQuery, p.minIndexRate, math.MaxInt64)
}

// queryInflux queries InfluxDB with given query, returning error if
// a returned timeseries has any value outside given bounds.
func queryInflux(query string, min, max int64) error {
	c, err := client.NewClient(&client.ClientConfig{
		Host:     fmt.Sprintf("%s:8086", xc.Vm.Monitoring.Address),
		Username: xc.Monitoring.Influx.Monitoring.User,
		Password: xc.Monitoring.Influx.Monitoring.Password,
		Database: xc.Monitoring.Influx.Database.Live,
	})
	if err != nil {
		return fmt.Errorf("can't connect to influxdb: %v", err)
	}
	glog.V(2).Infof("querying influxdb: %q\n", query)
	r, err := c.Query(query, client.Second)
	if err != nil {
		return fmt.Errorf("influxdb query failed: %v", err)
	}
	glog.V(2).Infof("got %d results: %v\n", len(r), r)
	for _, s := range r {
		glog.V(2).Infof("reading time series %s\n", s.Name)
		if len(s.Columns) != len(dataColumns) || s.Columns[0] != dataColumns[0] || s.Columns[1] != dataColumns[1] {
			return fmt.Errorf("unexpected columns: want %v, got %v", dataColumns, s.Columns)
		}
		for _, p := range s.Points {
			ts, ok := p[0].(float64)
			if !ok {
				return fmt.Errorf("bad 'time' value: %+v", p[0])
			}
			t := time.Unix(int64(ts), 0).UTC()
			c, ok := p[1].(float64)
			if !ok {
				return fmt.Errorf("bad 'count' value: %+v", p[1])
			}
			glog.V(2).Infof("time: %+v, count: %+v\n", t, c)
			if int64(c) < min {
				return fmt.Errorf("value %q is too low at %v: %v, want >= %d", s.Name, t, c, min)
			} else if int64(c) > max {
				return fmt.Errorf("value %q is too high at %v: %v, want <= %d", s.Name, t, c, max)
			}
		}
	}
	return nil
}
