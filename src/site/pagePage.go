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
	u := params.U
	db := params.DB

	// Get actual page id
	pageAlias := mux.Vars(params.R)["alias"]
	pageID, ok, err := core.LoadAliasToPageID(db, u, pageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	} else if !ok {
		return pages.Fail("Couldn't find page", err)
	}

	// Check if we need to redirect.
	var pageType string
	var seeGroupID string
	row := database.NewQuery(`
		SELECT type,seeGroupId
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		WHERE pageId=?`, pageID).ToStatement(db).QueryRow()
	exists, err := row.Scan(&pageType, &seeGroupID)
	if err != nil {
		return pages.Fail("Couldn't get page info", err)
	}

	// Check if a subdomain redirect is necessary.
	if exists && seeGroupID != params.PrivateGroupID {
		subdomain := ""
		if core.IsIDValid(seeGroupID) {
			row := database.NewQuery(`
					SELECT alias
					FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
					WHERE pageId=?`, seeGroupID).ToStatement(db).QueryRow()
			exists, err := row.Scan(&subdomain)
			if err != nil || !exists {
				return pages.Fail("Failed to redirect to subdomain", err)
			}
		}
		return pages.RedirectWith(core.GetPageFullURL(subdomain, pageID))
	}

	p := core.NewPage(pageID)
	pageMap := map[string]*core.Page{
		pageID: p,
	}
	err = core.LoadPages(db, params.U, pageMap)
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	}

	return pages.Success(map[string]string{
		"Title":       p.Title,
		"Url":         "https://" + params.R.Host + params.R.RequestURI,
		"Description": p.Clickbait,
	})
}
