// Package site is used to manage our website
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"zanaduu3/src/logger"
	"zanaduu3/src/sessions"

	"github.com/gorilla/mux"
)

var (
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
	w.WriteHeader(http.StatusOK)
}

func init() {
	logger.SetLogger(func(r *http.Request) logger.Logger {
		return appengine.NewContext(r)
	})

	r := mux.NewRouter()
	r.StrictSlash(true)

	// Public facing handlers for pages
	r.HandleFunc(domainIndexPage.URI, stdHandler(domainIndexPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(domainsPage.URI, stdHandler(domainsPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(editPagePage.URI, stdHandler(editPagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(editPrivatePagePage.URI, stdHandler(editPrivatePagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(explorePage.URI, stdHandler(explorePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(exploreAllPage.URI, stdHandler(explorePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(groupsPage.URI, stdHandler(groupsPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(indexPage.URI, stdHandler(indexPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(newPagePage.URI, stdHandler(newPagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(signupPage.URI, stdHandler(signupPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(pagePage.URI, stdHandler(pagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(privatePagePage.URI, stdHandler(privatePagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(updatesPage.URI, stdHandler(updatesPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(userPage.URI, stdHandler(userPage.ServeHTTP)).Methods("GET", "HEAD")

	// JSON handlers (API)
	r.HandleFunc("/json/search/", searchJsonHandler).Methods("GET")
	r.HandleFunc("/json/pages/", pagesJsonHandler).Methods("GET")
	r.HandleFunc("/json/edit/", editJsonHandler).Methods("GET")
	r.HandleFunc("/json/children/", childrenJsonHandler).Methods("GET")
	r.HandleFunc("/json/parents/", parentsJsonHandler).Methods("GET")
	r.HandleFunc("/json/aliases/", aliasesJsonHandler).Methods("GET")

	// POST handlers (API)
	r.HandleFunc("/abandonPage/", abandonPageHandler).Methods("POST")
	r.HandleFunc("/editPage/", editPageHandler).Methods("POST")
	r.HandleFunc("/newSubscription/", newSubscriptionHandler).Methods("POST")
	r.HandleFunc("/newLike/", newLikeHandler).Methods("POST")
	r.HandleFunc("/newMember/", newMemberHandler).Methods("POST")
	r.HandleFunc("/newGroup/", newGroupHandler).Methods("POST")
	r.HandleFunc("/newVote/", newVoteHandler).Methods("POST")
	r.HandleFunc("/deleteSubscription/", deleteSubscriptionHandler).Methods("POST")
	r.HandleFunc("/deletePage/", deletePageHandler).Methods("POST")
	r.HandleFunc("/revertPage/", revertPageHandler).Methods("POST")

	// Admin stuff
	r.HandleFunc("/updatePageIndex/", updatePageIndexHandler).Methods("GET")
	r.HandleFunc("/updateMetadata/", updateMetadataHandler).Methods("GET")
	r.HandleFunc("/becomeUser/", becomeUserHandler).Methods("GET")

	// Various internal handlers
	r.HandleFunc("/mon", reportMonitoring).Methods("POST")
	r.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	// Error handlers
	r.HandleFunc(errorPage.URI, stdHandler(errorPage.ServeHTTP)).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(stdHandler(page404.ServeHTTP))

	http.Handle("/", r)
}

// writeJson converts the given map to JSON and writes it to the given writer.
func writeJson(w http.ResponseWriter, m interface{}) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("Error marshalling data into json:", err)
	}
	// Write some stuff for "JSON Vulnerability Protection"
	w.Write([]byte(")]}',\n"))
	w.Write(jsonData)
	return nil
}
