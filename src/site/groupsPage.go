// groupsPage.go serves the groups template.
package site

import (
	"zanaduu3/src/pages"
)

// groupsTmplData stores the data that we pass to the template to render the page
type groupsTmplData struct {
	commonPageData
}

// groupsPage serves the recent pages page.
var groupsPage = newPage(
	"/groups/",
	groupsRenderer,
	append(baseTmpls,
		"tmpl/groupsPage.tmpl", "tmpl/angular.tmpl.js"))

// groupsRenderer renders the page page.
func groupsRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U

	var data groupsTmplData
	data.User = u

	return pages.StatusOK(data)
}
