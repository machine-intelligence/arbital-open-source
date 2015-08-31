// updatesPage.go serves the update page.
package site

import (
	"database/sql"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// updateRow is a row from updates table
type updateRow struct {
	UserId             int64
	CreatedAt          string
	Type               string
	GroupByPageId      int64
	GroupByUserId      int64
	NewCount           int
	SubscribedToPageId int64
	SubscribedToUserId int64
	GoToPageId         int64
}

// updatedGroupKey is what we group updateEntries by
type updatedGroupKey struct {
	GroupByPageId int64
	GroupByUserId int64
	// True if this is the first time the user is seeing this update
	IsNew bool
}

// updateEntry corresponds to one update entry we'll display.
type updateEntry struct {
	UserId             int64
	Type               string
	Repeated           int
	SubscribedToPageId int64
	SubscribedToUserId int64
	GoToPageId         int64
	// True if the user has gone to the GoToPage
	IsVisited bool
}

// updatedGroup is a collection of updates groupped by the context page.
type updatedGroup struct {
	Key updatedGroupKey
	// The date of the most recent update
	MostRecentDate string
	Updates        []*updateEntry
}

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	commonPageData
	UpdatedGroups []*updatedGroup
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
	var err error
	var data updatesTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the updates and populate page & user maps
	data.PageMap = make(map[int64]*page)
	data.UserMap = make(map[int64]*dbUser)
	updateRows := make([]*updateRow, 0, 0)
	query := fmt.Sprintf(`
		SELECT userId,createdAt,type,newCount,
			groupByPageId,groupByUserId,
			subscribedToUserId,subscribedToPageId,goToPageId
		FROM updates
		WHERE userId=%d
		ORDER BY createdAt DESC
		LIMIT 100`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var row updateRow
		err := rows.Scan(
			&row.UserId,
			&row.CreatedAt,
			&row.Type,
			&row.NewCount,
			&row.GroupByPageId,
			&row.GroupByUserId,
			&row.SubscribedToUserId,
			&row.SubscribedToPageId,
			&row.GoToPageId)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		data.PageMap[row.GoToPageId] = &page{PageId: row.GoToPageId}
		if row.GroupByPageId > 0 {
			data.PageMap[row.GroupByPageId] = &page{PageId: row.GroupByPageId}
		}
		if row.SubscribedToPageId > 0 {
			data.PageMap[row.SubscribedToPageId] = &page{PageId: row.SubscribedToPageId}
		}

		data.UserMap[row.UserId] = &dbUser{Id: row.UserId}
		if row.GroupByUserId > 0 {
			data.UserMap[row.GroupByUserId] = &dbUser{Id: row.GroupByUserId}
		}
		if row.SubscribedToUserId > 0 {
			data.UserMap[row.SubscribedToUserId] = &dbUser{Id: row.SubscribedToUserId}
		}
		updateRows = append(updateRows, &row)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while loading updates: %v", err)
	}

	// Load pages.
	err = loadPages(c, data.PageMap, data.User.Id, loadPageOptions{})
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
	data.UpdatedGroups = make([]*updatedGroup, 0)
	updatedGroupMap := make(map[updatedGroupKey]*updatedGroup)
	for _, updateRow := range updateRows {
		key := updatedGroupKey{
			GroupByPageId: updateRow.GroupByPageId,
			GroupByUserId: updateRow.GroupByUserId,
			IsNew:         updateRow.NewCount > 0,
		}
		// Create/update the group.
		group, ok := updatedGroupMap[key]
		if !ok {
			group = &updatedGroup{
				Key:            key,
				MostRecentDate: updateRow.CreatedAt,
				Updates:        make([]*updateEntry, 0),
			}
			updatedGroupMap[key] = group
			data.UpdatedGroups = append(data.UpdatedGroups, group)
		} else if group.MostRecentDate > updateRow.CreatedAt {
			group.MostRecentDate = updateRow.CreatedAt
		}

		createNewEntry := true
		if updateRow.Type == pageEditUpdateType || updateRow.Type == commentEditUpdateType {
			// Check if this kind of update already exists
			for _, entry := range group.Updates {
				if entry.Type == updateRow.Type && entry.SubscribedToPageId == updateRow.SubscribedToPageId {
					createNewEntry = false
					entry.Repeated++
					break
				}
			}
		}
		if createNewEntry {
			// Add new entry to the group
			entry := &updateEntry{
				UserId:             updateRow.UserId,
				Type:               updateRow.Type,
				Repeated:           1,
				SubscribedToUserId: updateRow.SubscribedToUserId,
				SubscribedToPageId: updateRow.SubscribedToPageId,
				GoToPageId:         updateRow.GoToPageId,
				IsVisited:          updateRow.CreatedAt < data.PageMap[updateRow.GoToPageId].LastVisit,
			}
			group.Updates = append(group.Updates, entry)
		}
	}

	// Load all the groups.
	data.GroupMap = make(map[int64]*group)
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Load the names for all users.
	data.UserMap[u.Id] = &dbUser{Id: u.Id}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &dbUser{Id: p.CreatorId}
	}
	err = loadUsersInfo(c, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading user names: %v", err)
	}

	// Load subscriptions to users
	err = loadUserSubscriptions(c, u.Id, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading subscriptions to users: %v", err)
	}

	// Zero out all counts.
	query = fmt.Sprintf(`
		UPDATE updates
		SET newCount=0
		WHERE userId=%d`, data.User.Id)
	database.ExecuteSql(c, query)

	c.Inc("updates_page_served_success")
	return &data, nil
}
