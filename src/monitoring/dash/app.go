// app.go is some tools for the monitoring app.
package dash

import (
	"net/http"

	"github.com/golang/glog"
)

// Handler is a function that handles a HTTP endpoint.
type Handler func(http.ResponseWriter, *http.Request) *Error

// Error is something going wrong when handling a HTTP request.
type Error struct {
	Err     error
	Message string
	Code    int
}

// serveHTTP formats and passes up an error
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil {
		glog.Errorf("%s %s: %v\n", r.Method, r.URL, e.Err)
		http.Error(w, e.Message, e.Code)
	}
}
