// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/user"
)

var editHandler = siteHandler{
	URI:         "/json/edit/",
	HandlerFunc: editJsonHandler,
	Options: pages.PageOptions{
		RequireLogin:    true,
		LoadUpdateCount: true,
	},
}

// editJsonData contains parameters passed in via the request.
type editJsonData struct {
	PageAlias      string
	SpecificEdit   int
	EditLimit      int
	CreatedAtLimit string
}

// editJsonHandler handles the request.
func editJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data editJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	return editJsonInternalHandler(params, &data)
}

func editJsonInternalHandler(params *pages.HandlerParams, data *editJsonData) *pages.Result {
	db := params.DB
	u := params.U

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		// No alias found. Assume user is trying to create a new page with an alias.
		newPageId, err := user.GetNextAvailableId(db)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't get next available Id", err)
		}
		return pages.RedirectWith(core.GetEditPageUrl(newPageId) + "?alias=" + data.PageAlias)
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
		if core.IsIdValid(p.SeeGroupId) {
			return pages.HandlerBadRequestFail("Trying to edit a private page. Go to the corresponding group", err)
		} else {
			return pages.HandlerBadRequestFail("Trying to edit a public page. Go to arbital.com", err)
		}
	}

	// Remove the primary page from the pageMap and add it to the editMap
	returnData := newHandlerData(false)
	returnData.EditMap[pageId] = p
	delete(returnData.PageMap, pageId)

	// Load parents, tags, requirement, and lens pages (to display in Relationship tab)
	// HARDCODED
	/*
		core.AddPageToMap("8992241719442104138", returnData.PageMap, core.TitlePlusLoadOptions)
		core.AddPageToMap("7648631253816709800", returnData.PageMap, core.TitlePlusLoadOptions)
		core.AddPageToMap("6686682198220623534", returnData.PageMap, core.TitlePlusLoadOptions)
		core.AddPageToMap("1407630090992422901", returnData.PageMap, core.TitlePlusLoadOptions)
	*/
	core.AddPageToMap("3l", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("177", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("1ln", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("179", returnData.PageMap, core.TitlePlusLoadOptions)
	// Load data
	core.AddPageToMap(pageId, returnData.PageMap, core.PrimaryEditLoadOptions)
	core.AddPageIdToMap(p.EditGroupId, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	// We need to copy some data from the loaded live version to the edit
	// NOTE: AAAAARGH! This is such an ugly workaround
	livePage := returnData.PageMap[pageId]
	p.LensIds = livePage.LensIds
	p.ChildIds = livePage.ChildIds
	p.ParentIds = livePage.ParentIds
	p.TaggedAsIds = livePage.TaggedAsIds
	p.RequirementIds = livePage.RequirementIds
	p.SubjectIds = livePage.SubjectIds
	p.ChangeLogs = livePage.ChangeLogs
	livePage.ChangeLogs = []*core.ChangeLog{}

	// Grab the lock to this page, but only if we have the right group permissions
	if !core.IsIdValid(p.SeeGroupId) || u.IsMemberOfGroup(p.SeeGroupId) {
		now := database.Now()
		if p.LockedBy == "" || p.LockedUntil < now {
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
			p.LockedBy = hashmap["lockedBy"].(string)
			p.LockedUntil = hashmap["lockedUntil"].(string)
		}
	}

	return pages.StatusOK(returnData.toJson())
}
