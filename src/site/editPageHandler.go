// editPageHandler.go contains the handler for creating a new page edit.
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// editPageData contains parameters passed in to create a page.
type editPageData struct {
	PageId          string
	PrevEdit        int
	Title           string
	Clickbait       string
	Text            string
	MetaText        string
	IsMinorEditStr  string
	IsAutosave      bool
	IsSnapshot      bool
	SnapshotText    string
	AnchorContext   string
	AnchorText      string
	AnchorOffset    int
	IsEditorComment bool
	// Edit that FE thinks is the current edit
	CurrentEdit int

	// These parameters are only accepted from internal BE calls
	RevertToEdit int  `json:"-"`
	DeleteEdit   bool `json:"-"`
}

var editPageHandler = siteHandler{
	URI:         "/editPage/",
	HandlerFunc: editPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// editPageHandlerFunc handles requests to create a new edit.
func editPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data editPageData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	return editPageInternalHandler(params, &data)
}

func editPageInternalHandler(params *pages.HandlerParams, data *editPageData) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(params.U, false)

	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("No pageId specified", nil)
	}

	// Load the published page.
	var oldPage *core.Page
	oldPage, err := core.LoadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the old page", err)
	} else if oldPage == nil {
		// Likely the page hasn't been published yet, so let's load the unpublished version.
		oldPage, err = core.LoadFullEdit(db, data.PageId, u.Id, &core.LoadEditOptions{LoadNonliveEdit: true})
		if err != nil || oldPage == nil {
			return pages.HandlerErrorFail("Couldn't load the old page2", err)
		}
	}

	// If the client think the current edit is X, but it's actually Y (X!=Y), then
	if oldPage.WasPublished && data.CurrentEdit != oldPage.Edit {
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
		`).QueryRow(data.PageId, u.Id)
	_, err = row.Scan(&myLastAutosaveEdit)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load additional page info", err)
	}

	// Edit number for this new edit will be one higher than the max edit we've had so far...
	isLiveEdit := !data.IsAutosave && !data.IsSnapshot
	newEditNum := oldPage.MaxEditEver + 1
	if myLastAutosaveEdit.Valid {
		// ... unless we can just replace a existing autosave.
		newEditNum = int(myLastAutosaveEdit.Int64)
	}
	if data.RevertToEdit > 0 {
		// ... or unless we are reverting an edit
		newEditNum = data.RevertToEdit
	}

	// Set the see-group
	var seeGroupId string
	if core.IsIdValid(params.PrivateGroupId) {
		seeGroupId = params.PrivateGroupId
	}

	// Error checking.
	if data.IsAutosave && data.IsSnapshot {
		return pages.HandlerBadRequestFail("Can't set autosave and snapshot", nil)
	}
	// Check the page isn't locked by someone else
	if oldPage.LockedUntil > database.Now() && oldPage.LockedBy != u.Id {
		return pages.HandlerBadRequestFail("Can't change locked page", nil)
	}
	// Check the group settings
	if oldPage.SeeGroupId != seeGroupId && newEditNum != 1 {
		return pages.HandlerBadRequestFail("Editing this page in incorrect private group", nil)
	}
	if core.IsIdValid(seeGroupId) && !u.IsMemberOfGroup(seeGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
	}
	if core.IsIdValid(oldPage.SeeGroupId) && !u.IsMemberOfGroup(oldPage.SeeGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
	}
	if core.IsIdValid(oldPage.EditGroupId) && !u.IsMemberOfGroup(oldPage.EditGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to edit this page", nil)
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	if isLiveEdit {
		if len(data.Title) <= 0 && oldPage.Type != core.CommentPageType && oldPage.Type != core.AnswerPageType {
			return pages.HandlerBadRequestFail("Need title", nil)
		}
		if len(data.Text) <= 0 {
			return pages.HandlerBadRequestFail("Need text", nil)
		}
	}
	if !data.IsAutosave {
		if data.AnchorContext == "" && data.AnchorText != "" {
			return pages.HandlerBadRequestFail("Anchor context isn't set", nil)
		}
		if data.AnchorContext != "" && data.AnchorText == "" {
			return pages.HandlerBadRequestFail("Anchor text isn't set", nil)
		}
		if data.AnchorOffset < 0 || data.AnchorOffset > len(data.AnchorContext) {
			return pages.HandlerBadRequestFail("Anchor offset out of bounds", nil)
		}
	}
	// Make sure the user has the right permissions to edit this page
	if oldPage.WasPublished {
		editLevel := core.GetEditLevel(oldPage, u)
		if editLevel != "" && editLevel != "admin" {
			return pages.HandlerBadRequestFail("Not enough karma to edit this page.", nil)
		}
	}
	if isLiveEdit {
		// Process meta text
		_, err := core.ParseMetaText(data.MetaText)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't unmarshal meta-text", err)
		}
	}

	// Load parents for comments
	var commentParentId string
	var commentPrimaryPageId string
	if isLiveEdit && oldPage.Type == core.CommentPageType {
		commentParentId, commentPrimaryPageId, err = core.GetCommentParents(db, data.PageId)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load comment's parents", err)
		}
	}

	primaryPageMap := make(map[string]*core.Page)
	primaryPage := core.AddPageIdToMap(data.PageId, primaryPageMap)
	pageMap := make(map[string]*core.Page)
	if isLiveEdit && !oldPage.WasPublished {
		// Load parents and children.
		err = core.LoadParentIds(db, pageMap, u, &core.LoadParentIdsOptions{ForPages: primaryPageMap})
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load parents", err)
		}
		err = core.LoadChildIds(db, pageMap, u, &core.LoadChildIdsOptions{ForPages: primaryPageMap})
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load children", err)
		}
	}

	// Standardize text
	data.Text = strings.Replace(data.Text, "\r\n", "\n", -1)
	data.Text, err = core.StandardizeLinks(db, data.Text)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't standardize links", err)
	}
	data.MetaText = strings.Replace(data.MetaText, "\r\n", "\n", -1)
	if !data.IsSnapshot {
		data.SnapshotText = ""
	}

	// Compute title
	if oldPage.Type == core.LensPageType {
		if strings.ContainsAny(data.Title, ":") {
			return pages.HandlerBadRequestFail(`Lens title can't include ":" character`, nil)
		}
		// Load parent's title
		parentTitle := ""
		found, err := db.NewStatement(`
			SELECT p.title
			FROM pageInfos AS pi
			JOIN pagePairs AS pp
			ON (pi.pageId=pp.parentId)
			JOIN pages AS p
			ON (pi.pageId=p.pageId)
			WHERE pp.type=? AND pp.childId=? AND p.isLiveEdit
			`).QueryRow(core.ParentPagePairType, data.PageId).Scan(&parentTitle)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load lens parent", err)
		} else if found {
			data.Title = fmt.Sprintf("%s: %s", parentTitle, data.Title)
		}
	}

	isMinorEdit := data.IsMinorEditStr == "on"

	// Check if something is actually different from live edit
	if isLiveEdit && oldPage.WasPublished {
		if data.Title == oldPage.Title &&
			data.Clickbait == oldPage.Clickbait &&
			data.Text == oldPage.Text &&
			data.MetaText == oldPage.MetaText &&
			data.AnchorContext == oldPage.AnchorContext &&
			data.AnchorText == oldPage.AnchorText &&
			data.AnchorOffset == oldPage.AnchorOffset {
			return pages.StatusOK(returnData.ToJson())
		}
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		if isLiveEdit {
			// Handle isLiveEdit and clearing previous isLiveEdit if necessary
			if oldPage.WasPublished {
				statement := tx.NewTxStatement("UPDATE pages SET isLiveEdit=false WHERE pageId=? AND isLiveEdit")
				if _, err = statement.Exec(data.PageId); err != nil {
					return "Couldn't update isLiveEdit for old edits", err
				}
			}
		}

		// Create a new edit.
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = newEditNum
		hashmap["prevEdit"] = data.PrevEdit
		hashmap["creatorId"] = u.Id
		hashmap["title"] = data.Title
		hashmap["clickbait"] = data.Clickbait
		hashmap["text"] = data.Text
		hashmap["metaText"] = data.MetaText
		hashmap["todoCount"] = core.ExtractTodoCount(data.Text)
		hashmap["isLiveEdit"] = isLiveEdit
		hashmap["isMinorEdit"] = isMinorEdit
		hashmap["isAutosave"] = data.IsAutosave
		hashmap["isSnapshot"] = data.IsSnapshot
		hashmap["snapshotText"] = data.SnapshotText
		hashmap["createdAt"] = database.Now()
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		statement := tx.NewInsertTxStatement("pages", hashmap, hashmap.GetKeys()...)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert a new page", err
		}

		// Update summaries
		if isLiveEdit {
			// Delete old page summaries
			statement = database.NewQuery(`
				DELETE FROM pageSummaries WHERE pageId=?`, data.PageId).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return "Couldn't delete existing page summaries", err
			}

			_, summaryValues := core.ExtractSummaries(data.PageId, data.Text)
			statement = tx.NewTxStatement(`
				INSERT INTO pageSummaries (pageId,name,text)
				VALUES ` + database.ArgsPlaceholder(len(summaryValues), 3))
			if _, err := statement.Exec(summaryValues...); err != nil {
				return "Couldn't insert page summaries", err
			}
		}

		// Update pageInfos
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		if !oldPage.WasPublished && isLiveEdit {
			hashmap["createdAt"] = database.Now()
			hashmap["createdBy"] = u.Id
		}
		hashmap["maxEdit"] = oldPage.MaxEditEver
		if oldPage.MaxEditEver < newEditNum {
			hashmap["maxEdit"] = newEditNum
		}
		if isLiveEdit {
			hashmap["currentEdit"] = newEditNum
			hashmap["lockedUntil"] = database.Now()
		} else if data.IsAutosave {
			hashmap["lockedBy"] = u.Id
			hashmap["lockedUntil"] = core.GetPageLockedUntilTime()
		}
		statement = tx.NewInsertTxStatement("pageInfos", hashmap, hashmap.GetKeys()...)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't update pageInfos", err
		}

		// Update change logs
		updateChangeLogs := true
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = newEditNum
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		if data.RevertToEdit != 0 {
			hashmap["type"] = core.RevertEditChangeLog
		} else if data.IsSnapshot {
			hashmap["type"] = core.NewSnapshotChangeLog
		} else if isLiveEdit {
			hashmap["type"] = core.NewEditChangeLog
		} else {
			updateChangeLogs = false
		}
		if updateChangeLogs {
			statement = tx.NewInsertTxStatement("changeLogs", hashmap)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't insert new child change log", err
			}
		}

		// Add subscription.
		if isLiveEdit && !oldPage.WasPublished {
			hashmap = make(map[string]interface{})
			hashmap["userId"] = u.Id
			hashmap["toId"] = data.PageId
			if oldPage.Type == core.CommentPageType && core.IsIdValid(commentParentId) {
				hashmap["toId"] = commentParentId // subscribe to the parent comment
			}
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("subscriptions", hashmap, "userId")
			if _, err = statement.Exec(); err != nil {
				return "Couldn't add a subscription", err
			}
		}

		// Update the links table.
		if isLiveEdit {
			err = core.UpdatePageLinks(tx, data.PageId, data.Text, sessions.GetDomain())
			if err != nil {
				return "Couldn't update links", err
			}
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// === Once the transaction has succeeded, we can't really fail on anything
	// else. So we print out errors, but don't return an error. ===

	if isLiveEdit {
		if data.DeleteEdit {
			// Delete it from the elastic index
			err = elastic.DeletePageFromIndex(c, data.PageId)
			if err != nil {
				c.Errorf("failed to update index: %v", err)
			}
		} else {
			// Update elastic search index.
			doc := &elastic.Document{
				PageId:     data.PageId,
				Type:       oldPage.Type,
				Title:      data.Title,
				Clickbait:  data.Clickbait,
				Text:       data.Text,
				Alias:      oldPage.Alias,
				SeeGroupId: seeGroupId,
				CreatorId:  u.Id,
			}
			err = elastic.AddPageToIndex(c, doc)
			if err != nil {
				c.Errorf("failed to update index: %v", err)
			}
		}

		// Generate "edit" update for users who are subscribed to this page.
		if oldPage.WasPublished && !isMinorEdit {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.GoToPageId = data.PageId
			task.SubscribedToId = data.PageId
			if oldPage.Type != core.CommentPageType {
				task.UpdateType = core.PageEditUpdateType
				task.GroupByPageId = data.PageId
			} else {
				task.UpdateType = core.CommentEditUpdateType
				task.GroupByPageId = commentPrimaryPageId
			}
			if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate updates for users who are subscribed to the author.
		if !oldPage.WasPublished && oldPage.Type != core.CommentPageType && !isMinorEdit {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.UpdateType = core.NewPageByUserUpdateType
			task.GroupByUserId = u.Id
			task.SubscribedToId = u.Id
			task.GoToPageId = data.PageId
			if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Do some stuff for a new parent/child.
		if !oldPage.WasPublished && oldPage.Type != core.CommentPageType {
			// Generate updates for users who are subscribed to the parent pages.
			for _, parentIdStr := range primaryPage.ParentIds {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewChildUpdateType
				task.GroupByPageId = parentIdStr
				task.SubscribedToId = parentIdStr
				task.GoToPageId = data.PageId
				if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}

			// Generate updates for users who are subscribed to the child pages.
			for _, childIdStr := range primaryPage.ChildIds {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewParentUpdateType
				task.GroupByPageId = childIdStr
				task.SubscribedToId = childIdStr
				task.GoToPageId = data.PageId
				if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Do some stuff for a new comment.
		if !oldPage.WasPublished && oldPage.Type == core.CommentPageType {
			// Send updates.
			if !isMinorEdit {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.GroupByPageId = commentPrimaryPageId
				task.GoToPageId = data.PageId
				task.EditorsOnly = data.IsEditorComment
				if core.IsIdValid(commentParentId) {
					// This is a new reply
					task.UpdateType = core.ReplyUpdateType
					task.SubscribedToId = commentParentId
				} else {
					// This is a new top level comment
					task.UpdateType = core.TopLevelCommentUpdateType
					task.SubscribedToId = commentPrimaryPageId
				}
				if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}

			// Generate updates for @mentions
			// Find ids and aliases using [@text] syntax.
			exp := regexp.MustCompile("\\[@([0-9]+)\\]")
			submatches := exp.FindAllStringSubmatch(data.Text, -1)
			for _, submatch := range submatches {
				var task tasks.AtMentionUpdateTask
				task.UserId = u.Id
				task.MentionedUserId = submatch[1]
				task.GroupByPageId = commentPrimaryPageId
				task.GoToPageId = data.PageId
				if err := tasks.Enqueue(c, &task, "atMentionUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Create a task to propagate the domain change to all children
		if !oldPage.WasPublished {
			var task tasks.PropagateDomainTask
			task.PageId = data.PageId
			if err := tasks.Enqueue(c, &task, "propagateDomain"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.StatusOK(returnData.ToJson())
}
