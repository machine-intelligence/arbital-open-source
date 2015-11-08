// error.go: static error page.
package site

import (
	"fmt"

	"zanaduu3/src/pages"
)

// everyThing serves the "error" page.
var page404 = newPage(
	renderer404,
	append(baseTmpls, "tmpl/errorPage.tmpl"))

// renderer404 renders the error page.
func renderer404(params *pages.HandlerParams) *pages.Result {
	return pages.Fail("Page not found :(", fmt.Errorf("Not found"))
}
