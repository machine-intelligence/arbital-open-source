// updates.go contains all the updates stuff
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

const (
	// Various types of updates a user can get.
	TopLevelCommentUpdateType  = "topLevelComment"
	ReplyUpdateType            = "reply"
	PageEditUpdateType         = "pageEdit"
	CommentEditUpdateType      = "commentEdit"
	NewPageByUserUpdateType    = "newPageByUser"
	NewParentUpdateType        = "newParent"
	NewChildUpdateType         = "newChild"
	NewRequirementUpdateType   = "newRequirement"
	NewRequiredByUpdateType    = "newRequiredBy"
	NewTagUpdateType           = "newTag"
	NewTaggedByUpdateType      = "newTaggedBy"
	AtMentionUpdateType        = "atMention"
	AddedToGroupUpdateType     = "addedToGroup"
	RemovedFromGroupUpdateType = "removedFromGroup"
)

// UpdateRow is a row from updates table
type UpdateRow struct {
	Id                 int64
	UserId             int64
	ByUserId           int64
	CreatedAt          string
	Type               string
	GroupByPageId      int64
	GroupByUserId      int64
	NewCount           int
	SubscribedToPageId int64
	SubscribedToUserId int64
	GoToPageId         int64
}

// UpdateGroupKey is what we group updateEntries by
type UpdateGroupKey struct {
	GroupByPageId int64 `json:"groupByPageId,string"`
	GroupByUserId int64 `json:"groupByUserId,string"`
	// True if this is the first time the user is seeing this update
	IsNew bool `json:"isNew"`
}

// UpdateEntry corresponds to one update entry we'll display.
type UpdateEntry struct {
	UserId             int64  `json:"userId,string"`
	ByUserId           int64  `json:"byUserId,string"`
	Type               string `json:"type"`
	Repeated           int    `json:"repeated"`
	SubscribedToPageId int64  `json:"subscribedToPageId,string"`
	SubscribedToUserId int64  `json:"subscribedToUserId,string"`
	GoToPageId         int64  `json:"goToPageId,string"`
	// True if the user has gone to the GoToPage
	IsVisited bool `json:"isVisited"`
}

// UpdateGroup is a collection of updates groupped by the context page.
type UpdateGroup struct {
	Key *UpdateGroupKey `json:"key"`
	// The date of the most recent update
	MostRecentDate string         `json:"mostRecentUpdate"`
	Updates        []*UpdateEntry `json:"updates"`
}

// LoadUpdateRows loads all the updates for the given user, populating the
// given maps.
func LoadUpdateRows(db *database.DB, userId int64, pageMap map[int64]*Page, userMap map[int64]*User, forEmail bool) ([]*UpdateRow, error) {
	emailFilter := ""
	if forEmail {
		emailFilter = "AND newCount>0 AND NOT emailed"
	}

	// Create group loading options
	groupLoadOptions := (&PageLoadOptions{
		OriginalCreatedAt: true,
		LastVisit:         true,
		IsSubscribed:      true,
	}).Add(TitlePlusLoadOptions)
	goToPageLoadOptions := (&PageLoadOptions{
		OriginalCreatedAt: true,
		LastVisit:         true,
	}).Add(TitlePlusLoadOptions)

	updateRows := make([]*UpdateRow, 0, 0)
	rows := db.NewStatement(`
		SELECT id,userId,byUserId,createdAt,type,newCount,groupByPageId,groupByUserId,
			subscribedToUserId,subscribedToPageId,goToPageId
		FROM updates
		WHERE userId=? ` + emailFilter + `
		ORDER BY createdAt DESC
		LIMIT 100`).Query(userId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row UpdateRow
		err := rows.Scan(&row.Id, &row.UserId, &row.ByUserId, &row.CreatedAt, &row.Type,
			&row.NewCount, &row.GroupByPageId, &row.GroupByUserId, &row.SubscribedToUserId,
			&row.SubscribedToPageId, &row.GoToPageId)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		AddPageToMap(row.GoToPageId, pageMap, goToPageLoadOptions)
		AddPageToMap(row.GroupByPageId, pageMap, groupLoadOptions)
		AddPageIdToMap(row.SubscribedToPageId, pageMap)

		userMap[row.UserId] = &User{Id: row.UserId}
		if row.ByUserId > 0 {
			userMap[row.ByUserId] = &User{Id: row.ByUserId}
		}
		if row.GroupByUserId > 0 {
			userMap[row.GroupByUserId] = &User{Id: row.GroupByUserId}
		}
		if row.SubscribedToUserId > 0 {
			userMap[row.SubscribedToUserId] = &User{Id: row.SubscribedToUserId}
		}
		updateRows = append(updateRows, &row)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while loading updates: %v", err)
	}
	return updateRows, nil
}

// ConvertUpdateRowsToGroups converts a list of Rows into a list of Groups
func ConvertUpdateRowsToGroups(rows []*UpdateRow, pageMap map[int64]*Page) []*UpdateGroup {
	// Now that we have load last visit time for all pages,
	// go through all the update rows and group them.
	groups := make([]*UpdateGroup, 0)
	groupMap := make(map[UpdateGroupKey]*UpdateGroup)
	for _, row := range rows {
		key := UpdateGroupKey{
			GroupByPageId: row.GroupByPageId,
			GroupByUserId: row.GroupByUserId,
			IsNew:         row.NewCount > 0,
		}
		// Create/update the group.
		group, ok := groupMap[key]
		if !ok {
			group = &UpdateGroup{
				Key:            &key,
				MostRecentDate: row.CreatedAt,
				Updates:        make([]*UpdateEntry, 0),
			}
			groupMap[key] = group
			groups = append(groups, group)
		} else if group.MostRecentDate > row.CreatedAt {
			group.MostRecentDate = row.CreatedAt
		}

		createNewEntry := true
		if row.Type == PageEditUpdateType || row.Type == CommentEditUpdateType {
			// Check if this kind of update already exists
			for _, entry := range group.Updates {
				if entry.Type == row.Type && entry.SubscribedToPageId == row.SubscribedToPageId {
					createNewEntry = false
					entry.Repeated++
					break
				}
			}
		}
		if _, ok := pageMap[row.GoToPageId]; !ok {
			createNewEntry = false
		}
		if createNewEntry {
			// Add new entry to the group
			entry := &UpdateEntry{
				UserId:             row.UserId,
				ByUserId:           row.ByUserId,
				Type:               row.Type,
				Repeated:           1,
				SubscribedToUserId: row.SubscribedToUserId,
				SubscribedToPageId: row.SubscribedToPageId,
				GoToPageId:         row.GoToPageId,
				IsVisited:          pageMap != nil && row.CreatedAt < pageMap[row.GoToPageId].LastVisit,
			}
			group.Updates = append(group.Updates, entry)
		}
	}
	return groups
}
