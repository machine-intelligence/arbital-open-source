// editPageHandler.go contains the handler for creating a new page edit.

package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

const (
	CommentTitleLength = 50 // number of characters to extract from comment text to make its title
)

// editPageData contains parameters passed in to create a page.
type editPageData struct {
	PageID        string
	PrevEdit      int
	Title         string
	Clickbait     string
	Text          string
	MetaText      string
	IsMinorEdit   bool
	EditSummary   string
	IsAutosave    bool
	IsSnapshot    bool
	SnapshotText  string
	AnchorContext string
	AnchorText    string
	AnchorOffset  int
	// If set, the user wants this edit to be a proposal rather than actual edit,
	// even if they have permissions
	IsProposal bool
	// Edit that FE thinks is the current edit
	CurrentEdit int

	// These parameters are only accepted from internal BE calls
	RevertToEdit int `json:"-"`
}

var editPageHandler = siteHandler{
	URI:         "/editPage/",
	HandlerFunc: editPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// editPageHandlerFunc handles requests to create a new edit.
func editPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data editPageData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	return editPageInternalHandler(params, &data)
}

func editPageInternalHandler(params *pages.HandlerParams, data *editPageData) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	if !core.IsIDValid(data.PageID) {
		return pages.Fail("No pageId specified", nil).Status(http.StatusBadRequest)
	}

	// Load the published page.
	editLoadOptions := &core.LoadEditOptions{
		LoadNonliveEdit: true,
		PreferLiveEdit:  true,
	}
	oldPage, err := core.LoadFullEdit(db, data.PageID, u, editLoadOptions)
	if err != nil {
		return pages.Fail("Couldn't load the old page", err)
	} else if oldPage == nil {
		return pages.Fail("Couldn't find the old page", err)
	}

	// If the client think the current edit is X, but it's actually Y where X!=Y
	// (e.g. if someone else published a new version since we started editing), then
	if oldPage.WasPublished && data.RevertToEdit == 0 && data.CurrentEdit != oldPage.Edit {
		// Notify the client with an error
		returnData.ResultMap["obsoleteEdit"] = oldPage
		// And save a snapshot
		data.IsAutosave = false
		data.IsSnapshot = true
		data.SnapshotText = fmt.Sprintf("Automatically saved snapshot (%s)", database.Now())
	}

	// Load additional info
	var myLastAutosaveEdit sql.NullInt64
	row := db.NewStatement(`
		SELECT max(edit)
		FROM pages
		WHERE pageId=? AND creatorId=? AND isAutosave
		`).QueryRow(data.PageID, u.ID)
	_, err = row.Scan(&myLastAutosaveEdit)
	if err != nil {
		return pages.Fail("Couldn't load additional page info", err)
	}

	// If this edit will be visible to the public
	isPublicEdit := !data.IsAutosave && !data.IsSnapshot
	// If this edit will replace the current edit
	isNewCurrentEdit := isPublicEdit

	// Edit number for this new edit will be one higher than the max edit we've had so far...
	newEditNum := oldPage.MaxEditEver + 1
	if oldPage.IsDeleted {
		newEditNum = data.CurrentEdit
	} else if data.RevertToEdit > 0 {
		// ... unless we are reverting an edit
		newEditNum = data.RevertToEdit
	} else if myLastAutosaveEdit.Valid {
		// ... or unless we can just replace an existing autosave.
		newEditNum = int(myLastAutosaveEdit.Int64)
	}

	// Set the see-group
	var seeGroupID string
	if core.IsIDValid(params.PrivateGroupID) {
		seeGroupID = params.PrivateGroupID
	}

	// Error checking.
	// Make sure the user has the right permissions to edit this page
	if !oldPage.Permissions.ProposeEdit.Has && !oldPage.Permissions.Edit.Has {
		return pages.Fail("Can't edit: "+oldPage.Permissions.ProposeEdit.Reason, nil).Status(http.StatusBadRequest)
	} else if !oldPage.Permissions.Edit.Has || data.IsProposal {
		isNewCurrentEdit = false
	}
	if data.IsAutosave && data.IsSnapshot {
		return pages.Fail("Can't set autosave and snapshot", nil).Status(http.StatusBadRequest)
	}
	// Check the group settings
	if oldPage.SeeGroupID != seeGroupID && newEditNum != 1 {
		return pages.Fail("Editing this page in incorrect private group", nil).Status(http.StatusBadRequest)
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	if isPublicEdit {
		if len(data.Title) <= 0 && oldPage.Type != core.CommentPageType {
			return pages.Fail("Need title", nil).Status(http.StatusBadRequest)
		}
		if len(data.Text) <= 0 && oldPage.Type != core.QuestionPageType {
			return pages.Fail("Need text", nil).Status(http.StatusBadRequest)
		}
	}
	if !data.IsAutosave {
		if data.AnchorContext == "" && data.AnchorText != "" {
			return pages.Fail("Anchor context isn't set", nil).Status(http.StatusBadRequest)
		}
		if data.AnchorContext != "" && data.AnchorText == "" {
			return pages.Fail("Anchor text isn't set", nil).Status(http.StatusBadRequest)
		}
		if data.AnchorOffset < 0 || data.AnchorOffset > len(data.AnchorContext) {
			return pages.Fail("Anchor offset out of bounds", nil).Status(http.StatusBadRequest)
		}
	}

	// Load parents for comments
	var commentParentID string
	var commentPrimaryPageID string
	if isNewCurrentEdit && oldPage.Type == core.CommentPageType {
		commentParentID, commentPrimaryPageID, err = core.GetCommentParents(db, data.PageID)
		if err != nil {
			return pages.Fail("Couldn't load comment's parents", err)
		}
	}

	// Standardize text
	data.Text = strings.Replace(data.Text, "\r\n", "\n", -1)
	data.Text, err = core.StandardizeLinks(db, data.Text)
	if err != nil {
		return pages.Fail("Couldn't standardize links", err)
	}
	data.MetaText = strings.Replace(data.MetaText, "\r\n", "\n", -1)
	if !data.IsSnapshot {
		data.SnapshotText = ""
	}

	// Compute title
	if oldPage.Type == core.CommentPageType {
		if len(data.Text) > CommentTitleLength {
			data.Title = fmt.Sprintf("\"%s...\"", data.Text[:CommentTitleLength-3])
		} else {
			data.Title = fmt.Sprintf("\"%s\"", data.Text)
		}
	}

	// Check if something is actually different from live edit
	// NOTE: we do this as the last step before writing data, just so we can be sure
	// exactly what date we'll be writing
	if isNewCurrentEdit && oldPage.WasPublished && !oldPage.IsDeleted {
		if data.Title == oldPage.Title &&
			data.Clickbait == oldPage.Clickbait &&
			data.Text == oldPage.Text &&
			data.MetaText == oldPage.MetaText &&
			data.AnchorContext == oldPage.AnchorContext &&
			data.AnchorText == oldPage.AnchorText &&
			data.AnchorOffset == oldPage.AnchorOffset {
			return pages.Success(returnData)
		}
	}

	// The id of the changeLog for this edit
	var editChangeLogID int64
	// Whether we created a changeLog for this edit
	var createEditChangeLog bool

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		if oldPage.WasPublished && isNewCurrentEdit {
			// Clear previous isNewCurrentEdit
			statement := tx.DB.NewStatement("UPDATE pages SET isLiveEdit=false WHERE pageId=? AND isLiveEdit").WithTx(tx)
			if _, err = statement.Exec(data.PageID); err != nil {
				return sessions.NewError("Couldn't update isLiveEdit", err)
			}
		}

		// Create a new edit.
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageID
		hashmap["edit"] = newEditNum
		hashmap["prevEdit"] = data.PrevEdit
		hashmap["creatorId"] = u.ID
		hashmap["title"] = data.Title
		hashmap["clickbait"] = data.Clickbait
		hashmap["text"] = data.Text
		hashmap["metaText"] = data.MetaText
		hashmap["todoCount"] = core.ExtractTodoCount(data.Text)
		hashmap["isLiveEdit"] = isNewCurrentEdit
		hashmap["isMinorEdit"] = data.IsMinorEdit
		hashmap["editSummary"] = data.EditSummary
		hashmap["isAutosave"] = data.IsAutosave
		hashmap["isSnapshot"] = data.IsSnapshot
		hashmap["snapshotText"] = data.SnapshotText
		hashmap["createdAt"] = database.Now()
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		statement := tx.DB.NewInsertStatement("pages", hashmap, hashmap.GetKeys()...).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't insert a new page", err)
		}

		// Update summaries
		if isNewCurrentEdit {
			// Delete old summaries
			statement = database.NewQuery(`
				DELETE FROM pageSummaries WHERE pageId=?`, data.PageID).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't delete existing page summaries", err)
			}

			// Insert new summaries
			_, summaryValues := core.ExtractSummaries(data.PageID, data.Text)
			statement = tx.DB.NewStatement(`
				INSERT INTO pageSummaries (pageId,name,text)
				VALUES ` + database.ArgsPlaceholder(len(summaryValues), 3)).WithTx(tx)
			if _, err := statement.Exec(summaryValues...); err != nil {
				return sessions.NewError("Couldn't insert page summaries", err)
			}
		}

		// Update pageInfos
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageID
		if isNewCurrentEdit && oldPage.IsDeleted {
			hashmap["isDeleted"] = false
			hashmap["mergedInto"] = ""
		}
		if !oldPage.WasPublished && isNewCurrentEdit {
			hashmap["createdAt"] = database.Now()
			hashmap["createdBy"] = u.ID
		}
		hashmap["maxEdit"] = oldPage.MaxEditEver
		if oldPage.MaxEditEver < newEditNum {
			hashmap["maxEdit"] = newEditNum
		}
		if isNewCurrentEdit {
			hashmap["currentEdit"] = newEditNum
		}
		if isPublicEdit {
			hashmap["lockedUntil"] = database.Now()
		} else if data.IsAutosave {
			hashmap["lockedBy"] = u.ID
			hashmap["lockedUntil"] = core.GetPageLockedUntilTime()
		}
		statement = tx.DB.NewInsertStatement("pageInfos", hashmap, hashmap.GetKeys()...).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos", err)
		}

		// Update change logs. We only create a changeLog for some types of edits.
		createEditChangeLog = true
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageID
		hashmap["edit"] = newEditNum
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		if oldPage.IsDeleted {
			hashmap["type"] = core.UndeletePageChangeLog
			hashmap["newSettingsValue"] = data.EditSummary
		} else if data.RevertToEdit != 0 {
			hashmap["type"] = core.RevertEditChangeLog
		} else if data.IsSnapshot {
			hashmap["type"] = core.NewSnapshotChangeLog
		} else if isNewCurrentEdit {
			hashmap["type"] = core.NewEditChangeLog
			hashmap["newSettingsValue"] = data.EditSummary
		} else if isPublicEdit {
			hashmap["type"] = core.NewEditProposalChangeLog
			hashmap["newSettingsValue"] = data.EditSummary
		} else {
			createEditChangeLog = false
		}
		if createEditChangeLog {
			statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
			result, err := statement.Exec()
			if err != nil {
				return sessions.NewError("Couldn't insert new child change log", err)
			}
			editChangeLogID, err = result.LastInsertId()
			if err != nil {
				return sessions.NewError("Couldn't get id of changeLog", err)
			}
		}

		// Subscribe this user to the page that they just created.
		if !oldPage.WasPublished && isNewCurrentEdit {
			toID := data.PageID
			if oldPage.Type == core.CommentPageType && core.IsIDValid(commentParentID) {
				toID = commentParentID // subscribe to the parent comment
			}
			err2 := addSubscription(tx, u.ID, toID, true)
			if err2 != nil {
				return err2
			}
		}

		// Update the links table.
		if isNewCurrentEdit {
			err = core.UpdatePageLinks(tx, data.PageID, data.Text, sessions.GetDomain())
			if err != nil {
				return sessions.NewError("Couldn't update links", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// This section is here to help us figure out the cause of http://zanaduu.myjetbrains.com/youtrack/issue/1-1658.
	// It will (if working as intended) send us an email whenever the bug repros.
	{
		var currentEditFound, currentEditIsLive bool
		row := db.NewStatement(`
			SELECT p.isLiveEdit
			FROM pageInfos AS pi
			JOIN pages AS p
			ON pi.pageId=p.pageId AND pi.currentEdit=p.edit
			WHERE pi.pageId=? AND pi.currentEdit>0 AND !pi.isDeleted
			`).QueryRow(data.PageID)
		currentEditFound, err := row.Scan(&currentEditIsLive)
		if err != nil {
			return pages.Fail("Couldn't double-check that the current edit is marked as live", err)
		}
		if currentEditFound && !currentEditIsLive {
			debugText := fmt.Sprintf("New instance of vanishing page bug!!! (http://zanaduu.myjetbrains.com/youtrack/issue/1-1658)"+
				"\n\n\noldPage: %+v\n\n\ndata: %+v", oldPage, data)

			db.C.Debugf(debugText)

			var task tasks.SendFeedbackEmailTask
			task.UserID = "5"
			task.UserEmail = "esrogs+debug@gmail.com"
			task.Text = debugText
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	// === Once the transaction has succeeded, we can't really fail on anything
	// else. So we print out errors, but don't return an error. ===

	if isPublicEdit {
		// Generate "edit" update for users who are subscribed to this page.
		if oldPage.WasPublished && !data.IsMinorEdit && createEditChangeLog && oldPage.Type != core.CommentPageType {
			var task tasks.NewUpdateTask
			task.UserID = u.ID
			task.GoToPageID = data.PageID
			task.SubscribedToID = data.PageID
			task.ChangeLogID = editChangeLogID
			if oldPage.IsDeleted {
				task.UpdateType = core.ChangeLogUpdateType
			} else {
				task.UpdateType = core.PageEditUpdateType
			}
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	if isNewCurrentEdit {
		// Update elastic
		var task tasks.UpdateElasticPageTask
		task.PageID = data.PageID
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		// Generate updates for users who are subscribed to the author.
		if !oldPage.WasPublished && oldPage.Type != core.CommentPageType && !data.IsMinorEdit {
			var task tasks.NewUpdateTask
			task.UserID = u.ID
			task.UpdateType = core.NewPageByUserUpdateType
			task.SubscribedToID = u.ID
			task.GoToPageID = data.PageID
			if createEditChangeLog {
				task.ChangeLogID = editChangeLogID
			}
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Do some stuff for a new comment.
		if !oldPage.WasPublished && oldPage.Type == core.CommentPageType {
			// Send "new comment" updates.
			if !data.IsMinorEdit {
				var task tasks.NewUpdateTask
				task.UserID = u.ID
				task.GoToPageID = data.PageID
				task.ForceMaintainersOnly = oldPage.IsEditorComment
				if createEditChangeLog {
					task.ChangeLogID = editChangeLogID
				}
				if core.IsIDValid(commentParentID) {
					// This is a new reply
					task.UpdateType = core.ReplyUpdateType
					task.SubscribedToID = commentParentID
				} else {
					// This is a new top level comment
					task.UpdateType = core.TopLevelCommentUpdateType
					task.SubscribedToID = commentPrimaryPageID
				}
				if err := tasks.Enqueue(c, &task, nil); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}

			// Generate updates for @mentions
			// Find ids and aliases using [@text] syntax.
			exp := regexp.MustCompile("\\[@([0-9]+)\\]")
			submatches := exp.FindAllStringSubmatch(data.Text, -1)
			for _, submatch := range submatches {
				var task tasks.AtMentionUpdateTask
				task.UserID = u.ID
				task.MentionedUserID = submatch[1]
				task.GoToPageID = data.PageID
				if err := tasks.Enqueue(c, &task, nil); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Create a task to check if any of the relationships need to be published
		// TODO: condition: this page went from unpublished to published or deleted to undeleted
		{
			var task tasks.UpdatePagePairsTask
			task.PageID = data.PageID
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Create a task to propagate the domain change to all children
		if oldPage.IsDeleted || !oldPage.WasPublished {
			var task tasks.PropagateDomainTask
			task.PageID = data.PageID
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.Success(returnData)
}
