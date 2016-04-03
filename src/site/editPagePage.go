// editPagePage.go serves the editPage.tmpl.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

var editPagePage = newPage(editPageRenderer, dynamicTmpls)

// editPageRenderer renders the page page.
func editPageRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// If it's not a page id but an alias, the redirect
	pageAlias := mux.Vars(params.R)["alias"]
	if len(pageAlias) > 0 {
		// Get page id
		pageId, ok, err := core.LoadAliasToPageId(db, pageAlias)
		if err != nil {
			return pages.Fail("Couldn't convert alias", err)
		} else if !ok {
			// Couldn't find alias, so let's create a page with it
			return pages.RedirectWith(core.GetNewPageUrl(pageAlias))
		}

		// Check if we need to redirect.
		var seeGroupId string
		row := database.NewQuery(`
			SELECT seeGroupId
			FROM pageInfos
			WHERE currentEdit>0 AND pageId=?`, pageId).ToStatement(db).QueryRow()
		exists, err := row.Scan(&seeGroupId)
		if err != nil {
			return pages.Fail("Couldn't get page info", err)
		}

		// Check if a subdomain redirect is necessary.
		if exists && seeGroupId != params.PrivateGroupId {
			subdomain := ""
			if core.IsIdValid(seeGroupId) {
				row := database.NewQuery(`
					SELECT alias
					FROM pages
					WHERE pageId=? and isLiveEdit`, seeGroupId).ToStatement(db).QueryRow()
				exists, err := row.Scan(&subdomain)
				if err != nil || !exists {
					return pages.Fail("Failed to redirect to subdomain", err)
				}
			}
			return pages.RedirectWith(core.GetPageFullUrl(subdomain, pageId))
		}
	}

	return pages.StatusOK(nil)
}
