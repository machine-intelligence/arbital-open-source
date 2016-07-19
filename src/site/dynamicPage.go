// dynamicPage.go serves a page which then loads more data dynamically.

package site

import (
	"zanaduu3/src/pages"
)

var (
	dynamicPage = newPage(dynamicPageRenderer, dynamicTmpls)
)

// dynamicPageRenderer renders the dynamic page.
func dynamicPageRenderer(params *pages.HandlerParams) *pages.Result {
	return pages.Success(nil)
}
