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
	dynamicTmpls = []string{"tmpl/scripts.tmpl", "tmpl/style.tmpl", "tmpl/dynamicPage.tmpl"}
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
	s := r.Host(fmt.Sprintf("{www:w?w?w?\\.?}{subdomain:%s}{subdomaindot:\\.?}%s", core.SubdomainAliasRegexpStr, sessions.GetMuxDomain())).Subrouter()
	s.StrictSlash(true)

	// Pages
	s.HandleFunc("/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/dashboard/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/domains/{domain:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/edit/", pageHandlerWrapper(&editPagePage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/edit/{alias:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&editPagePage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/edit/{alias:%s}/{edit:[0-9]+}", core.AliasRegexpStr),
		pageHandlerWrapper(&editPagePage)).Methods("GET", "HEAD")
	s.HandleFunc("/explore/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/explore/{domain:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/groups/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/login/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/requisites/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/pages/{alias:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&pagePage)).Methods("GET", "HEAD")
	s.HandleFunc("/sequences/{pageId:[0-9]+}", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/settings/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/signup/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/updates/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/user/{domain:%s}", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/verifyEmail/", pageHandlerWrapper(&verifyEmailPage)).Methods("GET", "HEAD")

	// JSON handlers (API)
	s.HandleFunc(dashboardPageHandler.URI, handlerWrapper(dashboardPageHandler)).Methods("POST")
	s.HandleFunc(childrenHandler.URI, handlerWrapper(childrenHandler)).Methods("POST")
	s.HandleFunc(domainPageHandler.URI, handlerWrapper(domainPageHandler)).Methods("POST")
	s.HandleFunc(editHandler.URI, handlerWrapper(editHandler)).Methods("POST")
	s.HandleFunc(exploreHandler.URI, handlerWrapper(exploreHandler)).Methods("POST")
	s.HandleFunc(defaultHandler.URI, handlerWrapper(defaultHandler)).Methods("POST")
	s.HandleFunc(groupsHandler.URI, handlerWrapper(groupsHandler)).Methods("POST")
	s.HandleFunc(indexHandler.URI, handlerWrapper(indexHandler)).Methods("POST")
	s.HandleFunc(intrasitePopoverHandler.URI, handlerWrapper(intrasitePopoverHandler)).Methods("POST")
	s.HandleFunc(requisitesHandler.URI, handlerWrapper(requisitesHandler)).Methods("POST")
	s.HandleFunc(lensHandler.URI, handlerWrapper(lensHandler)).Methods("POST")
	s.HandleFunc(newPageHandler.URI, handlerWrapper(newPageHandler)).Methods("POST")
	s.HandleFunc(parentsHandler.URI, handlerWrapper(parentsHandler)).Methods("POST")
	s.HandleFunc(parentsSearchHandler.URI, handlerWrapper(parentsSearchHandler)).Methods("POST")
	s.HandleFunc(primaryPageHandler.URI, handlerWrapper(primaryPageHandler)).Methods("POST")
	s.HandleFunc(searchHandler.URI, handlerWrapper(searchHandler)).Methods("POST")
	s.HandleFunc(sequenceHandler.URI, handlerWrapper(sequenceHandler)).Methods("POST")
	s.HandleFunc(similarPageSearchHandler.URI, handlerWrapper(similarPageSearchHandler)).Methods("POST")
	s.HandleFunc(titleHandler.URI, handlerWrapper(titleHandler)).Methods("POST")
	s.HandleFunc(updatesHandler.URI, handlerWrapper(updatesHandler)).Methods("POST")
	s.HandleFunc(updateLensOrderHandler.URI, handlerWrapper(updateLensOrderHandler)).Methods("POST")
	s.HandleFunc(userPageHandler.URI, handlerWrapper(userPageHandler)).Methods("POST")
	s.HandleFunc(userPopoverHandler.URI, handlerWrapper(userPopoverHandler)).Methods("POST")
	s.HandleFunc(userSearchHandler.URI, handlerWrapper(userSearchHandler)).Methods("POST")

	// POST handlers (API)
	s.HandleFunc(discardPageHandler.URI, handlerWrapper(discardPageHandler)).Methods("POST")
	s.HandleFunc(deleteMemberHandler.URI, handlerWrapper(deleteMemberHandler)).Methods("POST")
	s.HandleFunc(deletePageHandler.URI, handlerWrapper(deletePageHandler)).Methods("POST")
	s.HandleFunc(deletePagePairHandler.URI, handlerWrapper(deletePagePairHandler)).Methods("POST")
	s.HandleFunc(deleteSubscriptionHandler.URI, handlerWrapper(deleteSubscriptionHandler)).Methods("POST")
	s.HandleFunc(editPageHandler.URI, handlerWrapper(editPageHandler)).Methods("POST")
	s.HandleFunc(editPageInfoHandler.URI, handlerWrapper(editPageInfoHandler)).Methods("POST")
	s.HandleFunc(loginHandler.URI, handlerWrapper(loginHandler)).Methods("POST")
	s.HandleFunc(logoutHandler.URI, handlerWrapper(logoutHandler)).Methods("POST")
	s.HandleFunc(newGroupHandler.URI, handlerWrapper(newGroupHandler)).Methods("POST")
	s.HandleFunc(newLikeHandler.URI, handlerWrapper(newLikeHandler)).Methods("POST")
	s.HandleFunc(newMemberHandler.URI, handlerWrapper(newMemberHandler)).Methods("POST")
	s.HandleFunc(newPagePairHandler.URI, handlerWrapper(newPagePairHandler)).Methods("POST")
	s.HandleFunc(newSubscriptionHandler.URI, handlerWrapper(newSubscriptionHandler)).Methods("POST")
	s.HandleFunc(newVoteHandler.URI, handlerWrapper(newVoteHandler)).Methods("POST")
	s.HandleFunc(revertPageHandler.URI, handlerWrapper(revertPageHandler)).Methods("POST")
	s.HandleFunc(feedbackHandler.URI, handlerWrapper(feedbackHandler)).Methods("POST")
	s.HandleFunc(signupHandler.URI, handlerWrapper(signupHandler)).Methods("POST")
	s.HandleFunc(updateMasteriesHandler.URI, handlerWrapper(updateMasteriesHandler)).Methods("POST")
	s.HandleFunc(updateMemberHandler.URI, handlerWrapper(updateMemberHandler)).Methods("POST")
	s.HandleFunc(updateSettingsHandler.URI, handlerWrapper(updateSettingsHandler)).Methods("POST")

	// Admin stuff
	s.HandleFunc(domainsPageHandler.URI, handlerWrapper(domainsPageHandler)).Methods("POST")
	s.HandleFunc(adminTaskHandler.URI, handlerWrapper(adminTaskHandler)).Methods("GET")

	// Various internal handlers
	s.HandleFunc("/mon", reportMonitoring).Methods("POST")
	s.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	// Error handlers
	s.NotFoundHandler = http.HandlerFunc(pageHandlerWrapper(&dynamicPage))

	// Raw handlers
	http.HandleFunc("/sendTestEmail/", sendTestEmailHandler)

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
