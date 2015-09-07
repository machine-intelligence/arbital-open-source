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
	newPageOptions{RequireLogin: true, LoadUserGroups: true})

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

	// Load the updates and populate page & user maps
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	updateRows, err := core.LoadUpdateRows(c, data.User.Id, data.PageMap, data.UserMap, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load updates: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, data.User.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(c, data.User.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading aux data: %v", err)
	}

	// Now that we have load last visit time for all pages,
	// go through all the update rows and group them.
	data.UpdateGroups = core.ConvertUpdateRowsToGroups(updateRows, data.PageMap)

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Load the names for all users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading user names: %v", err)
	}

	// Load subscriptions to users
	err = loadUserSubscriptions(c, u.Id, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading subscriptions to users: %v", err)
	}

	// Zero out all counts.
	query := fmt.Sprintf(`
		UPDATE updates
		SET newCount=0
		WHERE userId=%d`, data.User.Id)
	database.ExecuteSql(c, query)

	c.Inc("updates_page_served_success")
	return &data, nil
}
