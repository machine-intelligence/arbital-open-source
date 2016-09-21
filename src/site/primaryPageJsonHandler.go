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
type primaryPageJSONData struct {
	// Alias of the primary page to load
	PageAlias string
	// Optional lens id that was specified in the url
	LensID string
	// Optional mark id that was specified in the url
	MarkID string
	// Optional path instance id that was specified in the url
	PathInstanceID string
	// Optional id of the page that specifies the path the user is on
	PathPageID string
	// Optional page id for the hub the user is going from
	HubID string
}

var primaryPageHandler = siteHandler{
	URI:         "/json/primaryPage/",
	HandlerFunc: primaryPageJSONHandler,
	Options:     pages.PageOptions{},
}

// primaryPageJsonHandler handles the request.
func primaryPageJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data primaryPageJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageID, ok, err := core.LoadAliasToPageID(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.Fail("Couldn't find page", err)
	}

	// If lens id wasn't explicitly set, check to see if this page is a lens for some other page,
	// and if so, load the primary page.
	if data.LensID == "" {
		page := core.AddPageIDToMap(pageID, returnData.PageMap)
		err := core.LoadLensParentIDs(db, returnData.PageMap, nil)
		if err != nil {
			return pages.Fail("Couldn't load lens parent id", err)
		} else if page.LensParentID != "" {
			c.Infof("Page %s redirected to its primary page %s", pageID, page.LensParentID)
			data.LensID = pageID
			pageID = page.LensParentID
		}
	}

	// Check if page is a user page
	row := database.NewQuery(`
		SELECT id
		FROM users
		WHERE id=?`, pageID).ToStatement(db).QueryRow()
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
			WHERE pi.createdBy=?`, pageID).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupID).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
			ORDER BY pi.createdAt DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIDs(rows, returnData.PageMap, pageOptions)
		if err != nil {
			return pages.Fail("error while loading recently created page ids", err)
		}

		// Load recently created by me comment ids.
		rows = database.NewQuery(`
			SELECT p.pageId
			FROM pages AS p
			JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
			WHERE p.creatorId=?`, pageID).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupID).Add(`
				AND pi.type=?`, core.CommentPageType).Add(`
			ORDER BY pi.createdAt DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["recentlyCreatedCommentIds"], err =
			core.LoadPageIDs(rows, returnData.PageMap, core.TitlePlusLoadOptions)
		if err != nil {
			return pages.Fail("error while loading recently created page ids", err)
		}

		// Load recently edited by me page ids.
		rows = database.NewQuery(`
			SELECT p.pageId
			FROM pages AS p
			JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			ON (p.pageId=pi.pageId)
			WHERE p.creatorId=?`, pageID).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupID).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
			GROUP BY 1
			ORDER BY MAX(p.createdAt) DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIDs(rows, returnData.PageMap, pageOptions)
		if err != nil {
			return pages.Fail("error while loading recently edited page ids", err)
		}

		// Load top pages by me
		rows = database.NewQuery(`
			SELECT pi.pageId
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			JOIN likes AS l2
			ON (pi.likeableId=l2.likeableId)
			WHERE pi.editGroupId=?`, pageID).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupID).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
			GROUP BY 1
			ORDER BY SUM(l2.value) DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
		returnData.ResultMap["topPagesIds"], err = core.LoadPageIDs(rows, returnData.PageMap, core.TitlePlusLoadOptions)
		if err != nil {
			return pages.Fail("error while loading recently edited by me page ids", err)
		}

		returnData.UserMap[pageID] = &core.User{ID: pageID}
	}

	if data.PathInstanceID != "" {
		// Load the path instance
		instance, err := core.LoadPathInstance(db, data.PathInstanceID, u)
		if err != nil {
			return pages.Fail("Couldn't load the path instance: %v", err)
		} else if instance == nil {
			return pages.Fail("Couldn't find the path instance", nil).Status(http.StatusBadRequest)
		}
		returnData.ResultMap["path"] = instance
		core.AddPageIDToMap(instance.GuideID, returnData.PageMap)
	}

	// Load data
	returnData.ResultMap["primaryPageId"] = pageID
	core.AddPageIDToMap("14z", returnData.PageMap)
	core.AddPageIDToMap("4yg", returnData.PageMap) // "Arbital quality"
	core.AddPageIDToMap("3hs", returnData.PageMap) // "Author's guide"
	core.AddPageIDToMap("58l", returnData.PageMap) // "Arbital user groups"
	core.AddPageToMap(pageID, returnData.PageMap, core.PrimaryPageLoadOptions)
	if data.LensID != "" {
		returnData.ResultMap["lensId"] = data.LensID
		core.AddPageToMap(data.LensID, returnData.PageMap, core.LensFullLoadOptions)
	}
	if data.MarkID != "" {
		core.AddMarkToMap(data.MarkID, returnData.MarkMap)
	}
	if data.PathPageID != "" {
		core.AddPageToMap(data.PathPageID, returnData.PageMap, &core.PageLoadOptions{Path: true})
	}
	if data.HubID != "" {
		core.AddPageToMap(data.HubID, returnData.PageMap, &core.PageLoadOptions{HubContent: true})
	}
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
