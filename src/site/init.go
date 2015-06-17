// Package site is used to manage our website
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"appengine"

	"zanaduu3/src/config"
	"zanaduu3/src/logger"
	"zanaduu3/src/sessions"

	"github.com/gorilla/mux"
)

var (
	xc        = config.Load()
	baseTmpls = []string{"tmpl/scripts.tmpl", "tmpl/style.tmpl"}
)

func getConfigAddress() string {
	address := xc.Site.Dev.Address
	if sessions.Live {
		return xc.Site.Live.Address
	}
	return strings.TrimPrefix(address, "http://")
}

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
	r.HandleFunc(editPagePage.URI, stdHandler(editPagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(editPrivatePagePage.URI, stdHandler(editPrivatePagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(explorePage.URI, stdHandler(explorePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(filterPage.URI, stdHandler(filterPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(groupsPage.URI, stdHandler(groupsPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(indexPage.URI, stdHandler(indexPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(newPagePage.URI, stdHandler(newPagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(signupPage.URI, stdHandler(signupPage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(pagePage.URI, stdHandler(pagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(privatePagePage.URI, stdHandler(privatePagePage.ServeHTTP)).Methods("GET", "HEAD")
	r.HandleFunc(updatesPage.URI, stdHandler(updatesPage.ServeHTTP)).Methods("GET", "HEAD")

	// JSON handlers (API)
	r.HandleFunc("/json/pages/", pagesJsonHandler).Methods("GET")
	r.HandleFunc("/json/children/", childrenJsonHandler).Methods("GET")
	r.HandleFunc("/json/parents/", parentsJsonHandler).Methods("GET")

	// POST handlers (API)
	r.HandleFunc("/editPage/", editPageHandler).Methods("POST")
	r.HandleFunc("/newComment/", newCommentHandler).Methods("POST")
	r.HandleFunc("/newSubscription/", newSubscriptionHandler).Methods("POST")
	r.HandleFunc("/newLike/", newLikeHandler).Methods("POST")
	r.HandleFunc("/newMember/", newMemberHandler).Methods("POST")
	r.HandleFunc("/newGroup/", newGroupHandler).Methods("POST")
	r.HandleFunc("/newTag/", newTagHandler).Methods("POST")
	r.HandleFunc("/newVote/", newVoteHandler).Methods("POST")
	r.HandleFunc("/pageInfo/", pageInfoHandler).Methods("POST")
	r.HandleFunc("/updateInput/", updateInputHandler).Methods("POST")
	r.HandleFunc("/updateComment/", updateCommentHandler).Methods("POST")
	r.HandleFunc("/updateCommentLike/", updateCommentLikeHandler).Methods("POST")
	r.HandleFunc("/deleteSubscription/", deleteSubscriptionHandler).Methods("POST")
	r.HandleFunc("/deletePage/", deletePageHandler).Methods("POST")

	// Admin stuff
	r.HandleFunc("/becomeUser/", becomeUserHandler).Methods("GET")

	// Various internal handlers
	r.HandleFunc("/mon", reportMonitoring).Methods("POST")
	r.HandleFunc("/_ah/start", ahHandler).Methods("GET")
	r.HandleFunc(errorPage.URI, stdHandler(errorPage.ServeHTTP)).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(stdHandler(page404.ServeHTTP))

	http.Handle("/", r)
}

// writeJson converts the given map to JSON and writes it to the given writer.
func writeJson(w http.ResponseWriter, m map[string]interface{}) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("Error marshalling data into json:", err)
	}
	// Write some stuff for "JSON Vulnerability Protection"
	w.Write([]byte(")]}',\n"))
	w.Write(jsonData)
	return nil
}
