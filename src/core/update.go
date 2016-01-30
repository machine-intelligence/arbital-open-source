// updates.go contains all the updates stuff
package core

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"time"

	"appengine/urlfetch"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/dustin/go-humanize"
)

const (
	// Various types of updates a user can get.
	TopLevelCommentUpdateType  = "topLevelComment"
	ReplyUpdateType            = "reply"
	PageEditUpdateType         = "pageEdit"
	PageInfoEditUpdateType     = "pageInfoEdit"
	CommentEditUpdateType      = "commentEdit"
	NewPageByUserUpdateType    = "newPageByUser"
	NewParentUpdateType        = "newParent"
	NewChildUpdateType         = "newChild"
	NewTagUpdateType           = "newTag"
	NewTaggedByUpdateType      = "newTaggedBy"
	NewRequirementUpdateType   = "newRequirement"
	NewRequiredByUpdateType    = "newRequiredBy"
	NewSubjectUpdateType       = "newSubject"
	NewTeachesUpdateType       = "newTeaches"
	AtMentionUpdateType        = "atMention"
	AddedToGroupUpdateType     = "addedToGroup"
	RemovedFromGroupUpdateType = "removedFromGroup"
)

// UpdateRow is a row from updates table
type UpdateRow struct {
	Id             int64
	UserId         int64
	ByUserId       int64
	CreatedAt      string
	Type           string
	GroupByPageId  int64
	GroupByUserId  int64
	NewCount       int
	SubscribedToId int64
	GoToPageId     int64
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
	UserId         int64  `json:"userId,string"`
	ByUserId       int64  `json:"byUserId,string"`
	Type           string `json:"type"`
	Repeated       int    `json:"repeated"`
	SubscribedToId int64  `json:"subscribedToId,string"`
	GoToPageId     int64  `json:"goToPageId,string"`
	// True if the user has gone to the GoToPage
	IsVisited bool   `json:"isVisited"`
	CreatedAt string `json:"createdAt"`
}

// UpdateGroup is a collection of updates groupped by the context page.
type UpdateGroup struct {
	Key *UpdateGroupKey `json:"key"`
	// The date of the most recent update
	MostRecentDate string         `json:"mostRecentDate"`
	Updates        []*UpdateEntry `json:"updates"`
}

// UpdateData is all the data collected by LoadUpdateEmail()
type UpdateData struct {
	UpdateCount        int
	UpdateRows         []*UpdateRow
	UpdateGroups       []*UpdateGroup
	UpdateEmailAddress string
	UpdateEmailText    string
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
		LastVisit:    true,
		IsSubscribed: true,
	}).Add(TitlePlusLoadOptions)
	goToPageLoadOptions := (&PageLoadOptions{
		LastVisit: true,
	}).Add(TitlePlusLoadOptions)

	updateRows := make([]*UpdateRow, 0, 0)
	rows := db.NewStatement(`
		SELECT id,userId,byUserId,createdAt,type,newCount,groupByPageId,groupByUserId,
			subscribedToId,goToPageId
		FROM updates
		WHERE userId=? ` + emailFilter + `
		ORDER BY createdAt DESC
		LIMIT 100`).Query(userId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row UpdateRow
		err := rows.Scan(&row.Id, &row.UserId, &row.ByUserId, &row.CreatedAt, &row.Type,
			&row.NewCount, &row.GroupByPageId, &row.GroupByUserId, &row.SubscribedToId,
			&row.GoToPageId)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		AddPageToMap(row.GoToPageId, pageMap, goToPageLoadOptions)
		AddPageToMap(row.GroupByPageId, pageMap, groupLoadOptions)
		AddPageToMap(row.SubscribedToId, pageMap, groupLoadOptions)

		userMap[row.UserId] = &User{Id: row.UserId}
		if row.ByUserId > 0 {
			userMap[row.ByUserId] = &User{Id: row.ByUserId}
		}
		if row.GroupByUserId > 0 {
			userMap[row.GroupByUserId] = &User{Id: row.GroupByUserId}
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
		} else if group.MostRecentDate < row.CreatedAt {
			group.MostRecentDate = row.CreatedAt
		}

		createNewEntry := true
		if row.Type == PageEditUpdateType || row.Type == CommentEditUpdateType {
			// Check if this kind of update already exists
			for _, entry := range group.Updates {
				if entry.Type == row.Type && entry.SubscribedToId == row.SubscribedToId {
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
				UserId:         row.UserId,
				ByUserId:       row.ByUserId,
				Type:           row.Type,
				Repeated:       1,
				SubscribedToId: row.SubscribedToId,
				GoToPageId:     row.GoToPageId,
				CreatedAt:      row.CreatedAt,
				IsVisited:      pageMap != nil && row.CreatedAt < pageMap[row.GoToPageId].LastVisit,
			}
			group.Updates = append(group.Updates, entry)
		}
	}
	return groups
}

// LoadUpdateEmail loads the text and other data for the update email
func LoadUpdateEmail(db *database.DB, userId int64) (resultData *UpdateData, retErr error) {
	c := db.C

	resultData = &UpdateData{}

	u := &user.User{}
	row := db.NewStatement(`
		SELECT id,email,emailFrequency,emailThreshold
		FROM users
		WHERE id=?`).QueryRow(userId)
	_, err := row.Scan(&u.Id, &u.Email, &u.EmailFrequency, &u.EmailThreshold)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	}

	// Load the groups the user belongs to.
	if err = LoadUserGroupIds(db, u); err != nil {
		return nil, fmt.Errorf("Couldn't load user groups: %v", err)
	}

	// Load updates and populate the maps
	pageMap := make(map[int64]*Page)
	userMap := make(map[int64]*User)
	masteryMap := make(map[int64]*Mastery)
	resultData.UpdateRows, err = LoadUpdateRows(db, u.Id, pageMap, userMap, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load updates: %v", err)
	}

	// Check to make sure there are enough updates
	resultData.UpdateCount = len(resultData.UpdateRows)
	if resultData.UpdateCount < u.EmailThreshold {
		return nil, nil
	}
	resultData.UpdateGroups = ConvertUpdateRowsToGroups(resultData.UpdateRows, pageMap)

	// Load pages.
	err = ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return nil, fmt.Errorf("Pipeline error: %v", err)
	}

	// Load the template file
	var templateBytes []byte
	if sessions.Live {
		resp, err := urlfetch.Client(c).Get(fmt.Sprintf("%s/static/updatesEmailInlined.html", sessions.GetDomain()))
		if err != nil {
			return nil, fmt.Errorf("Couldn't load the email template form URL: %v", err)
		}
		defer resp.Body.Close()
		templateBytes, err = ioutil.ReadAll(resp.Body)
	} else {
		templateBytes, err = ioutil.ReadFile("../site/static/updatesEmailInlined.html")
	}
	if err != nil {
		return nil, fmt.Errorf("Couldn't load the email template from file: %v", err)
	}

	funcMap := template.FuncMap{
		//"UserFirstName": func() int64 { return u.Id },
		"GetUserUrl": func(userId int64) string {
			return fmt.Sprintf(`%s/user/%d`, sessions.GetDomainForTestEmail(), userId)
		},
		"GetUserName": func(userId int64) string {
			return userMap[userId].FullName()
		},
		"GetPageUrl": func(pageId int64) string {
			return fmt.Sprintf("%s/pages/%d/"+ConvertTitleToUrlFormat(pageMap[pageId].Title), sessions.GetDomainForTestEmail(), pageId)
		},
		"GetPageTitle": func(pageId int64) string {
			return pageMap[pageId].Title
		},
		"RelativeDateTime": func(date string) string {
			t, err := time.Parse(database.TimeLayout, date)
			if err != nil {
				c.Errorf("Couldn't parse %s: %v", date, err)
				return ""
			}
			return humanize.Time(t)
		},
	}

	// Create and execute template
	buffer := &bytes.Buffer{}
	t, err := template.New("email").Funcs(funcMap).Parse(string(templateBytes))
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse template: %v", err)
	}
	err = t.Execute(buffer, resultData)
	if err != nil {
		return nil, fmt.Errorf("Couldn't execute template: %v", err)
	}

	resultData.UpdateEmailAddress = u.Email
	resultData.UpdateEmailText = buffer.String()

	return resultData, nil
}
