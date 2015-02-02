// Package site is used to manage our website
package site

import (
	"net/http"

	"appengine"

	"zanaduu3/src/config"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"

	"github.com/gorilla/mux"
)

var (
	xc        = config.Load()
	baseTmpls = []string{"tmpl/scripts.tmpl", "tmpl/style.tmpl"}
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
	pages.SetLogger(func(r *http.Request) pages.Logger {
		return appengine.NewContext(r)
	})

	r := mux.NewRouter()
	r.StrictSlash(true)

	// Public facing handlers for pages
	r.HandleFunc(indexPage.URI, handler(indexPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(signupPage.URI, handler(signupPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(claimPage.URI, stdHandler(claimPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(privateClaimPage.URI, stdHandler(privateClaimPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(claimsPage.URI, stdHandler(claimsPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(updatesPage.URI, stdHandler(updatesPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(newClaimPage.URI, stdHandler(newClaimPage.ServeHTTP)).Methods("GET", "HEAD")

	// POST handlers (API)
	r.HandleFunc("/newClaim/", newClaimHandler).Methods("POST")
	r.HandleFunc("/newInput/", newInputHandler).Methods("POST")
	r.HandleFunc("/newComment/", newCommentHandler).Methods("POST")
	r.HandleFunc("/newSubscription/", newSubscriptionHandler).Methods("POST")
	r.HandleFunc("/updateQuestion/", updateQuestionHandler).Methods("POST")
	r.HandleFunc("/updateInput/", updateInputHandler).Methods("POST")
	r.HandleFunc("/updateComment/", updateCommentHandler).Methods("POST")
	r.HandleFunc("/priorVote/", priorVoteHandler).Methods("POST")
	r.HandleFunc("/deleteSubscription/", deleteSubscriptionHandler).Methods("POST")

	// Admin stuff
	r.HandleFunc("/becomeUser/", becomeUserHandler).Methods("GET")

	// Various internal handlers
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.HandleFunc(oopsPage.URI, handler(oopsPage.ServeHTTP).monitor()).Methods("GET")
	r.HandleFunc("/mon", reportMonitoring).Methods("POST")
	r.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	http.Handle("/", r)
}
