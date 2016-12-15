// Package site is used to manage our website

package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/logger"
	"zanaduu3/src/sessions"

	"github.com/gorilla/mux"
)

var (
	dynamicTmpls = []string{"tmpl/dynamicPage.tmpl"}
)

func ahHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func init() {
	logger.SetLogger(func(r *http.Request) logger.Logger {
		return sessions.NewContext(r)
	})

	r := mux.NewRouter()
	s := r.Host(fmt.Sprintf("{www:w?w?w?\\.?}{subdomain:%s}{subdomaindot:\\.?}%s", core.SubdomainAliasOrPageIDRegexpStr, sessions.GetMuxDomain())).Subrouter()
	s.StrictSlash(true)

	// Pages
	s.HandleFunc("/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")

	// JSON handlers (API)
	s.HandleFunc("/_ah/start", ahHandler).Methods("GET")

	http.Handle("/", r)
}
