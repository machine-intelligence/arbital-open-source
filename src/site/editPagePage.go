// editPagePage.go serves the editPage.tmpl.
package site

import (
	"fmt"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
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
	db := params.DB

	var data commonPageData
	data.User = params.U

	// If it's not a page id but an alias, the redirect
	pageAlias := mux.Vars(params.R)["alias"]
	pageId, err := strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		row := db.NewStatement(`
			SELECT pageId
			FROM pages
			WHERE alias=? AND isCurrentEdit`).QueryRow(pageAlias)
		exists, err := row.Scan(&pageId)
		if err != nil {
			return pages.Fail("Couldn't convert alias=>pageId", err)
		} else if exists {
			return pages.RedirectWith(core.GetEditPageUrl(pageId))
		}
	}

	// Load all the groups.
	data.PageMap = make(map[int64]*core.Page)
	core.AddUserGroupIdsToPageMap(data.User, data.PageMap)
	err = core.LoadPages(db, data.PageMap, params.U.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	return pages.StatusOK(data)
}
