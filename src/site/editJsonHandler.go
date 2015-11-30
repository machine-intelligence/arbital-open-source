// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"encoding/json"
	"math/rand"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// editJsonData contains parameters passed in via the request.
type editJsonData struct {
	PageAlias      string
	SpecificEdit   int
	EditLimit      int
	CreatedAtLimit string
}

var editHandler = siteHandler{
	URI:         "/json/edit/",
	HandlerFunc: editJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// editJsonHandler handles the request.
func editJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data editJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		// No alias found. Assume user is trying to create a new page with an alias.
		return pages.RedirectWith(core.GetEditPageUrl(rand.Int63()) + "?alias=" + data.PageAlias)
	}

	// Load full edit for one page.
	options := core.LoadEditOptions{
		LoadNonliveEdit:   true,
		LoadSpecificEdit:  data.SpecificEdit,
		LoadEditWithLimit: data.EditLimit,
		CreatedAtLimit:    data.CreatedAtLimit,
	}
	p, err := core.LoadFullEdit(db, pageId, u.Id, &options)
	if err != nil {
		return pages.HandlerErrorFail("Error while loading full edit", err)
	}
	if p == nil {
		return pages.HandlerErrorFail("Exact page not found", err)
	}
	if p.SeeGroupId != params.PrivateGroupId {
		if p.SeeGroupId > 0 {
			return pages.HandlerBadRequestFail("Trying to edit a private page. Go to the corresponding group", err)
		} else {
			return pages.HandlerBadRequestFail("Trying to edit a public page. Go to arbital.com", err)
		}
	}

	// Remove the primary page from the pageMap and add it to the editMap
	returnData := newHandlerData(false)
	returnData.EditMap[pageId] = p
	delete(returnData.PageMap, pageId)

	// Load data
	core.AddPageToMap(pageId, returnData.PageMap, core.PrimaryEditLoadOptions)
	core.AddPageIdToMap(p.EditGroupId, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	// We need to copy some data from the loaded live version to the edit
	livePage := returnData.PageMap[pageId]
	p.ChildIds = livePage.ChildIds
	p.ParentIds = livePage.ParentIds
	p.TaggedAsIds = livePage.TaggedAsIds
	p.RequirementIds = livePage.RequirementIds
	p.ChangeLogs = livePage.ChangeLogs
	livePage.ChangeLogs = []*core.ChangeLog{}

	// Grab the lock to this page, but only if we have the right group permissions
	if p.SeeGroupId <= 0 || u.IsMemberOfGroup(p.SeeGroupId) {
		now := database.Now()
		if p.LockedBy <= 0 || p.LockedUntil < now {
			hashmap := make(map[string]interface{})
			hashmap["pageId"] = pageId
			hashmap["createdAt"] = database.Now()
			hashmap["currentEdit"] = -1
			hashmap["lockedBy"] = u.Id
			hashmap["lockedUntil"] = core.GetPageQuickLockedUntilTime()
			statement := db.NewInsertStatement("pageInfos", hashmap, "lockedBy", "lockedUntil")
			if _, err = statement.Exec(); err != nil {
				return pages.Fail("Couldn't add a lock: %v", err)
			}
			p.LockedBy = hashmap["lockedBy"].(int64)
			p.LockedUntil = hashmap["lockedUntil"].(string)
		}
	}

	return pages.StatusOK(returnData.toJson())
}
