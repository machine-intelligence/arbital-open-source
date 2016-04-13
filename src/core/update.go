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
	NewUsedAsTagUpdateType     = "newUsedAsTag"
	NewRequirementUpdateType   = "newRequirement"
	NewRequiredByUpdateType    = "newRequiredBy"
	NewSubjectUpdateType       = "newSubject"
	NewTeacherUpdateType       = "newTeacher"
	AtMentionUpdateType        = "atMention"
	AddedToGroupUpdateType     = "addedToGroup"
	RemovedFromGroupUpdateType = "removedFromGroup"
	NewMarkUpdateType          = "newMark"
)

// UpdateRow is a row from updates table
type UpdateRow struct {
	Id                 string
	UserId             string
	ByUserId           string
	CreatedAt          string
	Type               string
	GroupByPageId      string
	GroupByUserId      string
	Unseen             bool
	SubscribedToId     string
	GoToPageId         string
	MarkId             int64
	IsGoToPageAlive    bool
	SettingsChangeType string
	OldSettingsValue   string
	NewSettingsValue   string
}

// UpdateGroupKey is what we group updateEntries by
type UpdateGroupKey struct {
	GroupByPageId string `json:"groupByPageId"`
	GroupByUserId string `json:"groupByUserId"`
	// True if this is the first time the user is seeing this update
	Unseen bool `json:"unseen"`
}

// UpdateEntry corresponds to one update entry we'll display.
type UpdateEntry struct {
	UserId             string `json:"userId"`
	ByUserId           string `json:"byUserId"`
	Type               string `json:"type"`
	Repeated           int    `json:"repeated"`
	SubscribedToId     string `json:"subscribedToId"`
	GoToPageId         string `json:"goToPageId"`
	IsGoToPageAlive    bool   `json:"isGoToPageAlive"`
	MarkId             int64  `json:"markId"`
	SettingsChangeType string `json:"settingsChangeType"`
	OldSettingsValue   string `json:"oldSettingsValue"`
	NewSettingsValue   string `json:"newSettingsValue"`
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
func LoadUpdateRows(db *database.DB, userId string, pageMap map[string]*Page, userMap map[string]*User, markMap map[string]*Mark, forEmail bool) ([]*UpdateRow, error) {
	emailFilter := ""
	if forEmail {
		emailFilter = "AND unseen AND NOT emailed"
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
		SELECT updates.id,updates.userId,updates.byUserId,updates.createdAt,updates.type,updates.unseen,
			updates.groupByPageId,updates.groupByUserId,updates.subscribedToId,updates.goToPageId,updates.markId,
			SUM(pages.isLiveEdit) > 0 AS isGoToPageAlive,
			COALESCE(changeLogs.type, ''),
			COALESCE(changeLogs.oldSettingsValue, ''),
			COALESCE(changeLogs.newSettingsValue, '')
		FROM updates
		JOIN pages ON updates.goToPageId = pages.pageId
		LEFT JOIN changeLogs ON updates.changeLogId = changeLogs.id
		WHERE updates.userId=? ` + emailFilter + `
		GROUP BY updates.id
		ORDER BY updates.createdAt DESC
		LIMIT 100`).Query(userId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row UpdateRow
		err := rows.Scan(&row.Id, &row.UserId, &row.ByUserId, &row.CreatedAt, &row.Type,
			&row.Unseen, &row.GroupByPageId, &row.GroupByUserId, &row.SubscribedToId,
			&row.GoToPageId, &row.MarkId, &row.IsGoToPageAlive, &row.SettingsChangeType,
			&row.OldSettingsValue, &row.NewSettingsValue)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		AddPageToMap(row.GoToPageId, pageMap, goToPageLoadOptions)
		AddPageToMap(row.GroupByPageId, pageMap, groupLoadOptions)
		AddPageToMap(row.SubscribedToId, pageMap, groupLoadOptions)

		userMap[row.UserId] = &User{Id: row.UserId}
		if IsIdValid(row.ByUserId) {
			userMap[row.ByUserId] = &User{Id: row.ByUserId}
		}
		if IsIdValid(row.GroupByUserId) {
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
func ConvertUpdateRowsToGroups(rows []*UpdateRow, pageMap map[string]*Page) []*UpdateGroup {
	// Now that we have load last visit time for all pages,
	// go through all the update rows and group them.
	groups := make([]*UpdateGroup, 0)
	groupMap := make(map[UpdateGroupKey]*UpdateGroup)
	for _, row := range rows {
		key := UpdateGroupKey{
			GroupByPageId: row.GroupByPageId,
			GroupByUserId: row.GroupByUserId,
			Unseen:        row.Unseen,
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
				if entry.Type == row.Type && entry.SubscribedToId == row.SubscribedToId &&
					entry.ByUserId == row.ByUserId {
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
				SubscribedToId:     row.SubscribedToId,
				GoToPageId:         row.GoToPageId,
				IsGoToPageAlive:    row.IsGoToPageAlive,
				MarkId:             row.MarkId,
				SettingsChangeType: row.SettingsChangeType,
				OldSettingsValue:   row.OldSettingsValue,
				NewSettingsValue:   row.NewSettingsValue,
				CreatedAt:          row.CreatedAt,
				IsVisited:          pageMap != nil && row.CreatedAt < pageMap[row.GoToPageId].LastVisit,
			}
			group.Updates = append(group.Updates, entry)
		}
	}
	return groups
}

// LoadUpdateEmail loads the text and other data for the update email
func LoadUpdateEmail(db *database.DB, userId string) (resultData *UpdateData, retErr error) {
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

	handlerData := NewHandlerData(u, true)

	// Load updates and populate the maps
	resultData.UpdateRows, err = LoadUpdateRows(db, u.Id, handlerData.PageMap, handlerData.UserMap, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load updates: %v", err)
	}

	// Check to make sure there are enough updates
	resultData.UpdateCount = len(resultData.UpdateRows)
	if resultData.UpdateCount < u.EmailThreshold {
		return nil, nil
	}
	resultData.UpdateGroups = ConvertUpdateRowsToGroups(resultData.UpdateRows, handlerData.PageMap)

	// Load pages.
	err = ExecuteLoadPipeline(db, handlerData)
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
		//"UserFirstName": func() string { return u.Id },
		"GetUserUrl": func(userId string) string {
			return fmt.Sprintf(`%s/user/%s`, sessions.GetDomainForTestEmail(), userId)
		},
		"GetUserName": func(userId string) string {
			return handlerData.UserMap[userId].FullName()
		},
		"GetPageUrl": func(pageId string) string {
			return fmt.Sprintf("%s/p/%s/"+handlerData.PageMap[pageId].Alias, sessions.GetDomainForTestEmail(), pageId)
		},
		"GetPageTitle": func(pageId string) string {
			return handlerData.PageMap[pageId].Title
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
