// context.go: handles xelaie contexts.
package sessions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"appengine"
	"appengine/taskqueue"
)

// requestID is the unique ID for a given request.
type requestID string

// contexts is a concurrency-safe map of context by request id.
type contexts struct {
	sync.RWMutex
	m map[requestID]Context
}

// change represents how a counter value changes (i.e. += 1).
type change struct {
	value int       // value to modify counter by
	last  time.Time // time of last change of value
}

// TODO: merge sessions into context.
// context represents the context of an in-flight HTTP request.
type Context struct {
	appengine.Context
	R        *http.Request
	counters map[string]change // monitoring counters to update for current request
	reported bool              // whether this context has been sent to collection
}

// NewContext returns a context for the request.
func NewContext(r *http.Request) Context {
	ac := appengine.NewContext(r)
	id := requestID(appengine.RequestID(ac))

	ctxs.RLock()
	c, ok := ctxs.m[id]
	ctxs.RUnlock()

	if ok {
		return c
	}

	c = Context{
		Context:  ac,
		R:        r,
		counters: map[string]change{},
		reported: false,
	}
	ctxs.Lock()
	ctxs.m[id] = c
	ctxs.Unlock()
	return c
}

// Inc increments a counter.
func (c *Context) Inc(ctr string) {
	value := c.counters[ctr].value
	c.counters[ctr] = change{
		value: value + 1,
		last:  time.Now(),
	}
}

// Report returns a task for reporting the monitoring info.
func (c Context) Report() (*taskqueue.Task, error) {
	if c.reported {
		// Only attempt to report once per context. If this happens, it's
		// a programming bug.
		return nil, fmt.Errorf("context for %q already reported", c.R.URL)
	}
	c.reported = true

	if len(c.counters) == 0 {
		return nil, nil
	}

	r := c.getReport()
	b, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal monitoring data: %v", err)
	}

	return &taskqueue.Task{
		Path:    "/mon",
		Payload: b,
		Method:  "POST",
	}, nil
}

// point represents a single data point to report.
type point []interface{}

// reportEntry is a single entry in the TSDB.
type reportEntry struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Points  []point  `json:"points"`
}

// getReport returns a report for the context in the format the monitoring API wants.
func (c Context) getReport() []reportEntry {
	r := make([]reportEntry, len(c.counters))
	i := 0
	for k, v := range c.counters {
		r[i] = reportEntry{
			Name:    k,
			Columns: []string{"time", "value"},
			Points: []point{{
				v.last.UnixNano() / 1e6,
				v.value}},
		}
		i += 1
	}
	return r
}
