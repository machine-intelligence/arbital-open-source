// primaryPageJsonHandler.go contains the handler for returning JSON with data
// to display a primary page.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// primaryPageJsonData contains parameters passed in via the request.
type primaryPageJsonData struct {
	// Alias of the primary page to load
	PageAlias string
	// Optional lens id that was specified in the url
	LensId string
	// Optional mark id that was specified in the url
	MarkId string
	// Optional path instance id that was specified in the url
	PathInstanceId string
}

var primaryPageHandler = siteHandler{
	URI:         "/json/primaryPage/",
	HandlerFunc: primaryPageJsonHandler,
	Options:     pages.PageOptions{},
}

// primaryPageJsonHandler handles the request.
func primaryPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data primaryPageJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.Fail("Couldn't find page", err)
	}

	// If lens id wasn't explicitly set, check to see if this page is a lens for some other page,
	// and if so, load the primary page.
	if data.LensId == "" {
		page := core.AddPageIdToMap(pageId, returnData.PageMap)
		err := core.LoadLensParentIds(db, returnData.PageMap, nil)
		if err != nil {
			return pages.Fail("Couldn't load lens parent id", err)
		} else if page.LensParentId != "" {
			c.Infof("Page %s redirected to its primary page %s", pageId, page.LensParentId)
			data.LensId = pageId
			pageId = page.LensParentId
		}
	}

	// Check if page is a user page
	row := database.NewQuery(`
		SELECT id
		FROM users
		WHERE id=?`, pageId).ToStatement(db).QueryRow()
	var id string
	exists, err := row.Scan(&id)
	if err != nil {
		fmt.Errorf("failed to scan for a member: %v", err)
	}
	// If page is a user page, add some values to returnData
	if exists {
		c.Infof("Page is a user page, id: %v", id)

		// Options to load the pages with
		pageOptions := (&core.PageLoadOptions{
			RedLinkCount: true,
		}).Add(core.TitlePlusLoadOptions)

		// Load recently created by me page ids.
		rows := database.NewQuery(`
			SELECT pi.pageId
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			WHERE pi.createdBy=?`, pageId).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
			ORDER BY pi.createdAt DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
		if err != nil {
			return pages.Fail("error while loading recently created page ids", err)
		}

		// Load recently created by me comment ids.
		rows = database.NewQuery(`
			SELECT p.pageId
			FROM pages AS p
			JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
			WHERE p.creatorId=?`, pageId).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
				AND pi.type=?`, core.CommentPageType).Add(`
			ORDER BY pi.createdAt DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["recentlyCreatedCommentIds"], err =
			core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
		if err != nil {
			return pages.Fail("error while loading recently created page ids", err)
		}

		// Load recently edited by me page ids.
		rows = database.NewQuery(`
			SELECT p.pageId
			FROM pages AS p
			JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			ON (p.pageId=pi.pageId)
			WHERE p.creatorId=?`, pageId).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
			GROUP BY 1
			ORDER BY MAX(p.createdAt) DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
		if err != nil {
			return pages.Fail("error while loading recently edited page ids", err)
		}

		// Load top pages by me
		rows = database.NewQuery(`
			SELECT pi.pageId
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			JOIN likes AS l2
			ON (pi.likeableId=l2.likeableId)
			WHERE pi.editGroupId=?`, pageId).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
			GROUP BY 1
			ORDER BY SUM(l2.value) DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["topPagesIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
		if err != nil {
			return pages.Fail("error while loading recently edited by me page ids", err)
		}

		returnData.UserMap[pageId] = &core.User{Id: pageId}
	}

	// Load data
	returnData.ResultMap["primaryPageId"] = pageId
	core.AddPageIdToMap("14z", returnData.PageMap)
	core.AddPageIdToMap("4yg", returnData.PageMap) // "Arbital quality"
	core.AddPageIdToMap("3hs", returnData.PageMap) // "Author's guide"
	core.AddPageIdToMap("58l", returnData.PageMap) // "Arbital user groups"
	core.AddPageToMap(pageId, returnData.PageMap, core.PrimaryPageLoadOptions)
	if data.LensId != "" {
		returnData.ResultMap["lensId"] = data.LensId
		core.AddPageToMap(data.LensId, returnData.PageMap, core.LensFullLoadOptions)
	}
	if data.MarkId != "" {
		core.AddMarkToMap(data.MarkId, returnData.MarkMap)
	}
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	if data.PathInstanceId != "" {
		// Load the path instance
		instance, err := core.LoadPathInstance(db, data.PathInstanceId, u)
		if err != nil {
			return pages.Fail("Couldn't load the path instance: %v", err)
		} else if instance == nil {
			return pages.Fail("Couldn't find the path instance", nil).Status(http.StatusBadRequest)
		}
		returnData.ResultMap["path"] = instance
	}

	return pages.Success(returnData)
}
