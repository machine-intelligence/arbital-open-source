// Logging support for pages.
//
// The default logger uses github.com/golang/glog, which requires
// that flags are initialized to override default behavior:
//   import "flag"
//
//   func init() {
//     flag.Parse()
//   }
//
// See https://godoc.org/github.com/golang/glog for full docs.
//
// Example using appengine's logging features:
//   pages.SetLogger(func(r *http.Request) pages.Logger {
//     return appengine.NewContext(r)
//   })
package pages

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

var (
	defaultLogger            = func(*http.Request) Logger { return Glogger{} } // the default logger uses glog
	logger        LoggerFunc = defaultLogger                                   // function to retrieve logger from http.Request
)

// LoggerFunc returns a Logger from a http request.
type LoggerFunc func(*http.Request) Logger

// GetLogger returns the Logger for given request.
func GetLogger(r *http.Request) Logger {
	return logger(r)
}

// SetLogger registers the logger function.
//
// By default, the glog package is used for logging.
func SetLogger(l LoggerFunc) {
	logger = l
}

// Logger specifies logging functions.
//
// The interface is borrowed from AppEngine.
type Logger interface {
	// Debugf formats its arguments according to the format, analogous to fmt.Printf,
	// and records the text as a log message at Debug level.
	Debugf(format string, args ...interface{})

	// Infof is like Debugf, but at Info level.
	Infof(format string, args ...interface{})

	// Warningf is like Debugf, but at Warning level.
	Warningf(format string, args ...interface{})

	// Errorf is like Debugf, but at Error level.
	Errorf(format string, args ...interface{})

	// Criticalf is like Debugf, but at Critical level.
	Criticalf(format string, args ...interface{})
}

// Glogger implements interface Logger using the glog package.
type Glogger struct{}

func (g Glogger) Debugf(format string, args ...interface{}) {
	glog.V(1).Infof(format, args...)
}
func (g Glogger) Infof(format string, args ...interface{}) {
	glog.Infof(format, args...)
}
func (g Glogger) Warningf(format string, args ...interface{}) {
	glog.Warningf(format, args...)
}
func (g Glogger) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}
func (g Glogger) Criticalf(format string, args ...interface{}) {
	// Don't use glog.Fatalf; in a web app we don't want to bring down
	// everything.
	glog.Errorf(fmt.Sprintf("CRITICAL: %s", format), args...)
}
