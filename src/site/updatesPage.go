// updatesPage.go serves the update page.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
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
		"tmpl/updatesPage.tmpl", "tmpl/pageHelpers.tmpl",
		"tmpl/navbar.tmpl", "tmpl/footer.tmpl",
		"tmpl/angular.tmpl.js"),
	newPageOptions{RequireLogin: true})

// updatesRenderer renders the updates page.
func updatesRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	data, err := updatesInternalRenderer(w, r, u)
	if err != nil {
		c.Errorf("%s", err)
		c.Inc("updates_page_served_fail")
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("updates_page_served_success")
	return pages.StatusOK(data)
}

// updatesInternalRenderer renders the updates page.
func updatesInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*updatesTmplData, error) {
	var data updatesTmplData
	data.User = u
	c := sessions.NewContext(r)

	db, err := database.GetDB(c)
	if err != nil {
		return nil, err
	}

	// Load the updates and populate page & user maps
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	updateRows, err := core.LoadUpdateRows(db, data.User.Id, data.PageMap, data.UserMap, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load updates: %v", err)
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, data.User.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}
	c.Debugf("=================== %+v", data.PageMap[4806221069651151873])

	// Load auxillary data.
	err = loadAuxPageData(db, data.User.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading aux data: %v", err)
	}

	// Now that we have load last visit time for all pages,
	// go through all the update rows and group them.
	data.UpdateGroups = core.ConvertUpdateRowsToGroups(updateRows, data.PageMap)

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err = loadGroupNames(db, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Load the names for all users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading user names: %v", err)
	}

	// Load subscriptions to users
	err = loadUserSubscriptions(db, u.Id, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading subscriptions to users: %v", err)
	}

	// Zero out all counts.
	statement := db.NewStatement(`
		UPDATE updates
		SET newCount=0
		WHERE userId=?`)
	if _, err = statement.Exec(data.User.Id); err != nil {
		c.Debugf("Couldn't mark updates seen: %v", err)
	}
	c.Inc("updates_page_served_success")
	return &data, nil
}
