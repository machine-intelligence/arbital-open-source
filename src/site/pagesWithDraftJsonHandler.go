// Serves JSON for pages with most drafts

package site

import (
	"fmt"
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var pagesWithDraftHandler = siteHandler{
	URI:         "/json/pagesWithDraft/",
	HandlerFunc: pagesWithDraftJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

const PagesWithDraftIds = "pagesWithDraftIds"

func pagesWithDraftJSONHandler(params *pages.HandlerParams) *pages.Result {
	return DashboardListJSONHandler(params, LoadPagesWithDraft, PagesWithDraftIds)
}

func LoadPagesWithDraft(u *core.CurrentUser, privateGroupID string, numToLoad int, db *database.DB, returnData *core.CommonHandlerData,
	pageOptions *core.PageLoadOptions) ([]string, error) {
	// Load pages with unpublished drafts
	pagesWithDraftIDs := make([]string, 0)
	rows := database.NewQuery(`
			SELECT p.pageId,p.title,p.createdAt,pi.currentEdit>0,pi.isDeleted
			FROM pages AS p
			JOIN`).AddPart(core.PageInfosTableAll(u)).Add(`AS pi
			ON (p.pageId = pi.pageId)
			WHERE p.creatorId=?`, u.ID).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
				AND pi.seeGroupId=?`, privateGroupID).Add(`
				AND p.edit>pi.currentEdit AND (p.text!="" OR p.title!="")
			GROUP BY p.pageId
			ORDER BY p.createdAt DESC
			LIMIT ?`, numToLoad).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var title, createdAt string
		var wasPublished bool
		var isDeleted bool
		err := rows.Scan(&pageID, &title, &createdAt, &wasPublished, &isDeleted)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		core.AddPageToMap(pageID, returnData.PageMap, pageOptions)
		pagesWithDraftIDs = append(pagesWithDraftIDs, pageID)
		page := core.AddPageIDToMap(pageID, returnData.EditMap)
		if title == "" {
			title = "*Untitled*"
		}
		page.Title = title
		page.EditCreatedAt = createdAt
		page.WasPublished = wasPublished
		page.IsDeleted = isDeleted
		return nil
	})
	return pagesWithDraftIDs, err
}
