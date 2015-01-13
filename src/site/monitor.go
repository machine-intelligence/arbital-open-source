// monitor.go: reports monitoring data to influxdb server.
package site

import (
	"fmt"
	"net/http"

	"appengine/urlfetch"

	"zanaduu3/src/sessions"
)

var monitoringAPI = getMonitoringAPI()

// getMonitoringAPI returns the URL to use when reporting monitoring data.
func getMonitoringAPI() string {
	host := fmt.Sprintf("http://%s:%s", xc.Vm.Monitoring.Address, "8086")
	monDb := xc.Monitoring.Influx.Database.Dev
	if sessions.Live {
		monDb = xc.Monitoring.Influx.Database.Live
	}
	return fmt.Sprintf(
		"%s/db/%s/series?u=%s&p=%s", host, monDb,
		xc.Monitoring.Influx.Monitoring.User, xc.Monitoring.Influx.Monitoring.Password)
}

// reportMonitoring handles the task queue for reporting of monitoring data.
func reportMonitoring(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	req, err := http.NewRequest("POST", monitoringAPI, r.Body)
	if err != nil {
		c.Errorf("failed to create monitoring API request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := urlfetch.Client(c)
	resp, err := client.Do(req)
	if err != nil {
		c.Errorf("failed to send monitoring API request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.Errorf("failed to report monitoring: %q from API %q\n", resp.Status, monitoringAPI)
		return
	}
	c.Debugf("successfully sent monitoring data\n")
}
