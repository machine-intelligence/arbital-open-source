// pagePage.go serves the page page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

// pagePage serves the page page.
var pagePage = newPage(pageRenderer, dynamicTmpls)

// pageRenderer renders the page page.
func pageRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// Get actual page id
	pageAlias := mux.Vars(params.R)["alias"]
	pageId, ok, err := core.LoadAliasToPageId(db, pageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.Fail("Couldn't find page", err)
	}
	// If the url has an actual alias, then redirect to use page id
	// TODO: do this on the FE
	if pageAlias != pageId {
		return pages.RedirectWith(core.GetPageUrl(pageId))
	}

	// Check if we need to redirect.
	var pageType string
	var seeGroupId string
	row := database.NewQuery(`
		SELECT type,seeGroupId
		FROM pageInfos
		WHERE currentEdit>0 AND pageId=?`, pageId).ToStatement(db).QueryRow()
	exists, err := row.Scan(&pageType, &seeGroupId)
	if err != nil {
		return pages.Fail("Couldn't get page info", err)
	}

	// Check if a subdomain redirect is necessary.
	if exists && seeGroupId != params.PrivateGroupId {
		subdomain := ""
		if core.IsIdValid(seeGroupId) {
			row := database.NewQuery(`
					SELECT alias
					FROM pageInfos
					WHERE pageId=? AND currentEdit>0`, seeGroupId).ToStatement(db).QueryRow()
			exists, err := row.Scan(&subdomain)
			if err != nil || !exists {
				return pages.Fail("Failed to redirect to subdomain", err)
			}
		}
		return pages.RedirectWith(core.GetPageFullUrl(subdomain, pageId))
	}

	return pages.StatusOK(nil)
}
