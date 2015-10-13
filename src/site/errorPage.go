// error.go: static error page.
package site

import (
	"zanaduu3/src/pages"
)

type errorData struct {
	commonPageData
	Error string
}

// errorPage serves the "error" page.
var errorPage = newPage(
	"/error/",
	nil,
	append(baseTmpls,
		"tmpl/errorPage.tmpl", "tmpl/angular.tmpl.js"))

// everyThing serves the "error" page.
var page404 = newPage(
	"/{catchall:.*}", // this is not used
	renderer404,
	append(baseTmpls, "tmpl/errorPage.tmpl"))

// renderer404 renders the error page.
func renderer404(params *pages.HandlerParams) *pages.Result {
	return pages.Fail("Page not found :(", nil)
}

// renderErrorPage is a custom fallback renderer so we can render an error
// message instead of the page.
func renderErrorPage(params *pages.HandlerParams, message string) *pages.Result {
	params.C.Inc("error_page_served_success")
	var data errorData
	data.Error = message
	data.User = params.U
	return pages.StatusOK(data)
}
