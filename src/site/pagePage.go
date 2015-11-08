// pagePage.go serves the page page.
package site

import (
	"fmt"

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
	if pageAlias != fmt.Sprintf("%d", pageId) {
		return pages.RedirectWith(core.GetPageUrl(pageId))
	}

	// Check if we need to redirect.
	var pageType string
	var seeGroupId int64
	row := database.NewQuery(`
		SELECT type,seeGroupId
		FROM pages
		WHERE isCurrentEdit AND pageId=?`, pageId).ToStatement(db).QueryRow()
	exists, err := row.Scan(&pageType, &seeGroupId)
	if err != nil {
		return pages.Fail("Couldn't get page info", err)
	}

	// Redirect certain types of pages to the corresponding primary page
	if pageType == core.LensPageType || pageType == core.CommentPageType {
		var parentId int64
		row := database.NewQuery(`
			SELECT pp.parentId
			FROM pages AS p
			JOIN pagePairs AS pp
			ON p.pageId=pp.parentId
			WHERE p.isCurrentEdit AND pp.childId=?`, pageId).Add(`
				AND p.type!=?`, core.CommentPageType).Add(`
			LIMIT 1`).ToStatement(db).QueryRow()
		exists, err := row.Scan(&parentId)
		if err != nil {
			return pages.Fail("Couldn't get parent", err)
		}
		if exists {
			// Need to redirect
			if pageType == core.LensPageType {
				// Redirect lens pages to the parent page.
				pageUrl := core.GetPageUrl(parentId)
				return pages.RedirectWith(fmt.Sprintf("%s?lens=%d", pageUrl, pageId))
			} else if pageType == core.CommentPageType {
				// Redirect comment pages to the primary page.
				// Note: we are actually redirecting blindly to a parent, which for replies
				// could be the parent comment. For now that's okay, since we just do another
				// redirect then.
				pageUrl := core.GetPageUrl(parentId)
				return pages.RedirectWith(fmt.Sprintf("%s#subpage-%d", pageUrl, pageId))
			}
		}
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

	return pages.StatusOK(nil)
}
