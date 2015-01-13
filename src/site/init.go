// Package site is used to manage our website
package site

import (
	"net/http"

	"zanaduu3/src/config"
	"zanaduu3/src/sessions"
	"zanaduu3/src/twitter"

	"github.com/gorilla/mux"
)

var (
	xc        = config.Load()
	baseTmpls = []string{"tmpl/base.tmpl", "tmpl/scripts.tmpl", "tmpl/style.tmpl", "tmpl/base_header.tmpl"}
)

// notFoundHandler serves HTTP 404 when no matching handler is registered.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: serve static error page here, once we have one.

	// Note that URLs that we have registered but that fail to match,
	// e.g. because of missing URL params or the wrong HTTP method also
	// end up here.
	c := sessions.NewContext(r)
	c.Warningf("Serving 404 on %q\n", r.URL)
	http.NotFound(w, r)
}

func ahHandler(w http.ResponseWriter, r *http.Request) {
	//c := sessions.NewContext(r)
	w.WriteHeader(http.StatusOK)
}

func init() {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	r.HandleFunc("/authorize_callback", handler(twitter.AuthHandler).monitor()).Methods("GET").Queries("oauth_token", "{token}", "oauth_verifier", "{verifier}")
	r.HandleFunc(oopsPage.URI, handler(oopsPage.ServeHTTP).monitor()).Methods("GET")
	r.HandleFunc(indexPage.URI, stdHandler(indexPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc("/mon", reportMonitoring).Methods("POST")
	r.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	http.Handle("/", r)
}
