// editPageHandler.go contains the handler for creating a new page edit.
package site

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
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
	PageId         int64 `json:",string"`
	Title          string
	Clickbait      string
	Text           string
	MetaText       string
	IsMinorEditStr string
	IsAutosave     bool
	IsSnapshot     bool
	AnchorContext  string
	AnchorText     string
	AnchorOffset   int

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
	aliasWarningList := make([]string, 0)

	if data.PageId <= 0 {
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

	// Load additional info
	row := db.NewStatement(`
		SELECT max(edit)
		FROM pages
		WHERE pageId=? AND creatorId=? AND isAutosave
		`).QueryRow(data.PageId, u.Id)
	_, err = row.Scan(&oldPage.MyLastAutosaveEdit)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load additional page info", err)
	}

	// Edit number for this new edit will be one higher than the max edit we've had so far...
	isCurrentEdit := !data.IsAutosave && !data.IsSnapshot
	newEditNum := oldPage.MaxEditEver + 1
	if oldPage.MyLastAutosaveEdit.Valid {
		// ... unless we can just replace a existing autosave.
		newEditNum = int(oldPage.MyLastAutosaveEdit.Int64)
	}
	if data.RevertToEdit > 0 {
		// ... or unless we are reverting an edit
		newEditNum = data.RevertToEdit
	}

	// Set the see-group
	var seeGroupId int64
	if params.PrivateGroupId > 0 {
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
	if seeGroupId > 0 && !u.IsMemberOfGroup(seeGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
	}
	if oldPage.SeeGroupId > 0 && !u.IsMemberOfGroup(oldPage.SeeGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
	}
	if oldPage.EditGroupId > 0 && !u.IsMemberOfGroup(oldPage.EditGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to edit this page", nil)
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	if isCurrentEdit {
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
	if isCurrentEdit {
		// Process meta text
		_, err := core.ParseMetaText(data.MetaText)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't unmarshal meta-text", err)
		}
	}

	// Load parents for comments
	var commentParentId int64
	var commentPrimaryPageId int64
	if isCurrentEdit && oldPage.Type == core.CommentPageType {
		rows := db.NewStatement(`
			SELECT pi.pageId,pi.type
			FROM pageInfos AS pi
			JOIN pagePairs AS pp
			ON (pi.pageId=pp.parentId)
			WHERE pp.type=? AND pp.childId=? AND pi.currentEdit>0
			`).Query(core.ParentPagePairType, data.PageId)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId int64
			var pageType string
			err := rows.Scan(&parentId, &pageType)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			if pageType == core.CommentPageType {
				if commentParentId > 0 {
					db.C.Errorf("Can't have more than one comment parent")
				}
				commentParentId = parentId
			} else {
				if commentPrimaryPageId > 0 {
					db.C.Errorf("Can't have more than one non-comment parent for a comment")
				}
				commentPrimaryPageId = parentId
			}
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load comment's parents", err)
		}
		if commentPrimaryPageId <= 0 {
			db.C.Errorf("Comment pages need at least one normal page parent")
		}
	}

	primaryPageMap := make(map[int64]*core.Page)
	primaryPage := core.AddPageIdToMap(data.PageId, primaryPageMap)
	pageMap := make(map[int64]*core.Page)
	if isCurrentEdit && !oldPage.WasPublished {
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

	isMinorEdit := data.IsMinorEditStr == "on"

	// Check if something is actually different from live edit
	if isCurrentEdit && oldPage.WasPublished {
		if data.Title == oldPage.Title &&
			data.Clickbait == oldPage.Clickbait &&
			data.Text == oldPage.Text &&
			data.MetaText == oldPage.MetaText &&
			data.AnchorContext == oldPage.AnchorContext &&
			data.AnchorText == oldPage.AnchorText &&
			data.AnchorOffset == oldPage.AnchorOffset {

			returnData := newHandlerData(false)
			returnData.ResultMap["aliasWarnings"] = aliasWarningList
			return pages.StatusOK(returnData.toJson())
		}
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		if isCurrentEdit {
			// Handle isCurrentEdit and clearing previous isCurrentEdit if necessary
			if oldPage.WasPublished {
				statement := tx.NewTxStatement("UPDATE pages SET isCurrentEdit=false WHERE pageId=? AND isCurrentEdit")
				if _, err = statement.Exec(data.PageId); err != nil {
					return "Couldn't update isCurrentEdit for old edits", err
				}
			}
		}

		// Create a new edit.
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = newEditNum
		hashmap["creatorId"] = u.Id
		hashmap["title"] = data.Title
		hashmap["clickbait"] = data.Clickbait
		hashmap["text"] = data.Text
		hashmap["metaText"] = data.MetaText
		hashmap["todoCount"] = core.ExtractTodoCount(data.Text)
		hashmap["isCurrentEdit"] = isCurrentEdit
		hashmap["isMinorEdit"] = isMinorEdit
		hashmap["isAutosave"] = data.IsAutosave
		hashmap["isSnapshot"] = data.IsSnapshot
		hashmap["createdAt"] = database.Now()
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		statement := tx.NewInsertTxStatement("pages", hashmap, hashmap.GetKeys()...)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert a new page", err
		}

		// Update summaries
		if isCurrentEdit {
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
		if !oldPage.WasPublished && isCurrentEdit {
			hashmap["createdAt"] = database.Now()
			hashmap["createdBy"] = u.Id
		}
		hashmap["maxEdit"] = oldPage.MaxEditEver
		if oldPage.MaxEditEver < newEditNum {
			hashmap["maxEdit"] = newEditNum
		}
		if isCurrentEdit {
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
		} else if isCurrentEdit {
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
		if isCurrentEdit && !oldPage.WasPublished {
			hashmap = make(map[string]interface{})
			hashmap["userId"] = u.Id
			hashmap["toId"] = data.PageId
			if oldPage.Type == core.CommentPageType && commentParentId > 0 {
				hashmap["toId"] = commentParentId // subscribe to the parent comment
			}
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("subscriptions", hashmap, "userId")
			if _, err = statement.Exec(); err != nil {
				return "Couldn't add a subscription", err
			}
		}

		// Update the links table.
		if isCurrentEdit {
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

	if isCurrentEdit {
		if data.DeleteEdit {
			// Delete it from the elastic index
			err = elastic.DeletePageFromIndex(c, data.PageId)
			if err != nil {
				return pages.HandlerErrorFail("failed to update index", err)
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
				parentId, _ := strconv.ParseInt(parentIdStr, 10, 64)
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewChildUpdateType
				task.GroupByPageId = parentId
				task.SubscribedToId = parentId
				task.GoToPageId = data.PageId
				if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}

			// Generate updates for users who are subscribed to the child pages.
			for _, childIdStr := range primaryPage.ChildIds {
				childId, _ := strconv.ParseInt(childIdStr, 10, 64)
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewParentUpdateType
				task.GroupByPageId = childId
				task.SubscribedToId = childId
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
				if commentParentId > 0 {
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
				task.MentionedUserId, _ = strconv.ParseInt(submatch[1], 10, 64)
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

	returnData := newHandlerData(false)
	returnData.ResultMap["aliasWarnings"] = aliasWarningList
	return pages.StatusOK(returnData.toJson())
}
