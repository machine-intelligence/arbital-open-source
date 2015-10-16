// editPagePage.go serves the editPage.tmpl.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var (
	editPageTmpls   = append(baseTmpls, "tmpl/editPage.tmpl", "tmpl/angular.tmpl.js")
	editPageOptions = pages.PageOptions{RequireLogin: true}
)

// These pages serve the edit page, but vary slightly in the parameters they take in the url.
var newPagePage = newPageWithOptions("/edit/", editPageRenderer, editPageTmpls, editPageOptions)
var editPagePage = newPageWithOptions(fmt.Sprintf("/edit/{alias:%s}", core.AliasRegexpStr), editPageRenderer, editPageTmpls, editPageOptions)

// editPageRenderer renders the page page.
func editPageRenderer(params *pages.HandlerParams) *pages.Result {
	var data commonPageData
	data.User = params.U

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err := core.LoadGroupNames(params.DB, data.User, data.GroupMap)
	if err != nil {
		return pages.Fail("Couldn't load group names: %v", err)
	}

	return pages.StatusOK(data)
}
