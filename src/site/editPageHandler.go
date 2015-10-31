// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"database/sql"
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
	Type           string
	Title          string
	Clickbait      string
	Text           string
	MetaText       string
	IsMinorEditStr string
	HasVoteStr     string
	VoteType       string
	SeeGroupId     int64 `json:",string"`
	EditGroupId    int64 `json:",string"`
	EditKarmaLock  int
	Alias          string // if empty, leave the current one
	SortChildrenBy string
	IsAutosave     bool
	IsSnapshot     bool
	AnchorContext  string
	AnchorText     string
	AnchorOffset   int

	// These parameters are only accepted from internal BE calls
	RevertToEdit int  `json:"-"`
	DeleteEdit   bool `json:"-"`
}

// editPageHandler handles requests to create a new page.
func editPageHandler(params *pages.HandlerParams) *pages.Result {
	if !params.U.IsLoggedIn {
		return pages.HandlerForbiddenFail("Need to be logged in", nil)
	}

	// Decode data
	decoder := json.NewDecoder(params.R.Body)
	var data editPageData
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

	if data.PageId <= 0 {
		return pages.HandlerBadRequestFail("No pageId specified", nil)
	}

	// Load user groups
	if err := core.LoadUserGroupIds(db, u); err != nil {
		return pages.HandlerForbiddenFail("Couldn't load user groups", err)
	}

	// Load the published page.
	var oldPage *core.Page
	oldPage, err := core.LoadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the old page", err)
	} else if oldPage == nil {
		oldPage = &core.Page{}
	}

	// Load additional info
	row := db.NewStatement(`
		SELECT currentEdit>0,maxEdit,lockedBy,lockedUntil,
			(SELECT max(edit) FROM pages WHERE pageId=? AND creatorId=? AND isAutosave) AS myLastAutosaveEdit,
			(SELECT ifnull(max(voteType),"") FROM pages WHERE pageId=? AND NOT isAutosave AND NOT isSnapshot AND voteType!="") AS lockedVoteType
		FROM pageInfos
		WHERE pageId=?`).QueryRow(data.PageId, u.Id, data.PageId, data.PageId)
	_, err = row.Scan(&oldPage.WasPublished, &oldPage.MaxEditEver,
		&oldPage.LockedBy, &oldPage.LockedUntil, &oldPage.MyLastAutosaveEdit, &oldPage.LockedVoteType)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load additional page info", err)
	}

	// Edit number for this new edit will be one higher than the max edit we've had so far...
	isCurrentEdit := !data.IsAutosave && !data.IsSnapshot
	overwritingEdit := !oldPage.WasPublished && isCurrentEdit
	newEditNum := oldPage.MaxEditEver + 1
	if oldPage.MyLastAutosaveEdit.Valid {
		// ... unless we can just replace a existing autosave.
		overwritingEdit = true
		newEditNum = int(oldPage.MyLastAutosaveEdit.Int64)
	}
	if data.RevertToEdit > 0 {
		// ... or unless we are reverting an edit
		overwritingEdit = true
		newEditNum = data.RevertToEdit
	}

	// Error checking.
	data.Type = strings.ToLower(data.Type)
	if data.IsAutosave && data.IsSnapshot {
		return pages.HandlerBadRequestFail("Can't set autosave and snapshot", nil)
	}
	// Check the page isn't locked by someone else
	if oldPage.LockedUntil > database.Now() && oldPage.LockedBy != u.Id {
		return pages.HandlerBadRequestFail("Can't change locked page", nil)
	}
	// Check the group settings
	if oldPage.SeeGroupId > 0 {
		if !u.IsMemberOfGroup(oldPage.SeeGroupId) {
			return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
		}
	}
	if oldPage.EditGroupId > 0 {
		if !u.IsMemberOfGroup(oldPage.EditGroupId) {
			return pages.HandlerBadRequestFail("Don't have group permission to edit this page", nil)
		}
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	if !data.IsAutosave {
		if len(data.Title) <= 0 && data.Type != core.CommentPageType {
			return pages.HandlerBadRequestFail("Need title", nil)
		}
		if data.Type != core.WikiPageType &&
			data.Type != core.LensPageType &&
			data.Type != core.QuestionPageType &&
			data.Type != core.AnswerPageType &&
			data.Type != core.CommentPageType {
			if data.Type == core.DeletedPageType {
				if !data.DeleteEdit {
					return pages.HandlerBadRequestFail("Can't delete the page like that.", nil)
				}
			} else {
				return pages.HandlerBadRequestFail("Invalid page type.", nil)
			}
		}
		if data.SortChildrenBy != core.LikesChildSortingOption &&
			data.SortChildrenBy != core.RecentFirstChildSortingOption &&
			data.SortChildrenBy != core.OldestFirstChildSortingOption &&
			data.SortChildrenBy != core.AlphabeticalChildSortingOption {
			return pages.HandlerBadRequestFail("Invalid sort children value.", nil)
		}
		if data.VoteType != "" && data.VoteType != core.ProbabilityVoteType && data.VoteType != core.ApprovalVoteType {
			return pages.HandlerBadRequestFail("Invalid vote type value.", nil)
		}
		if data.EditKarmaLock < 0 || data.EditKarmaLock > u.MaxKarmaLock {
			return pages.HandlerBadRequestFail("Karma value out of bounds", nil)
		}
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
		// Set the seeGroupId to primary page's group name
		// Make sure that SeeGroupId is the same as the parents' and childrens'
		seeGroupCount := 0
		var seeGroupId sql.NullInt64
		row = database.NewQuery(`
			SELECT max(p.seeGroupId),count(distinct p.seeGroupId)
			FROM pages AS p
			JOIN pagePairs AS pp
			ON ((p.pageId = pp.parentId AND pp.childId = ?)`, data.PageId).Add(`
				OR (p.pageId = pp.childId AND pp.parentId = ?))`, data.PageId).Add(`
			WHERE p.isCurrentEdit AND
				(pp.type=? OR pp.type=?)`, core.ParentPagePairType, core.TagPagePairType).ToStatement(db).QueryRow()
		_, err = row.Scan(&seeGroupId, &seeGroupCount)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't get primary page's group name", err)
		}
		if seeGroupCount > 1 {
			return pages.HandlerErrorFail("The page, its parents, and its children need to have the same See Group", nil)
		} else if seeGroupCount == 1 {
			data.SeeGroupId = seeGroupId.Int64
		}

		// Process meta text
		_, err := core.ParseMetaText(data.MetaText)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't unmarshal meta-text", err)
		}
	}

	// Data correction. Rewrite the data structure so that we can just use it
	// in a straight-forward way to populate the database.
	// Can't change certain parameters after the page has been published.
	var hasVote bool
	if oldPage.LockedVoteType != "" {
		hasVote = data.HasVoteStr == "on"
		data.VoteType = oldPage.LockedVoteType
	} else {
		hasVote = data.VoteType != ""
	}
	if oldPage.WasPublished && data.RevertToEdit <= 0 && !data.DeleteEdit {
		data.Type = oldPage.Type
	}
	// Enforce SortChildrenBy
	if data.Type == core.CommentPageType {
		data.SortChildrenBy = core.RecentFirstChildSortingOption
	} else if data.Type == core.QuestionPageType {
		data.SortChildrenBy = core.LikesChildSortingOption
	}

	// Make sure alias is valid
	if data.Alias == "" {
		data.Alias = fmt.Sprintf("%d", data.PageId)
	} else if isCurrentEdit && data.Alias != fmt.Sprintf("%d", data.PageId) {
		// Check if the alias matches the strict regexp
		if !core.StrictAliasRegexp.MatchString(data.Alias) {
			return pages.HandlerErrorFail("Invalid alias. Can only contain letters and digits. It cannot be a number.", nil)
		}

		// Prefix alias with the group alias, if appropriate
		if data.SeeGroupId > 0 {
			tempPageMap := map[int64]*core.Page{data.SeeGroupId: core.NewPage(data.SeeGroupId)}
			err = core.LoadPages(db, u, tempPageMap)
			if err != nil {
				return pages.HandlerErrorFail("Couldn't load the see group", err)
			}
			data.Alias = fmt.Sprintf("%s.%s", tempPageMap[data.SeeGroupId].Alias, data.Alias)
		}

		// Check if another page is already using the alias
		var existingPageId int64
		row := db.NewStatement(`
					SELECT pageId
					FROM pages
					WHERE isCurrentEdit AND pageId!=? AND alias=?`).QueryRow(data.PageId, data.Alias)
		exists, err := row.Scan(&existingPageId)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't read from aliases", err)
		} else if exists {
			return pages.HandlerErrorFail(fmt.Sprintf("Alias '%s' is already in use by: %d", data.Alias, existingPageId), nil)
		}
	}

	// Load parents for comments
	var commentParentId int64
	var commentPrimaryPageId int64
	if isCurrentEdit && data.Type == core.CommentPageType {
		rows := db.NewStatement(`
			SELECT p.pageId,p.type
			FROM pages AS p
			JOIN pagePairs AS pp
			ON (p.pageId=pp.parentId AND pp.type=?)
			WHERE pp.childId=?
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
		err = core.LoadParentIds(db, pageMap, &core.LoadParentIdsOptions{ForPages: primaryPageMap})
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load parents", err)
		}
		err = core.LoadChildIds(db, pageMap, &core.LoadChildIdsOptions{ForPages: primaryPageMap})
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
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = newEditNum
		hashmap["creatorId"] = u.Id
		hashmap["title"] = data.Title
		hashmap["clickbait"] = data.Clickbait
		hashmap["text"] = data.Text
		hashmap["metaText"] = data.MetaText
		hashmap["summary"] = core.ExtractSummary(data.Text)
		hashmap["todoCount"] = core.ExtractTodoCount(data.Text)
		hashmap["alias"] = data.Alias
		hashmap["sortChildrenBy"] = data.SortChildrenBy
		hashmap["isCurrentEdit"] = isCurrentEdit
		hashmap["isMinorEdit"] = isMinorEdit
		hashmap["hasVote"] = hasVote
		hashmap["voteType"] = data.VoteType
		hashmap["editKarmaLock"] = data.EditKarmaLock
		hashmap["isAutosave"] = data.IsAutosave
		hashmap["isSnapshot"] = data.IsSnapshot
		hashmap["type"] = data.Type
		hashmap["seeGroupId"] = data.SeeGroupId
		hashmap["editGroupId"] = data.EditGroupId
		hashmap["createdAt"] = database.Now()
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		var statement *database.Stmt
		if overwritingEdit {
			statement = tx.NewReplaceTxStatement("pages", hashmap)
		} else {
			statement = tx.NewInsertTxStatement("pages", hashmap)
		}
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert a new page", err
		}

		// Update pageInfos
		hashmap = make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["createdAt"] = database.Now()
		hashmap["maxEdit"] = oldPage.MaxEditEver
		if oldPage.MaxEditEver < newEditNum {
			hashmap["maxEdit"] = newEditNum
		}
		if isCurrentEdit {
			hashmap["currentEdit"] = newEditNum
			hashmap["lockedUntil"] = database.Now()
			statement = tx.NewInsertTxStatement("pageInfos", hashmap, "maxEdit", "currentEdit", "lockedUntil")
		} else if data.IsAutosave {
			hashmap["lockedBy"] = u.Id
			hashmap["lockedUntil"] = core.GetPageLockedUntilTime()
			statement = tx.NewInsertTxStatement("pageInfos", hashmap, "maxEdit", "lockedBy", "lockedUntil")
		} else {
			statement = tx.NewInsertTxStatement("pageInfos", hashmap, "maxEdit")
		}
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
			hashmap["toPageId"] = data.PageId
			if data.Type == core.CommentPageType && commentParentId > 0 {
				hashmap["toPageId"] = commentParentId // subscribe to the parent comment
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
				Type:       data.Type,
				Title:      data.Title,
				Clickbait:  data.Clickbait,
				Text:       data.Text,
				Alias:      data.Alias,
				SeeGroupId: data.SeeGroupId,
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
			task.SubscribedToPageId = data.PageId
			if data.Type != core.CommentPageType {
				task.UpdateType = core.PageEditUpdateType
				task.GroupByPageId = data.PageId
			} else {
				task.UpdateType = core.CommentEditUpdateType
				task.GroupByPageId = commentPrimaryPageId
			}
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate updates for users who are subscribed to the author.
		if !oldPage.WasPublished && data.Type != core.CommentPageType && !isMinorEdit {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.UpdateType = core.NewPageByUserUpdateType
			task.GroupByUserId = u.Id
			task.SubscribedToUserId = u.Id
			task.GoToPageId = data.PageId
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Do some stuff for a new parent/child.
		if !oldPage.WasPublished && data.Type != core.CommentPageType {
			// Generate updates for users who are subscribed to the parent pages.
			for _, pp := range primaryPage.Parents {
				parentId := pp.ParentId
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewChildUpdateType
				task.GroupByPageId = parentId
				task.SubscribedToPageId = parentId
				task.GoToPageId = data.PageId
				if err := task.IsValid(); err != nil {
					c.Errorf("Invalid task created: %v", err)
				} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}

			// Generate updates for users who are subscribed to the child pages.
			for _, pp := range primaryPage.Children {
				childId := pp.ChildId
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewParentUpdateType
				task.GroupByPageId = childId
				task.SubscribedToPageId = childId
				task.GoToPageId = data.PageId
				if err := task.IsValid(); err != nil {
					c.Errorf("Invalid task created: %v", err)
				} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Do some stuff for a new comment.
		if !oldPage.WasPublished && data.Type == core.CommentPageType {
			// Send updates.
			if !isMinorEdit {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.GroupByPageId = commentPrimaryPageId
				task.GoToPageId = data.PageId
				if commentParentId > 0 {
					// This is a new reply
					task.UpdateType = core.ReplyUpdateType
					task.SubscribedToPageId = commentParentId
				} else {
					// This is a new top level comment
					task.UpdateType = core.TopLevelCommentUpdateType
					task.SubscribedToPageId = commentPrimaryPageId
				}
				if err := task.IsValid(); err != nil {
					c.Errorf("Invalid task created: %v", err)
				} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
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
				if err := task.IsValid(); err != nil {
					c.Errorf("Invalid task created: %v", err)
				} else if err := tasks.Enqueue(c, task, "atMentionUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		// Create a task to propagate the domain change to all children
		if !oldPage.WasPublished {
			var task tasks.PropagateDomainTask
			task.PageId = data.PageId
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			} else if err := tasks.Enqueue(c, task, "propagateDomain"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	returnData := createReturnData(nil).AddResult(newEditNum)
	return pages.StatusOK(returnData)
}
