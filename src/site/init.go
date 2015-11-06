// Package site is used to manage our website
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"zanaduu3/src/core"
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
	s := r.Host(fmt.Sprintf("{subdomain:%s}{subdomaindot:\\.?}%s", core.SubdomainAliasRegexpStr, sessions.GetMuxDomain())).Subrouter()
	s.StrictSlash(true)

	// Pages
	r.HandleFunc(fmt.Sprintf("/domains/{domain:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	r.HandleFunc(domainsPage.URI, pageHandlerWrapper(&domainsPage)).Methods("GET", "HEAD")
	s.HandleFunc(editPagePage.URI, pageHandlerWrapper(&editPagePage)).Methods("GET", "HEAD")
	r.HandleFunc(fmt.Sprintf("/explore/{domain:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	r.HandleFunc("/explore/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/groups/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(newPagePage.URI, pageHandlerWrapper(&newPagePage)).Methods("GET", "HEAD")
	s.HandleFunc(signupPage.URI, pageHandlerWrapper(&signupPage)).Methods("GET", "HEAD")
	s.HandleFunc(pagePage.URI, pageHandlerWrapper(&pagePage)).Methods("GET", "HEAD")
	s.HandleFunc(updatesPage.URI, pageHandlerWrapper(&updatesPage)).Methods("GET", "HEAD")
	s.HandleFunc(userPage.URI, pageHandlerWrapper(&userPage)).Methods("GET", "HEAD")
	s.HandleFunc(settingsPage.URI, pageHandlerWrapper(&settingsPage)).Methods("GET", "HEAD")

	// JSON handlers (API)
	s.HandleFunc(childrenHandler.URI, handlerWrapper(childrenHandler)).Methods("POST")
	s.HandleFunc(domainIndexHandler.URI, handlerWrapper(domainIndexHandler)).Methods("POST")
	s.HandleFunc(editHandler.URI, handlerWrapper(editHandler)).Methods("POST")
	s.HandleFunc(exploreHandler.URI, handlerWrapper(exploreHandler)).Methods("POST")
	s.HandleFunc(groupsHandler.URI, handlerWrapper(groupsHandler)).Methods("POST")
	s.HandleFunc(indexHandler.URI, handlerWrapper(indexHandler)).Methods("POST")
	s.HandleFunc(intrasitePopoverHandler.URI, handlerWrapper(intrasitePopoverHandler)).Methods("POST")
	s.HandleFunc(lensHandler.URI, handlerWrapper(lensHandler)).Methods("POST")
	s.HandleFunc(newPageHandler.URI, handlerWrapper(newPageHandler)).Methods("POST")
	s.HandleFunc(parentsHandler.URI, handlerWrapper(parentsHandler)).Methods("POST")
	s.HandleFunc(parentsSearchHandler.URI, handlerWrapper(parentsSearchHandler)).Methods("POST")
	s.HandleFunc(primaryPageHandler.URI, handlerWrapper(primaryPageHandler)).Methods("POST")
	s.HandleFunc(privateIndexHandler.URI, handlerWrapper(privateIndexHandler)).Methods("POST")
	s.HandleFunc(searchHandler.URI, handlerWrapper(searchHandler)).Methods("POST")
	s.HandleFunc(similarPageSearchHandler.URI, handlerWrapper(similarPageSearchHandler)).Methods("POST")

	// POST handlers (API)
	s.HandleFunc(abandonPageHandler.URI, handlerWrapper(abandonPageHandler)).Methods("POST")
	s.HandleFunc(deleteMemberHandler.URI, handlerWrapper(deleteMemberHandler)).Methods("POST")
	s.HandleFunc(deletePageHandler.URI, handlerWrapper(deletePageHandler)).Methods("POST")
	s.HandleFunc(deletePagePairHandler.URI, handlerWrapper(deletePagePairHandler)).Methods("POST")
	s.HandleFunc(deleteSubscriptionHandler.URI, handlerWrapper(deleteSubscriptionHandler)).Methods("POST")
	s.HandleFunc(editPageHandler.URI, handlerWrapper(editPageHandler)).Methods("POST")
	s.HandleFunc(newGroupHandler.URI, handlerWrapper(newGroupHandler)).Methods("POST")
	s.HandleFunc(newLikeHandler.URI, handlerWrapper(newLikeHandler)).Methods("POST")
	s.HandleFunc(newMemberHandler.URI, handlerWrapper(newMemberHandler)).Methods("POST")
	s.HandleFunc(newPagePairHandler.URI, handlerWrapper(newPagePairHandler)).Methods("POST")
	s.HandleFunc(newSubscriptionHandler.URI, handlerWrapper(newSubscriptionHandler)).Methods("POST")
	s.HandleFunc(newVoteHandler.URI, handlerWrapper(newVoteHandler)).Methods("POST")
	s.HandleFunc(revertPageHandler.URI, handlerWrapper(revertPageHandler)).Methods("POST")
	s.HandleFunc(signupHandler.URI, handlerWrapper(signupHandler)).Methods("POST")
	s.HandleFunc(updateMasteryHandler.URI, handlerWrapper(updateMasteryHandler)).Methods("POST")
	s.HandleFunc(updateMemberHandler.URI, handlerWrapper(updateMemberHandler)).Methods("POST")
	s.HandleFunc(updateSettingsHandler.URI, handlerWrapper(updateSettingsHandler)).Methods("POST")

	// Admin stuff
	s.HandleFunc(fixTextHandler.URI, handlerWrapper(fixTextHandler)).Methods("GET")
	s.HandleFunc(updateElasticIndexHandler.URI, handlerWrapper(updateElasticIndexHandler)).Methods("GET")
	s.HandleFunc(updateMetadataHandler.URI, handlerWrapper(updateMetadataHandler)).Methods("GET")

	// Various internal handlers
	s.HandleFunc("/mon", reportMonitoring).Methods("POST")
	s.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	// Error handlers
	s.NotFoundHandler = http.HandlerFunc(pageHandlerWrapper(&page404))

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
