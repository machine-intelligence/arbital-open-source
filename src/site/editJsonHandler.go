// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"fmt"
	"math/rand"
	"strconv"

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

	// Check if the user is trying to create a new page with an alias.
	_, err = strconv.ParseInt(data.PageAlias, 10, 64)
	if err != nil {
		// Okay, it's not an id, but could be an alias.
		row := db.NewStatement(`
			SELECT pageId
			FROM pages
			WHERE isCurrentEdit AND alias=?`).QueryRow(data.PageAlias)
		exists, err := row.Scan(&data.PageAlias)
		if err != nil {
			return pages.Fail("Couldn't convert pageId=>alias", err)
		} else if !exists {
			// No alias found. Assume user is trying to create a new page with an alias.
			return pages.RedirectWith(core.GetEditPageUrl(rand.Int63()) + "?alias=" + data.PageAlias)
		}
	}

	// Get page id.
	pageIdStr := data.PageAlias
	pageId, err := strconv.ParseInt(pageIdStr, 10, 64)
	if err != nil {
		return pages.Fail(fmt.Sprintf("Invalid id passed: %s", pageIdStr), nil)
	}

	userMap := make(map[int64]*core.User)

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
		return pages.HandlerErrorFail("No page with such alias/id", err)
	}

	primaryPageMap := make(map[int64]*core.Page)
	primaryPageMap[pageId] = p

	// Load edit history.
	err = core.LoadEditHistory(db, p, u.Id)
	if err != nil {
		return pages.Fail("Couldn't load editHistory: %v", err)
	}

	pageMap := make(map[int64]*core.Page)
	if p.EditGroupId > 0 {
		if _, ok := pageMap[p.EditGroupId]; !ok {
			core.AddPageIdToMap(p.EditGroupId, pageMap)
		}
	}

	// Load links
	err = core.LoadLinks(db, pageMap, &core.LoadLinksOptions{FromPageMap: primaryPageMap})
	if err != nil {
		return pages.Fail("Couldn't load links", err)
	}

	// Load parents
	err = core.LoadParentsIds(db, pageMap, &core.LoadParentsIdsOptions{ForPages: primaryPageMap})
	if err != nil {
		return pages.Fail("Couldn't load parents: %v", err)
	}

	// Process change logs
	userMap = make(map[int64]*core.User)
	for _, log := range p.ChangeLogs {
		userMap[log.UserId] = &core.User{Id: log.UserId}
		core.AddPageIdToMap(log.AuxPageId, pageMap)
	}

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages: %v", err)
	}
	pageMap[pageId] = p

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

	// Load all the users.
	userMap[u.Id] = &core.User{Id: u.Id}
	userMap[p.LockedBy] = &core.User{Id: p.LockedBy}
	for _, p := range pageMap {
		userMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, userMap)
	if err != nil {
		return pages.HandlerErrorFail("error while loading users", err)
	}

	// Remove the primary page from the pageMap and add it to the editMap
	editMap := make(map[int64]*core.Page)
	editMap[pageId] = p
	delete(pageMap, pageId)

	returnData := createReturnData(pageMap).AddEditMap(editMap).AddUsers(userMap)
	return pages.StatusOK(returnData)
}
