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

// editPageData contains parameters passed in to create a page.
type editPageData struct {
	PageId                   string
	PrevEdit                 int
	Title                    string
	Clickbait                string
	Text                     string
	MetaText                 string
	IsMinorEditStr           string
	IsAutosave               bool
	IsSnapshot               bool
	SnapshotText             string
	AnchorContext            string
	AnchorText               string
	AnchorOffset             int
	IsEditorCommentIntention bool
	// Edit that FE thinks is the current edit
	CurrentEdit int

	// These parameters are only accepted from internal BE calls
	RevertToEdit int `json:"-"`
}

type relatedPageData struct {
	PairType    string
	PageId      string
	PageType    string
	CurrentEdit int
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

	if !core.IsIdValid(data.PageId) {
		return pages.Fail("No pageId specified", nil).Status(http.StatusBadRequest)
	}

	// Load the published page.
	var oldPage *core.Page
	oldPage, err := core.LoadFullEdit(db, data.PageId, u, nil)
	if err != nil {
		return pages.Fail("Couldn't load the old page", err)
	} else if oldPage == nil {
		// Likely the page hasn't been published yet, so let's load the unpublished version.
		oldPage, err = core.LoadFullEdit(db, data.PageId, u, &core.LoadEditOptions{LoadNonliveEdit: true})
		if err != nil || oldPage == nil {
			return pages.Fail("Couldn't load the old page2", err)
		}
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
		`).QueryRow(data.PageId, u.Id)
	_, err = row.Scan(&myLastAutosaveEdit)
	if err != nil {
		return pages.Fail("Couldn't load additional page info", err)
	}

	// Edit number for this new edit will be one higher than the max edit we've had so far...
	isLiveEdit := !data.IsAutosave && !data.IsSnapshot
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
	var seeGroupId string
	if core.IsIdValid(params.PrivateGroupId) {
		seeGroupId = params.PrivateGroupId
	}

	// Error checking.
	// Make sure the user has the right permissions to edit this page
	if !oldPage.Permissions.Edit.Has {
		return pages.Fail("Can't edit: "+oldPage.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}
	if data.IsAutosave && data.IsSnapshot {
		return pages.Fail("Can't set autosave and snapshot", nil).Status(http.StatusBadRequest)
	}
	// Check the group settings
	if oldPage.SeeGroupId != seeGroupId && newEditNum != 1 {
		return pages.Fail("Editing this page in incorrect private group", nil).Status(http.StatusBadRequest)
	}
	if core.IsIdValid(seeGroupId) && !u.IsMemberOfGroup(seeGroupId) {
		return pages.Fail("Don't have group permission to EVEN SEE this page", nil).Status(http.StatusBadRequest)
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	if isLiveEdit {
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
	if isLiveEdit {
		// Process meta text
		_, err := core.ParseMetaText(data.MetaText)
		if err != nil {
			return pages.Fail("Couldn't unmarshal meta-text", err)
		}
	}

	// Load parents for comments
	var commentParentId string
	var commentPrimaryPageId string
	if isLiveEdit && oldPage.Type == core.CommentPageType {
		commentParentId, commentPrimaryPageId, err = core.GetCommentParents(db, data.PageId)
		if err != nil {
			return pages.Fail("Couldn't load comment's parents", err)
		}
	}

	// Load relationships so we can send notifications on a page that had
	// relationships but is being published for the first time.
	// Also send notifications if we undelete a page that has had new relationships created since it was deleted.
	var newParents, newChildren []relatedPageData
	if isLiveEdit && (oldPage.IsDeleted || !oldPage.WasPublished) {
		newParents, newChildren, err = getUnpublishedRelationships(db, u, data.PageId)
		if err != nil {
			return pages.Fail("Couldn't get new parents and children for page", err)
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
	if oldPage.Type == core.LensPageType {
		if strings.ContainsAny(data.Title, ":") {
			return pages.Fail(`Lens title can't include ":" character`, nil).Status(http.StatusBadRequest)
		}
		parentTitle, err := core.GetPrimaryParentTitle(db, u, data.PageId)
		if err != nil {
			return pages.Fail("Couldn't load lens's parent", err)
		}
		data.Title = fmt.Sprintf("%s: %s", parentTitle, data.Title)
	} else if oldPage.Type == core.CommentPageType {
		if len(data.Text) > 30 {
			data.Title = fmt.Sprintf("\"%s...\"", data.Text[:27])
		} else {
			data.Title = fmt.Sprintf("\"%s\"", data.Text)
		}
	}

	isMinorEdit := data.IsMinorEditStr == "on"

	// Check if something is actually different from live edit
	if isLiveEdit && oldPage.WasPublished && !oldPage.IsDeleted {
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
	var editChangeLogId int64
	// Whether we created a changeLog for this edit
	var createEditChangeLog bool
	// The ids of the changeLogs for updated relationships
	newParentChangeLogIdMap := make(map[string]int64)

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		if isLiveEdit {
			// Handle isLiveEdit and clearing previous isLiveEdit if necessary
			if oldPage.WasPublished {
				statement := tx.DB.NewStatement("UPDATE pages SET isLiveEdit=false WHERE pageId=? AND isLiveEdit").WithTx(tx)
				if _, err = statement.Exec(data.PageId); err != nil {
					return sessions.NewError("Couldn't update isLiveEdit for old edits", err)
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
		statement := tx.DB.NewInsertStatement("pages", hashmap, hashmap.GetKeys()...).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't insert a new page", err)
		}

		// Update summaries
		if isLiveEdit {
			// Delete old page summaries
			statement = database.NewQuery(`
				DELETE FROM pageSummaries WHERE pageId=?`, data.PageId).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't delete existing page summaries", err)
			}

			_, summaryValues := core.ExtractSummaries(data.PageId, data.Text)
			statement = tx.DB.NewStatement(`
				INSERT INTO pageSummaries (pageId,name,text)
				VALUES ` + database.ArgsPlaceholder(len(summaryValues), 3)).WithTx(tx)
			if _, err := statement.Exec(summaryValues...); err != nil {
				return sessions.NewError("Couldn't insert page summaries", err)
			}
		}

		if isLiveEdit {
			// set pagePairs.everPublished where the current page is the child (and the parent is already published)
			statement = database.NewQuery(`
				UPDATE pagePairs, pageInfos SET pagePairs.everPublished = 1
				WHERE pagePairs.parentId = pageInfos.pageId
					AND pageInfos.currentEdit > 0 AND NOT pageInfos.isDeleted
					AND pagePairs.childId=?`, data.PageId).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't set everPublished on pagePairs", err)
			}

			// set pagePairs.everPublished where the current page is the parent (and the child is already published)
			statement = database.NewQuery(`
				UPDATE pagePairs, pageInfos SET pagePairs.everPublished = 1
				WHERE pagePairs.childId = pageInfos.pageId
					AND pageInfos.currentEdit > 0 AND NOT pageInfos.isDeleted
					AND pagePairs.parentId=?`, data.PageId).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't set everPublished on pagePairs", err)
			}

			// Now that we're publishing this page, create changeLogs for new relationships
			for _, parent := range newParents {
				_, err := addNewChildToChangelog(tx, u.Id, parent.PairType, oldPage.Type, parent.PageId, parent.CurrentEdit,
					data.PageId, true, false)
				if err != nil {
					return sessions.NewError("Couldn't create changeLog for new child", err)
				}
			}
			for _, child := range newChildren {
				newParentChangeLogId, err := addNewParentToChangelog(tx, u.Id, child.PairType, child.PageType, child.PageId, child.CurrentEdit,
					data.PageId, true, false)
				if err != nil {
					return sessions.NewError("Couldn't create changeLog for new parent", err)
				}
				newParentChangeLogIdMap[child.PageId] = newParentChangeLogId
			}
		}

		// Update pageInfos
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		if isLiveEdit && oldPage.IsDeleted {
			hashmap["isDeleted"] = false
			hashmap["mergedInto"] = ""
		}
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
		statement = tx.DB.NewInsertStatement("pageInfos", hashmap, hashmap.GetKeys()...).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos", err)
		}

		// Update change logs. We only create a changeLog for some types of edits.
		createEditChangeLog = true
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = newEditNum
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		if oldPage.IsDeleted {
			hashmap["type"] = core.UndeletePageChangeLog
		} else if data.RevertToEdit != 0 {
			hashmap["type"] = core.RevertEditChangeLog
		} else if data.IsSnapshot {
			hashmap["type"] = core.NewSnapshotChangeLog
		} else if isLiveEdit {
			hashmap["type"] = core.NewEditChangeLog
		} else {
			createEditChangeLog = false
		}
		if createEditChangeLog {
			statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
			result, err := statement.Exec()
			if err != nil {
				return sessions.NewError("Couldn't insert new child change log", err)
			}
			editChangeLogId, err = result.LastInsertId()
			if err != nil {
				return sessions.NewError("Couldn't get id of changeLog", err)
			}
		}

		// Subscribe this user to the page that they just created.
		if isLiveEdit && !oldPage.WasPublished {
			toId := data.PageId
			if oldPage.Type == core.CommentPageType && core.IsIdValid(commentParentId) {
				toId = commentParentId // subscribe to the parent comment
			}
			err2 := addSubscription(tx, u.Id, toId, true)
			if err2 != nil {
				return err2
			}
		}

		// Update the links table.
		if isLiveEdit {
			err = core.UpdatePageLinks(tx, data.PageId, data.Text, sessions.GetDomain())
			if err != nil {
				return sessions.NewError("Couldn't update links", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// === Once the transaction has succeeded, we can't really fail on anything
	// else. So we print out errors, but don't return an error. ===

	if isLiveEdit {
		// Update elastic
		var task tasks.UpdateElasticPageTask
		task.PageId = data.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		// Generate "edit" update for users who are subscribed to this page.
		if oldPage.WasPublished && !isMinorEdit && createEditChangeLog && oldPage.Type != core.CommentPageType {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.GoToPageId = data.PageId
			task.SubscribedToId = data.PageId
			task.ChangeLogId = editChangeLogId
			if oldPage.IsDeleted {
				task.UpdateType = core.ChangeLogUpdateType
			} else {
				task.UpdateType = core.PageEditUpdateType
			}
			task.GroupByPageId = data.PageId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
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
			if createEditChangeLog {
				task.ChangeLogId = editChangeLogId
			}
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate "new relationship" updates for users who are subscribed to this page.
		if (oldPage.IsDeleted || !oldPage.WasPublished) && oldPage.Type != core.CommentPageType {
			for _, child := range newChildren {
				err := tasks.EnqueueNewParentUpdate(c, u.Id, child.PageId, data.PageId, newParentChangeLogIdMap[child.PageId])
				if err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Do some stuff for a new comment.
		if !oldPage.WasPublished && oldPage.Type == core.CommentPageType {
			// Send "new comment" updates.
			if !isMinorEdit {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.GroupByPageId = commentPrimaryPageId
				task.GoToPageId = data.PageId
				task.ForceMaintainersOnly = oldPage.IsEditorComment
				if createEditChangeLog {
					task.ChangeLogId = editChangeLogId
				}
				if core.IsIdValid(commentParentId) {
					// This is a new reply
					task.UpdateType = core.ReplyUpdateType
					task.SubscribedToId = commentParentId
				} else {
					// This is a new top level comment
					task.UpdateType = core.TopLevelCommentUpdateType
					task.SubscribedToId = commentPrimaryPageId
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
				task.UserId = u.Id
				task.MentionedUserId = submatch[1]
				task.GroupByPageId = commentPrimaryPageId
				task.GoToPageId = data.PageId
				if err := tasks.Enqueue(c, &task, nil); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Create a task to propagate the domain change to all children
		if oldPage.IsDeleted || !oldPage.WasPublished {
			var task tasks.PropagateDomainTask
			task.PageId = data.PageId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.Success(returnData)
}

// Find all the relationships a given page is a part of, where the other page is published (and not deleted), but
// where the relationship has not yet become public (e.g. because the given page was deleted or not-yet-published
// when the relationship was created).
func getUnpublishedRelationships(db *database.DB, u *core.CurrentUser, pageId string) ([]relatedPageData, []relatedPageData, error) {
	parents := make([]relatedPageData, 0)
	children := make([]relatedPageData, 0)

	rows := database.NewQuery(`
		SELECT pairType,otherId,otherIsParent,pi.type AS otherPageType,pi.currentEdit AS otherCurrentEdit
		FROM (
			SELECT parentId AS otherId, type AS pairType, true AS otherIsParent
			FROM pagePairs
			WHERE childId=?`, pageId).Add(`AND NOT everPublished
			UNION
			SELECT childId AS otherId, type AS pairType, false AS otherIsParent
			FROM pagePairs
			WHERE parentId=?`, pageId).Add(`AND NOT everPublished
		) AS others
		JOIN`).AddPart(core.PageInfosTableAll(u)).Add(`AS pi
		ON pi.pageId=otherId
		WHERE (otherId=?`, pageId).Add(`) OR
		(
			pi.currentEdit>0 AND NOT pi.isDeleted
		)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pairType, otherId, otherPageType string
		var otherIsParent bool
		var otherCurrentEdit int
		err := rows.Scan(&pairType, &otherId, &otherIsParent, &otherPageType, &otherCurrentEdit)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}

		otherPageData := relatedPageData{
			PairType:    pairType,
			PageId:      otherId,
			PageType:    otherPageType,
			CurrentEdit: otherCurrentEdit,
		}

		if otherIsParent {
			parents = append(parents, otherPageData)
		} else {
			children = append(children, otherPageData)
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to load parents and children: %v", err)
	}

	return parents, children, nil
}
