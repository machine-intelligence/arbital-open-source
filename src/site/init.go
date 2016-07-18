// Package site is used to manage our website
package site

import (
	"encoding/json"
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
		return sessions.NewContext(r)
	})

	r := mux.NewRouter()
	s := r.Host(fmt.Sprintf("{www:w?w?w?\\.?}{subdomain:%s}{subdomaindot:\\.?}%s", core.SubdomainAliasRegexpStr, sessions.GetMuxDomain())).Subrouter()
	s.StrictSlash(true)

	// Pages
	s.HandleFunc("/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/achievements/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/adminDashboard/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/dashboard/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/discussion/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/domains/{domain:%s}/", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/edit/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/edit/{alias:%s}/", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/edit/{alias:%s}/{alias2:%s}/", core.AliasRegexpStr, core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/explore/{alias:%s}/", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/groups/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/learn/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/learn/{pageId:%s}/", core.AliasRegexpStr), pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/learn/{pageAlias:%s}/{pageAlias2:%s}/", core.AliasRegexpStr, core.AliasRegexpStr), pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/login/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/maintain/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/newsletter/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/notifications/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/read/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/recentChanges/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/requisites/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/pages/{alias:%s}/", core.AliasRegexpStr),
		pageHandlerWrapper(&pagePage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/p/{alias:%s}/", core.AliasRegexpStr),
		pageHandlerWrapper(&pagePage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/p/{alias:%s}/{alias2:%s}/", core.AliasRegexpStr, core.AliasRegexpStr),
		pageHandlerWrapper(&pagePage)).Methods("GET", "HEAD")
	s.HandleFunc("/settings/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/signup/", pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/user/{alias:%s}/", core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc(fmt.Sprintf("/user/{alias:%s}/{alias2:%s}/", core.AliasRegexpStr, core.AliasRegexpStr),
		pageHandlerWrapper(&dynamicPage)).Methods("GET", "HEAD")
	s.HandleFunc("/verifyEmail/", pageHandlerWrapper(&verifyEmailPage)).Methods("GET", "HEAD")

	// JSON handlers (API)
	s.HandleFunc(adminDashboardPageHandler.URI, handlerWrapper(adminDashboardPageHandler)).Methods("POST")
	s.HandleFunc(alternatePagesHandler.URI, handlerWrapper(alternatePagesHandler)).Methods("POST")
	s.HandleFunc(approvePageToDomainHandler.URI, handlerWrapper(approvePageToDomainHandler)).Methods("POST")
	s.HandleFunc(approvePageEditProposalHandler.URI, handlerWrapper(approvePageEditProposalHandler)).Methods("POST")
	s.HandleFunc(bellUpdatesHandler.URI, handlerWrapper(bellUpdatesHandler)).Methods("POST")
	s.HandleFunc(childrenHandler.URI, handlerWrapper(childrenHandler)).Methods("POST")
	s.HandleFunc(commentThreadHandler.URI, handlerWrapper(commentThreadHandler)).Methods("POST")
	s.HandleFunc(continueWritingModeHandler.URI, handlerWrapper(continueWritingModeHandler)).Methods("POST")
	s.HandleFunc(dashboardPageHandler.URI, handlerWrapper(dashboardPageHandler)).Methods("POST")
	s.HandleFunc(defaultHandler.URI, handlerWrapper(defaultHandler)).Methods("POST")
	s.HandleFunc(deleteAnswerHandler.URI, handlerWrapper(deleteAnswerHandler)).Methods("POST")
	s.HandleFunc(deleteLensHandler.URI, handlerWrapper(deleteLensHandler)).Methods("POST")
	s.HandleFunc(deleteMemberHandler.URI, handlerWrapper(deleteMemberHandler)).Methods("POST")
	s.HandleFunc(deletePageHandler.URI, handlerWrapper(deletePageHandler)).Methods("POST")
	s.HandleFunc(deletePagePairHandler.URI, handlerWrapper(deletePagePairHandler)).Methods("POST")
	s.HandleFunc(deletePathPageHandler.URI, handlerWrapper(deletePathPageHandler)).Methods("POST")
	s.HandleFunc(deleteSearchStringHandler.URI, handlerWrapper(deleteSearchStringHandler)).Methods("POST")
	s.HandleFunc(discardPageHandler.URI, handlerWrapper(discardPageHandler)).Methods("POST")
	s.HandleFunc(discussionModeHandler.URI, handlerWrapper(discussionModeHandler)).Methods("POST")
	s.HandleFunc(dismissUpdateHandler.URI, handlerWrapper(dismissUpdateHandler)).Methods("POST")
	s.HandleFunc(domainPageHandler.URI, handlerWrapper(domainPageHandler)).Methods("POST")
	s.HandleFunc(editHandler.URI, handlerWrapper(editHandler)).Methods("POST")
	s.HandleFunc(editPageHandler.URI, handlerWrapper(editPageHandler)).Methods("POST")
	s.HandleFunc(editPageInfoHandler.URI, handlerWrapper(editPageInfoHandler)).Methods("POST")
	s.HandleFunc(explanationRequestHandler.URI, handlerWrapper(explanationRequestHandler)).Methods("POST")
	s.HandleFunc(exploreHandler.URI, handlerWrapper(exploreHandler)).Methods("POST")
	s.HandleFunc(feedbackHandler.URI, handlerWrapper(feedbackHandler)).Methods("POST")
	s.HandleFunc(forgotPasswordHandler.URI, handlerWrapper(forgotPasswordHandler)).Methods("POST")
	s.HandleFunc(groupsHandler.URI, handlerWrapper(groupsHandler)).Methods("POST")
	s.HandleFunc(hedonsModeHandler.URI, handlerWrapper(hedonsModeHandler)).Methods("POST")
	s.HandleFunc(indexHandler.URI, handlerWrapper(indexHandler)).Methods("POST")
	s.HandleFunc(intrasitePopoverHandler.URI, handlerWrapper(intrasitePopoverHandler)).Methods("POST")
	s.HandleFunc(learnHandler.URI, handlerWrapper(learnHandler)).Methods("POST")
	s.HandleFunc(lensHandler.URI, handlerWrapper(lensHandler)).Methods("POST")
	s.HandleFunc(loginHandler.URI, handlerWrapper(loginHandler)).Methods("POST")
	s.HandleFunc(logoutHandler.URI, handlerWrapper(logoutHandler)).Methods("POST")
	s.HandleFunc(mailchimpSignupHandler.URI, handlerWrapper(mailchimpSignupHandler)).Methods("POST")
	s.HandleFunc(maintenanceModeHandler.URI, handlerWrapper(maintenanceModeHandler)).Methods("POST")
	s.HandleFunc(marksHandler.URI, handlerWrapper(marksHandler)).Methods("POST")
	s.HandleFunc(mergeQuestionsHandler.URI, handlerWrapper(mergeQuestionsHandler)).Methods("POST")
	s.HandleFunc(moreRelationshipsHandler.URI, handlerWrapper(moreRelationshipsHandler)).Methods("POST")
	s.HandleFunc(newAnswerHandler.URI, handlerWrapper(newAnswerHandler)).Methods("POST")
	s.HandleFunc(newGroupHandler.URI, handlerWrapper(newGroupHandler)).Methods("POST")
	s.HandleFunc(newInviteHandler.URI, handlerWrapper(newInviteHandler)).Methods("POST")
	s.HandleFunc(newLensHandler.URI, handlerWrapper(newLensHandler)).Methods("POST")
	s.HandleFunc(newLikeHandler.URI, handlerWrapper(newLikeHandler)).Methods("POST")
	s.HandleFunc(newMarkHandler.URI, handlerWrapper(newMarkHandler)).Methods("POST")
	s.HandleFunc(newMemberHandler.URI, handlerWrapper(newMemberHandler)).Methods("POST")
	s.HandleFunc(newPageHandler.URI, handlerWrapper(newPageHandler)).Methods("POST")
	s.HandleFunc(newPageToDomainSubmissionHandler.URI, handlerWrapper(newPageToDomainSubmissionHandler)).Methods("POST")
	s.HandleFunc(newPagePairHandler.URI, handlerWrapper(newPagePairHandler)).Methods("POST")
	s.HandleFunc(newPathPageHandler.URI, handlerWrapper(newPathPageHandler)).Methods("POST")
	s.HandleFunc(newSearchStringHandler.URI, handlerWrapper(newSearchStringHandler)).Methods("POST")
	s.HandleFunc(newVoteHandler.URI, handlerWrapper(newVoteHandler)).Methods("POST")
	s.HandleFunc(newsletterHandler.URI, handlerWrapper(newsletterHandler)).Methods("POST")
	s.HandleFunc(parentsHandler.URI, handlerWrapper(parentsHandler)).Methods("POST")
	s.HandleFunc(parentsSearchHandler.URI, handlerWrapper(parentsSearchHandler)).Methods("POST")
	s.HandleFunc(pendingModeHandler.URI, handlerWrapper(pendingModeHandler)).Methods("POST")
	s.HandleFunc(primaryPageHandler.URI, handlerWrapper(primaryPageHandler)).Methods("POST")
	s.HandleFunc(readModeHandler.URI, handlerWrapper(readModeHandler)).Methods("POST")
	s.HandleFunc(recentChangesHandler.URI, handlerWrapper(recentChangesHandler)).Methods("POST")
	s.HandleFunc(recentRelationshipChangesHandler.URI, handlerWrapper(recentRelationshipChangesHandler)).Methods("POST")
	s.HandleFunc(requisitesHandler.URI, handlerWrapper(requisitesHandler)).Methods("POST")
	s.HandleFunc(resolveMarkHandler.URI, handlerWrapper(resolveMarkHandler)).Methods("POST")
	s.HandleFunc(resolveThreadHandler.URI, handlerWrapper(resolveThreadHandler)).Methods("POST")
	s.HandleFunc(revertPageHandler.URI, handlerWrapper(revertPageHandler)).Methods("POST")
	s.HandleFunc(searchHandler.URI, handlerWrapper(searchHandler)).Methods("POST")
	s.HandleFunc(sendSlackInviteHandler.URI, handlerWrapper(sendSlackInviteHandler)).Methods("POST")
	s.HandleFunc(settingsPageHandler.URI, handlerWrapper(settingsPageHandler)).Methods("POST")
	s.HandleFunc(signupHandler.URI, handlerWrapper(signupHandler)).Methods("POST")
	s.HandleFunc(similarPageSearchHandler.URI, handlerWrapper(similarPageSearchHandler)).Methods("POST")
	s.HandleFunc(startPathHandler.URI, handlerWrapper(startPathHandler)).Methods("POST")
	s.HandleFunc(titleHandler.URI, handlerWrapper(titleHandler)).Methods("POST")
	s.HandleFunc(unassessedPagesHandler.URI, handlerWrapper(unassessedPagesHandler)).Methods("POST")
	s.HandleFunc(updateLensNameHandler.URI, handlerWrapper(updateLensNameHandler)).Methods("POST")
	s.HandleFunc(updateLensOrderHandler.URI, handlerWrapper(updateLensOrderHandler)).Methods("POST")
	s.HandleFunc(updateMarkHandler.URI, handlerWrapper(updateMarkHandler)).Methods("POST")
	s.HandleFunc(updateMasteriesHandler.URI, handlerWrapper(updateMasteriesHandler)).Methods("POST")
	s.HandleFunc(updateMemberHandler.URI, handlerWrapper(updateMemberHandler)).Methods("POST")
	s.HandleFunc(updatePageObjectHandler.URI, handlerWrapper(updatePageObjectHandler)).Methods("POST")
	s.HandleFunc(updatePagePairHandler.URI, handlerWrapper(updatePagePairHandler)).Methods("POST")
	s.HandleFunc(updatePathOrderHandler.URI, handlerWrapper(updatePathOrderHandler)).Methods("POST")
	s.HandleFunc(updatePathHandler.URI, handlerWrapper(updatePathHandler)).Methods("POST")
	s.HandleFunc(updateSettingsHandler.URI, handlerWrapper(updateSettingsHandler)).Methods("POST")
	s.HandleFunc(updateSubscriptionHandler.URI, handlerWrapper(updateSubscriptionHandler)).Methods("POST")
	s.HandleFunc(updateUserTrustHandler.URI, handlerWrapper(updateUserTrustHandler)).Methods("POST")
	s.HandleFunc(userPopoverHandler.URI, handlerWrapper(userPopoverHandler)).Methods("POST")
	s.HandleFunc(userSearchHandler.URI, handlerWrapper(userSearchHandler)).Methods("POST")
	s.HandleFunc(writeNewModeHandler.URI, handlerWrapper(writeNewModeHandler)).Methods("POST")

	// Admin stuff
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
