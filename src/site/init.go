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
	r.HandleFunc("/json/children/", handlerWrapper(childrenJsonHandler)).Methods("GET")
	r.HandleFunc("/json/edit/", handlerWrapper(editJsonHandler)).Methods("GET")
	r.HandleFunc("/json/groups/", handlerWrapper(groupsJsonHandler)).Methods("GET")
	r.HandleFunc("/json/newPage/", handlerWrapper(newPageJsonHandler)).Methods("GET")
	r.HandleFunc("/json/pages/", handlerWrapper(pagesJsonHandler)).Methods("GET")
	r.HandleFunc("/json/parents/", handlerWrapper(parentsJsonHandler)).Methods("GET")
	r.HandleFunc("/json/parentsSearch/", handlerWrapper(parentsSearchJsonHandler)).Methods("POST")
	r.HandleFunc("/json/search/", handlerWrapper(searchJsonHandler)).Methods("POST")
	r.HandleFunc("/json/similarPageSearch/", handlerWrapper(similarPageSearchJsonHandler)).Methods("POST")

	// POST handlers (API)
	r.HandleFunc("/abandonPage/", handlerWrapper(abandonPageHandler)).Methods("POST")
	r.HandleFunc("/deleteMember/", handlerWrapper(deleteMemberHandler)).Methods("POST")
	r.HandleFunc("/deletePage/", handlerWrapper(deletePageHandler)).Methods("POST")
	r.HandleFunc("/deletePagePair/", handlerWrapper(deletePagePairHandler)).Methods("POST")
	r.HandleFunc("/deleteSubscription/", handlerWrapper(deleteSubscriptionHandler)).Methods("POST")
	r.HandleFunc("/editPage/", handlerWrapper(editPageHandler)).Methods("POST")
	r.HandleFunc("/newGroup/", handlerWrapper(newGroupHandler)).Methods("POST")
	r.HandleFunc("/newLike/", handlerWrapper(newLikeHandler)).Methods("POST")
	r.HandleFunc("/newMember/", handlerWrapper(newMemberHandler)).Methods("POST")
	r.HandleFunc("/newPagePair/", handlerWrapper(newPagePairHandler)).Methods("POST")
	r.HandleFunc("/newSubscription/", handlerWrapper(newSubscriptionHandler)).Methods("POST")
	r.HandleFunc("/newVote/", handlerWrapper(newVoteHandler)).Methods("POST")
	r.HandleFunc("/revertPage/", handlerWrapper(revertPageHandler)).Methods("POST")
	r.HandleFunc("/updateMastery/", handlerWrapper(updateMasteryHandler)).Methods("POST")
	r.HandleFunc("/updateMember/", handlerWrapper(updateMemberHandler)).Methods("POST")
	r.HandleFunc("/updateSettings/", handlerWrapper(updateSettingsHandler)).Methods("POST")

	// Admin stuff
	r.HandleFunc("/fixText/", handlerWrapper(fixTextHandler)).Methods("GET")
	r.HandleFunc("/updateElasticIndex/", handlerWrapper(updateElasticIndexHandler)).Methods("GET")
	r.HandleFunc("/updateMetadata/", handlerWrapper(updateMetadataHandler)).Methods("GET")
	//r.HandleFunc("/sendTestEmail/", handlerWrapper(sendTestEmailHandler)).Methods("GET")

	// Various internal handlers
	r.HandleFunc("/mon", reportMonitoring).Methods("POST")
	r.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	// Error handlers
	r.NotFoundHandler = http.HandlerFunc(pageHandlerWrapper(&page404))

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
