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

	// Pages
	r.HandleFunc(domainIndexPage.URI, pageHandlerWrapper(&domainIndexPage)).Methods("GET", "HEAD")
	r.HandleFunc(domainsPage.URI, pageHandlerWrapper(&domainsPage)).Methods("GET", "HEAD")
	r.HandleFunc(editPagePage.URI, pageHandlerWrapper(&editPagePage)).Methods("GET", "HEAD")
	r.HandleFunc(exploreAllPage.URI, pageHandlerWrapper(&explorePage)).Methods("GET", "HEAD")
	r.HandleFunc(explorePage.URI, pageHandlerWrapper(&explorePage)).Methods("GET", "HEAD")
	r.HandleFunc(groupsPage.URI, pageHandlerWrapper(&groupsPage)).Methods("GET", "HEAD")
	r.HandleFunc(indexPage.URI, pageHandlerWrapper(&indexPage)).Methods("GET", "HEAD")
	r.HandleFunc(newPagePage.URI, pageHandlerWrapper(&newPagePage)).Methods("GET", "HEAD")
	r.HandleFunc(signupPage.URI, pageHandlerWrapper(&signupPage)).Methods("GET", "HEAD")
	r.HandleFunc(pagePage.URI, pageHandlerWrapper(&pagePage)).Methods("GET", "HEAD")
	r.HandleFunc(updatesPage.URI, pageHandlerWrapper(&updatesPage)).Methods("GET", "HEAD")
	r.HandleFunc(userPage.URI, pageHandlerWrapper(&userPage)).Methods("GET", "HEAD")
	r.HandleFunc(settingsPage.URI, pageHandlerWrapper(&settingsPage)).Methods("GET", "HEAD")

	// JSON handlers (API)
	r.HandleFunc(childrenHandler.URI, handlerWrapper(childrenHandler)).Methods("POST")
	r.HandleFunc(editHandler.URI, handlerWrapper(editHandler)).Methods("POST")
	r.HandleFunc(groupsHandler.URI, handlerWrapper(groupsHandler)).Methods("POST")
	r.HandleFunc(intrasitePopoverHandler.URI, handlerWrapper(intrasitePopoverHandler)).Methods("POST")
	r.HandleFunc(lensHandler.URI, handlerWrapper(lensHandler)).Methods("POST")
	r.HandleFunc(newPageHandler.URI, handlerWrapper(newPageHandler)).Methods("POST")
	r.HandleFunc(parentsHandler.URI, handlerWrapper(parentsHandler)).Methods("POST")
	r.HandleFunc(parentsSearchHandler.URI, handlerWrapper(parentsSearchHandler)).Methods("POST")
	r.HandleFunc(primaryPageHandler.URI, handlerWrapper(primaryPageHandler)).Methods("POST")
	r.HandleFunc(searchHandler.URI, handlerWrapper(searchHandler)).Methods("POST")
	r.HandleFunc(similarPageSearchHandler.URI, handlerWrapper(similarPageSearchHandler)).Methods("POST")

	// POST handlers (API)
	r.HandleFunc(abandonPageHandler.URI, handlerWrapper(abandonPageHandler)).Methods("POST")
	r.HandleFunc(deleteMemberHandler.URI, handlerWrapper(deleteMemberHandler)).Methods("POST")
	r.HandleFunc(deletePageHandler.URI, handlerWrapper(deletePageHandler)).Methods("POST")
	r.HandleFunc(deletePagePairHandler.URI, handlerWrapper(deletePagePairHandler)).Methods("POST")
	r.HandleFunc(deleteSubscriptionHandler.URI, handlerWrapper(deleteSubscriptionHandler)).Methods("POST")
	r.HandleFunc(editPageHandler.URI, handlerWrapper(editPageHandler)).Methods("POST")
	r.HandleFunc(newGroupHandler.URI, handlerWrapper(newGroupHandler)).Methods("POST")
	r.HandleFunc(newLikeHandler.URI, handlerWrapper(newLikeHandler)).Methods("POST")
	r.HandleFunc(newMemberHandler.URI, handlerWrapper(newMemberHandler)).Methods("POST")
	r.HandleFunc(newPagePairHandler.URI, handlerWrapper(newPagePairHandler)).Methods("POST")
	r.HandleFunc(newSubscriptionHandler.URI, handlerWrapper(newSubscriptionHandler)).Methods("POST")
	r.HandleFunc(newVoteHandler.URI, handlerWrapper(newVoteHandler)).Methods("POST")
	r.HandleFunc(revertPageHandler.URI, handlerWrapper(revertPageHandler)).Methods("POST")
	r.HandleFunc(updateMasteryHandler.URI, handlerWrapper(updateMasteryHandler)).Methods("POST")
	r.HandleFunc(updateMemberHandler.URI, handlerWrapper(updateMemberHandler)).Methods("POST")
	r.HandleFunc(updateSettingsHandler.URI, handlerWrapper(updateSettingsHandler)).Methods("POST")

	// Admin stuff
	r.HandleFunc(fixTextHandler.URI, handlerWrapper(fixTextHandler)).Methods("GET")
	r.HandleFunc(updateElasticIndexHandler.URI, handlerWrapper(updateElasticIndexHandler)).Methods("GET")
	r.HandleFunc(updateMetadataHandler.URI, handlerWrapper(updateMetadataHandler)).Methods("GET")

	// Various internal handlers
	r.HandleFunc("/mon", reportMonitoring).Methods("POST")
	r.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	// Error handlers
	r.NotFoundHandler = http.HandlerFunc(pageHandlerWrapper(&page404))

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
