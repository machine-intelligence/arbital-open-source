// editPagePage.go serves the editPage.tmpl.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

var (
	editPageTmpls   = append(baseTmpls, "tmpl/dynamicPage.tmpl", "tmpl/angular.tmpl.js")
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
	if len(pageAlias) > 0 {
		// Get page id
		pageId, ok, err := core.LoadAliasToPageId(db, pageAlias)
		if err != nil {
			return pages.Fail("Couldn't convert alias", err)
		}
		if !ok {
			return pages.Fail("Couldn't find alias", err)
		}
		if pageAlias != fmt.Sprintf("%d", pageId) {
			return pages.RedirectWith(core.GetEditPageUrl(pageId))
		}

		// Check if we need to redirect.
		var seeGroupId int64
		row := database.NewQuery(`
			SELECT seeGroupId
			FROM pages
			WHERE isCurrentEdit AND pageId=?`, pageId).ToStatement(db).QueryRow()
		exists, err := row.Scan(&seeGroupId)
		if err != nil {
			return pages.Fail("Couldn't get page info", err)
		}

		// Check if a subdomain redirect is necessary.
		if exists && seeGroupId != params.PrivateGroupId {
			subdomain := ""
			if seeGroupId > 0 {
				row := database.NewQuery(`
					SELECT alias
					FROM pages
					WHERE pageId=? and isCurrentEdit`, seeGroupId).ToStatement(db).QueryRow()
				exists, err := row.Scan(&subdomain)
				if err != nil || !exists {
					return pages.Fail("Failed to redirect to subdomain", err)
				}
			}
			return pages.RedirectWith(core.GetPageFullUrl(subdomain, pageId))
		}
	}

	// Load all the groups.
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	err := core.ExecuteLoadPipeline(db, params.U, data.PageMap, data.UserMap, data.MasteryMap)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.StatusOK(data)
}
