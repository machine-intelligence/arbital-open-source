// updatesPage.go serves the update page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	commonPageData
	UpdateGroups []*core.UpdateGroup
}

// updatesPage serves the updates page.
var updatesPage = newPageWithOptions(
	"/updates/",
	updatesRenderer,
	append(baseTmpls,
		"tmpl/updatesPage.tmpl",
		"tmpl/angular.tmpl.js"),
	pages.PageOptions{RequireLogin: true})

// updatesRenderer renders the updates page.
func updatesRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data updatesTmplData
	data.User = u

	// Load the updates and populate page & user maps
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	updateRows, err := core.LoadUpdateRows(db, data.User.Id, data.PageMap, data.UserMap, false)
	if err != nil {
		return pages.Fail("failed to load updates", err)
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, data.User.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(db, data.User.Id, data.PageMap, nil)
	if err != nil {
		return pages.Fail("error while loading aux data", err)
	}

	// Now that we have load last visit time for all pages,
	// go through all the update rows and group them.
	data.UpdateGroups = core.ConvertUpdateRowsToGroups(updateRows, data.PageMap)

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err = loadGroupNames(db, u, data.GroupMap)
	if err != nil {
		return pages.Fail("Couldn't load group names", err)
	}

	// Load the names for all users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, data.UserMap)
	if err != nil {
		return pages.Fail("error while loading user names", err)
	}

	// Load subscriptions to users
	err = loadUserSubscriptions(db, u.Id, data.UserMap)
	if err != nil {
		return pages.Fail("error while loading subscriptions to users", err)
	}

	// Zero out all counts.
	statement := db.NewStatement(`
		UPDATE updates
		SET newCount=0
		WHERE userId=?`)
	if _, err = statement.Exec(data.User.Id); err != nil {
		return pages.Fail("Couldn't mark updates seen", err)
	}
	return pages.StatusOK(&data)
}
