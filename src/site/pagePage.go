// pagePage.go serves the page page.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

// pageTmplData stores the data that we pass to the index.tmpl to render the page
type pageTmplData struct {
	commonPageData
	Page        *core.Page
	LinkedPages []*core.Page
}

// pagePage serves the page page.
var pagePage = newPageWithOptions(
	fmt.Sprintf("/pages/{alias:%s}", core.AliasRegexpStr),
	pageRenderer,
	append(baseTmpls,
		"tmpl/pagePage.tmpl",
		"tmpl/angular.tmpl.js"), pages.PageOptions{})

// pageRenderer renders the page page.
func pageRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data pageTmplData
	data.User = u

	// Get actual page id
	pageAlias := mux.Vars(params.R)["alias"]
	pageId, ok, err := core.LoadAliasToPageId(db, pageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}
	// If the url has an actual alias, then redirect to use page id
	// TODO: do this on the FE
	if pageAlias != fmt.Sprintf("%d", pageId) {
		return pages.RedirectWith(core.GetPageUrl(pageId))
	}

	// Check if we need to redirect.
	var pageType string
	var parentId int64
	row := database.NewQuery(`
		SELECT p.type,pp.parentId
		FROM pages AS p
		JOIN pagePairs AS pp
		ON (p.pageId=pp.childId AND p.pageId=? AND`, pageId).Add(`
			p.isCurrentEdit AND (p.type=? OR p.type=?))`, core.LensPageType, core.CommentPageType).Add(`
		LIMIT 1`).ToStatement(db).QueryRow()
	exists, err := row.Scan(&pageType, &parentId)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't get parents", err)
	}

	if exists {
		// Check if a redirect is necessary.
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

	return pages.StatusOK(&data)
}
