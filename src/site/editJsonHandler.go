// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"math/rand"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/schema"
)

// editJsonData contains parameters passed in via the request.
type editJsonData struct {
	PageAlias      string
	SpecificEdit   int
	EditLimit      int
	CreatedAtLimit string
}

// editJsonHandler handles the request.
func editJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data editJsonData
	params.R.ParseForm()
	err := schema.NewDecoder().Decode(&data, params.R.Form)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	aliasToIdMap, err := core.LoadAliasToPageIdMap(db, []string{data.PageAlias})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	pageId, ok := aliasToIdMap[data.PageAlias]
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
	p.LoadOptions.Add(core.PrimaryEditLoadOptions)

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)
	masteryMap := make(map[int64]*core.Mastery)
	pageMap[p.PageId] = p
	core.AddPageIdToMap(p.EditGroupId, pageMap)
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	// Grab the lock to this page, but only if we have the right group permissions
	if p.SeeGroupId <= 0 || u.IsMemberOfGroup(p.SeeGroupId) {
		now := database.Now()
		if p.LockedBy <= 0 || p.LockedUntil < now {
			hashmap := make(map[string]interface{})
			hashmap["pageId"] = pageId
			hashmap["createdAt"] = database.Now()
			hashmap["currentEdit"] = -1
			hashmap["lockedBy"] = u.Id
			hashmap["lockedUntil"] = core.GetPageLockedUntilTime()
			statement := db.NewInsertStatement("pageInfos", hashmap, "lockedBy", "lockedUntil")
			if _, err = statement.Exec(); err != nil {
				return pages.Fail("Couldn't add a lock: %v", err)
			}
			p.LockedBy = hashmap["lockedBy"].(int64)
			p.LockedUntil = hashmap["lockedUntil"].(string)
		}
	}

	// Remove the primary page from the pageMap and add it to the editMap
	editMap := make(map[int64]*core.Page)
	editMap[pageId] = p
	delete(pageMap, pageId)

	returnData := createReturnData(pageMap).AddEditMap(editMap).AddUsers(userMap).AddMasteries(masteryMap)
	return pages.StatusOK(returnData)
}
