// pagePage.go serves the page page.

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"zanaduu3/vendor/github.com/gorilla/mux"
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
		// Try to find it in the alias redirects
		var newAlias string
		row := database.NewQuery(`
			SELECT newAlias
			FROM aliasRedirects
			WHERE oldAlias=?`, pageAlias).ToStatement(db).QueryRow()
		exists, err := row.Scan(&newAlias)
		if err != nil {
			return pages.Fail("Couldn't get page info", err)
		}
		if exists {
			return pages.RedirectWith("/p/" + newAlias)
		}

		// Show a page as though the user had searched for this alias
		return pages.RedirectWith("/search/" + pageAlias)
	}

	// Check if we need to redirect.
	var pageType string
	var seeDomainID string
	row := database.NewQuery(`
		SELECT type,seeDomainId
		FROM pageInfos AS pi
		WHERE pageId=?`, pageID).Add(`
			AND`).AddPart(core.PageInfosFilter(u)).ToStatement(db).QueryRow()
	exists, err := row.Scan(&pageType, &seeDomainID)
	if err != nil {
		return pages.Fail("Couldn't get page info", err)
	}

	// Check if a subdomain redirect is necessary.
	if exists && seeDomainID != params.PrivateDomain.ID {
		subdomain := ""
		if core.IsIDValid(seeDomainID) {
			row := database.NewQuery(`
				SELECT alias
				FROM pageInfos AS pi
				WHERE pageId=?`, seeDomainID).Add(`
					AND`).AddPart(core.PageInfosFilter(u)).ToStatement(db).QueryRow()
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

	return pages.Success(dynamicPageTmplData{
		Title:       p.Title,
		URL:         "https://" + params.R.Host + params.R.RequestURI,
		Description: p.Clickbait,
	})
}
