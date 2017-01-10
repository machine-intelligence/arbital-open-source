// updates.go contains all the updates stuff
package core

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"

	"zanaduu3/vendor/github.com/dustin/go-humanize"
	"zanaduu3/vendor/google.golang.org/appengine/urlfetch"
)

const (
	// Various types of updates a user can get.
	TopLevelCommentUpdateType        = "topLevelComment"
	ReplyUpdateType                  = "reply"
	ChangeLogUpdateType              = "changeLog"
	PageEditUpdateType               = "pageEdit"
	EditProposalAcceptedUpdateType   = "editProposalAccepted"
	NewPageByUserUpdateType          = "newPageByUser"
	PageToDomainSubmissionUpdateType = "pageToDomainSubmission"
	PageToDomainAcceptedUpdateType   = "pageToDomainAccepted"
	AtMentionUpdateType              = "atMention"
	AddedToGroupUpdateType           = "addedToGroup"
	RemovedFromGroupUpdateType       = "removedFromGroup"
	InviteReceivedUpdateType         = "inviteReceived"
	UserTrustUpdateType              = "userTrust"
	NewMarkUpdateType                = "newMark"
	ResolvedThreadUpdateType         = "resolvedThread"
	ResolvedMarkUpdateType           = "resolvedMark"
	AnsweredMarkUpdateType           = "answeredMark"
	QuestionMergedUpdateType         = "questionMerged"
	QuestionMergedReverseUpdateType  = "questionMergedReverse"
)

// UpdateRow is a row from updates table
type UpdateRow struct {
	ID              string
	UserID          string
	ByUserID        string
	CreatedAt       string
	Type            string
	Seen            bool
	SubscribedToID  string
	GoToPageID      string
	MarkID          string
	IsGoToPageAlive bool
	ChangeLog       *ChangeLog
}

// UpdateEntry corresponds to one update entry we'll display.
type UpdateEntry struct {
	ID              string `json:"id"`
	UserID          string `json:"userId"`
	ByUserID        string `json:"byUserId"`
	Type            string `json:"type"`
	SubscribedToID  string `json:"subscribedToId"`
	GoToPageID      string `json:"goToPageId"`
	IsGoToPageAlive bool   `json:"isGoToPageAlive"`
	MarkID          string `json:"markId"`
	// Optional changeLog associated with this update
	ChangeLog *ChangeLog `json:"changeLog"`
	CreatedAt string     `json:"createdAt"`
	Seen      bool       `json:"seen"`
}

// UpdateData is all the data collected by LoadUpdateEmail()
type UpdateData struct {
	UpdateCount        int
	UpdateRows         []*UpdateRow
	UpdateEmailAddress string
	UpdateEmailText    string
}

// LoadUpdateRows loads all the updates for the given user, populating the
// given maps.
func LoadUpdateRows(db *database.DB, u *CurrentUser, resultData *CommonHandlerData, forEmail bool, updateTypes []string, limit int) ([]*UpdateRow, error) {
	emailFilter := database.NewQuery("")
	if forEmail {
		emailFilter = database.NewQuery("AND NOT updates.seen AND NOT updates.emailed")
	}

	updateTypeFilter := database.NewQuery("")
	if len(updateTypes) > 0 {
		updateTypeFilter = database.NewQuery("AND updates.type IN").AddArgsGroupStr(updateTypes)
	}

	if limit <= 0 {
		limit = 100
	}

	// Create group loading options
	groupLoadOptions := (&PageLoadOptions{Subscriptions: true}).Add(TitlePlusIncludeDeletedLoadOptions)
	goToPageLoadOptions := (&PageLoadOptions{
		Parents: true, // to show comment threads properly
	}).Add(TitlePlusIncludeDeletedLoadOptions)
	domainSubmissionLoadOptions := &PageLoadOptions{SubmittedTo: true, DomainsAndPermissions: true}

	updateRows := make([]*UpdateRow, 0)
	changeLogIDs := make([]string, 0)
	changeLogMap := make(map[string]*ChangeLog)

	rows := database.NewQuery(`
		SELECT updates.id,updates.userId,updates.byUserId,updates.createdAt,updates.type,updates.seen,
			updates.subscribedToId,updates.goToPageId,updates.markId,updates.changeLogId,
			COALESCE((
				SELECT !isDeleted
				FROM pageInfos AS pi
				WHERE updates.goToPageId = pageId
					AND`).AddPart(WherePageInfosWithOptions(u, &PageInfosOptions{Deleted: true})).Add(`
			), false) AS isGoToPageAlive
		FROM updates
		WHERE updates.userId=?`, u.ID).AddPart(emailFilter).AddPart(updateTypeFilter).Add(`
			AND NOT updates.dismissed
		HAVING isGoToPageAlive OR updates.type IN`).AddArgsGroupStr(getOkayToShowWhenGoToPageIsDeletedUpdateTypes()).Add(`
		ORDER BY updates.createdAt DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row UpdateRow
		var changeLogID string
		err := rows.Scan(&row.ID, &row.UserID, &row.ByUserID, &row.CreatedAt, &row.Type,
			&row.Seen, &row.SubscribedToID, &row.GoToPageID, &row.MarkID, &changeLogID, &row.IsGoToPageAlive)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}

		AddPageToMap(row.GoToPageID, resultData.PageMap, goToPageLoadOptions)
		AddPageToMap(row.SubscribedToID, resultData.PageMap, groupLoadOptions)
		if row.Type == PageToDomainSubmissionUpdateType {
			// Load domains and permissions, so that if I'm on the page when I see the update, I don't get the orphan page message.
			AddPageToMap(row.GoToPageID, resultData.PageMap, domainSubmissionLoadOptions)
		}

		AddUserIDToMap(row.UserID, resultData.UserMap)
		if IsIDValid(row.ByUserID) {
			AddUserIDToMap(row.ByUserID, resultData.UserMap)
		}
		if row.MarkID == "0" {
			row.MarkID = ""
		} else {
			AddMarkToMap(row.MarkID, resultData.MarkMap)
		}

		// Process the change log
		if changeLogID != "" {
			changeLog, ok := changeLogMap[changeLogID]
			if !ok {
				changeLog = &ChangeLog{ID: changeLogID}
				changeLogMap[changeLogID] = changeLog
				changeLogIDs = append(changeLogIDs, changeLogID)
			}
			row.ChangeLog = changeLog
		}

		updateRows = append(updateRows, &row)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while loading updates: %v", err)
	}

	// Load the changelogs
	if len(changeLogIDs) > 0 {
		changeLogs := make([]*ChangeLog, 0)
		queryPart := database.NewQuery(`WHERE id IN`).AddArgsGroupStr(changeLogIDs)
		err = LoadChangeLogs(db, queryPart, resultData, func(db *database.DB, changeLog *ChangeLog) error {
			*changeLogMap[changeLog.ID] = *changeLog
			changeLogs = append(changeLogs, changeLogMap[changeLog.ID])
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("Couldn't load changlogs: %v", err)
		}
		err = LoadLikesForChangeLogs(db, u, changeLogs)
		if err != nil {
			return nil, fmt.Errorf("error while loading likes for changelogs: %v", err)
		}
	}

	return updateRows, nil
}

// LoadUpdateEmail loads the text and other data for the update email
func LoadUpdateEmail(db *database.DB, userID string) (resultData *UpdateData, retErr error) {
	c := db.C
	resultData = &UpdateData{}

	// TODO: replace this with a helper function (like loadUserFromDb)
	u := NewCurrentUser()
	row := db.NewStatement(`
		SELECT id,email,emailFrequency,emailThreshold
		FROM users
		WHERE id=?`).QueryRow(userID)
	_, err := row.Scan(&u.ID, &u.Email, &u.EmailFrequency, &u.EmailThreshold)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	}

	// Load the groups the user belongs to.
	if err = LoadUserDomainMembership(db, &u.User, nil); err != nil {
		return nil, fmt.Errorf("Couldn't load user domain membership: %v", err)
	}

	handlerData := NewHandlerData(u)

	// Load updates and populate the maps
	updateRows, err := LoadUpdateRows(db, u, handlerData, true, make([]string, 0), -1)
	if err != nil {
		return nil, fmt.Errorf("failed to load updates: %v", err)
	}

	// Filter update rows
	for _, updateRow := range updateRows {
		if updateRow.Type != PageToDomainSubmissionUpdateType {
			resultData.UpdateRows = append(resultData.UpdateRows, updateRow)
		}
	}

	// Check to make sure there are enough updates
	resultData.UpdateCount = len(resultData.UpdateRows)
	u.EmailThreshold = 1 // for now, we are ignoring the threshold
	if resultData.UpdateCount < u.EmailThreshold {
		db.C.Debugf("Not enough updates to send the email: %d < %d", resultData.UpdateCount, u.EmailThreshold)
		return nil, nil
	}

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
		"GetUserUrl": func(userID string) string {
			return fmt.Sprintf(`%s/user/%s`, sessions.GetDomainForTestEmail(), userID)
		},
		"GetUserName": func(userID string) string {
			return handlerData.UserMap[userID].FullName()
		},
		"GetPageUrl": func(pageID string) string {
			return fmt.Sprintf("%s/p/%s/"+handlerData.PageMap[pageID].Alias, sessions.GetDomainForTestEmail(), pageID)
		},
		"GetPageTitle": func(pageID string) string {
			return handlerData.PageMap[pageID].Title
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

func GetAchievementUpdateTypes() []string {
	return []string{
		AddedToGroupUpdateType,
		RemovedFromGroupUpdateType,
		InviteReceivedUpdateType,
		PageToDomainAcceptedUpdateType,
		EditProposalAcceptedUpdateType,
		UserTrustUpdateType,
	}
}

func GetNotificationUpdateTypes() []string {
	return []string{
		TopLevelCommentUpdateType,
		ReplyUpdateType,
		AtMentionUpdateType,
		ResolvedThreadUpdateType,
		ResolvedMarkUpdateType,
		AnsweredMarkUpdateType,
	}
}

func GetMaintenanceUpdateTypes() []string {
	return []string{
		PageEditUpdateType,
		ChangeLogUpdateType,
		PageToDomainSubmissionUpdateType,
		QuestionMergedUpdateType,
		QuestionMergedReverseUpdateType,
		NewMarkUpdateType,
	}
}

func getOkayToShowWhenGoToPageIsDeletedUpdateTypes() []string {
	return []string{
		ChangeLogUpdateType,
		PageEditUpdateType,
	}
}

func MarkUpdatesAsSeen(db *database.DB, userID string, types []string) error {
	statement := database.NewQuery(`
		UPDATE updates
		SET seen=TRUE
		WHERE userId=?`, userID).Add(`
			AND type IN`).AddArgsGroupStr(types).ToStatement(db)
	_, err := statement.Exec()
	return err
}
