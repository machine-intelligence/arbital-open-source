// groupsPage.go serves the groups template.
package site

import (
	"zanaduu3/src/pages"
)

// groupsPage serves the recent pages page.
var groupsPage = newPage(
	"/groups/",
	groupsRenderer,
	append(baseTmpls,
		"tmpl/groupsPage.tmpl", "tmpl/angular.tmpl.js"))

// groupsRenderer renders the page page.
func groupsRenderer(params *pages.HandlerParams) *pages.Result {
	var data commonPageData
	data.User = params.U
	return pages.StatusOK(data)
}
