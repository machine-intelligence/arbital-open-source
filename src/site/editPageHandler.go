// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"

	"gopkg.in/yaml.v2"
)

// editPageData contains parameters passed in to create a page.
type editPageData struct {
	PageId         int64 `json:",string"`
	PrevEdit       int   `json:",string"`
	Type           string
	Title          string
	Clickbait      string
	Text           string
	MetaText       string
	IsMinorEditStr string
	HasVoteStr     string
	VoteType       string
	GroupId        int64 `json:",string"`
	KarmaLock      int
	ParentIds      string
	Alias          string // if empty, leave the current one
	SortChildrenBy string
	IsAutosave     bool
	IsSnapshot     bool
	AnchorContext  string
	AnchorText     string
	AnchorOffset   int
}

// editPageHandler handles requests to create a new page.
func editPageHandler(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Need to be logged in", nil)
	}

	// Decode data
	decoder := json.NewDecoder(params.R.Body)
	var data editPageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.PageId <= 0 {
		return pages.HandlerBadRequestFail("No pageId specified", nil)
	}
	parentIds := make([]int64, 0)
	parentIdArgs := make([]interface{}, 0)
	if data.ParentIds != "" {
		parentStrIds := strings.Split(data.ParentIds, ",")
		for _, parentStrId := range parentStrIds {
			parentId, err := strconv.ParseInt(parentStrId, 10, 64)
			if err != nil {
				return pages.HandlerBadRequestFail(fmt.Sprintf("Invalid parent id: %s", parentStrId), nil)
			}
			parentIds = append(parentIds, parentId)
			parentIdArgs = append(parentIdArgs, parentId)
		}
	}

	// Load user groups
	if err = loadUserGroups(db, u); err != nil {
		return pages.HandlerForbiddenFail("Couldn't load user groups", err)
	}

	// Load the published page.
	var oldPage *core.Page
	oldPage, err = loadFullEdit(db, data.PageId, u.Id, &loadEditOptions{})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the old page", err)
	} else if oldPage == nil {
		oldPage = &core.Page{}
	}
	oldPage.ProcessParents(c, nil)

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

	// Error checking.
	data.Type = strings.ToLower(data.Type)
	if data.IsAutosave && data.IsSnapshot {
		return pages.HandlerBadRequestFail("Can't set autosave and snapshot", nil)
	}
	// Check that we have the lock.
	if oldPage.LockedUntil > database.Now() && oldPage.LockedBy != u.Id {
		return pages.HandlerBadRequestFail("Can't change locked page", nil)
	}
	// Check the group settings
	if oldPage.GroupId > 0 {
		if !u.IsMemberOfGroup(oldPage.GroupId) {
			return pages.HandlerBadRequestFail("Don't have group permissions to edit this page", nil)
		}
	}
	// Check PrevEdit number.
	if data.PrevEdit < 0 {
		return pages.HandlerBadRequestFail("PrevEdit number is not valid", nil)
	}
	// TODO: check that this user has access to that edit
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
			return pages.HandlerBadRequestFail("Invalid page type.", nil)
		}
		if data.SortChildrenBy != core.LikesChildSortingOption &&
			data.SortChildrenBy != core.ChronologicalChildSortingOption &&
			data.SortChildrenBy != core.AlphabeticalChildSortingOption {
			return pages.HandlerBadRequestFail("Invalid sort children value.", nil)
		}
		if data.VoteType != "" && data.VoteType != core.ProbabilityVoteType && data.VoteType != core.ApprovalVoteType {
			return pages.HandlerBadRequestFail("Invalid vote type value.", nil)
		}
		if data.KarmaLock < 0 || data.KarmaLock > getMaxKarmaLock(u.Karma) {
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
		for _, parentId := range parentIds {
			if parentId == data.PageId {
				return pages.HandlerBadRequestFail("Can't set a page as its own parent", nil)
			}
		}
	}
	if oldPage.WasPublished {
		editLevel := getEditLevel(oldPage, u)
		if editLevel != "" && editLevel != "admin" {
			if editLevel == core.CommentPageType {
				return pages.HandlerBadRequestFail("Can't edit a comment page you didn't create.", nil)
			}
			return pages.HandlerBadRequestFail("Not enough karma to edit this page.", nil)
		}
	}

	// Check parents for errors
	var commentParentId int64
	var commentPrimaryPageId int64
	if isCurrentEdit {
		if len(parentIds) > 0 {
			// Check that the user has group access to all the pages they are linking to.
			count := 0
			typeConstraint := "TRUE"
			if data.Type == core.AnswerPageType {
				typeConstraint = `type="question"`
			}
			row := database.NewQuery(`
				SELECT COUNT(DISTINCT pageId)
				FROM pages
				WHERE pageId IN`).AddArgsGroup(parentIdArgs).Add(`
					AND (isCurrentEdit OR creatorId=?)`, u.Id).Add(`AND
					`+typeConstraint+` AND
					(groupId="" OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=?))
				`, u.Id).ToStatement(db).QueryRow()
			_, err = row.Scan(&count)
			if err != nil {
				return pages.HandlerErrorFail("Couldn't check parents", err)
			}
			if count != len(parentIds) {
				if data.Type == core.AnswerPageType {
					return pages.HandlerBadRequestFail("Some of the parents are invalid. Perhaps because one of them is not a question.", nil)
				}
				return pages.HandlerBadRequestFail("Some of the parents are invalid. Perhaps one of them doesn't exist or is owned by a group you are a not a part of?", nil)
			}

			// Compute parent comment and primary page for the comment.
			if data.Type == core.CommentPageType {
				rows := db.NewStatement(`
					SELECT pageId,type
					FROM pages
					WHERE pageId IN ` + database.InArgsPlaceholder(len(parentIdArgs)) + ` AND isCurrentEdit
					`).Query(parentIdArgs...)
				err := rows.Process(func(db *database.DB, rows *database.Rows) error {
					var pageId int64
					var pageType string
					err := rows.Scan(&pageId, &pageType)
					if err != nil {
						return fmt.Errorf("failed to scan: %v", err)
					}
					if pageType == core.CommentPageType {
						if commentParentId > 0 {
							return fmt.Errorf("Can't have more than one comment parent")
						}
						commentParentId = pageId
					} else {
						if commentPrimaryPageId > 0 {
							return fmt.Errorf("Can't have more than one non-comment parent for a comment")
						}
						commentPrimaryPageId = pageId
					}
					return nil
				})
				if err != nil {
					return pages.HandlerBadRequestFail("Couldn't load comment's parents", err)
				}
				if commentPrimaryPageId <= 0 {
					return pages.HandlerBadRequestFail("Comment pages need at least one normal page parent", nil)
				}
			}
		}

		if len(parentIds) <= 0 && (data.Type == core.CommentPageType || data.Type == core.AnswerPageType) {
			return pages.HandlerBadRequestFail(fmt.Sprintf("%s pages need to have a parent", data.Type), nil)
		}

		// Lens pages need to have exactly one parent.
		if data.Type == core.LensPageType && len(parentIds) != 1 {
			return pages.HandlerBadRequestFail("Lens pages need to have exactly one parent", nil)
		}

		// We can only change the group from to a more relaxed group:
		// personal group -> any group -> no group
		if oldPage.WasPublished && data.GroupId != oldPage.GroupId {
			if oldPage.GroupId != u.Id && data.GroupId != 0 {
				return pages.HandlerBadRequestFail("Can't change group to a more restrictive one", nil)
			}
		}

		// Process meta text
		var metaData core.PageMetaData
		err = yaml.Unmarshal([]byte(data.MetaText), &metaData)
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
	if oldPage.WasPublished {
		data.Type = oldPage.Type
	} else if data.Type == core.CommentPageType {
		// Set the groupId to primary page's group name
		row := db.NewStatement(`
			SELECT max(groupId)
			FROM pages
			WHERE pageId IN ` + database.InArgsPlaceholder(len(parentIdArgs)) + `
				AND type!="comment" AND isCurrentEdit`).QueryRow(parentIdArgs...)
		_, err := row.Scan(&data.GroupId)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't get primary page's group name", err)
		}
	}
	// Enforce SortChildrenBy
	if data.Type == core.CommentPageType {
		data.SortChildrenBy = core.ChronologicalChildSortingOption
	} else if data.Type == core.QuestionPageType {
		data.SortChildrenBy = core.LikesChildSortingOption
	}

	// Make sure alias is valid
	if data.Alias == "" {
		data.Alias = fmt.Sprintf("%d", data.PageId)
	} else if data.Alias != fmt.Sprintf("%d", data.PageId) {
		// Check if the alias matches the strict regexp
		if !core.StrictAliasRegexp.MatchString(data.Alias) {
			return pages.HandlerErrorFail("Invalid alias. Can only contain letters and digits. It cannot be a number.", nil)
		}

		// Prefix alias with the group alias, if appropriate
		if data.GroupId > 0 {
			groupMap := map[int64]*core.Group{data.GroupId: &core.Group{Id: data.GroupId}}
			err = loadGroupNames(db, u, groupMap)
			if err != nil {
				return pages.HandlerErrorFail("Couldn't load the group", err)
			}
			data.Alias = fmt.Sprintf("%s.%s", groupMap[data.GroupId].Alias, data.Alias)
		}

		// Check if another page is already using the alias
		var existingPageId int64
		row := db.NewStatement(`
					SELECT pageId
					FROM pages
					WHERE isCurrentEdit AND pageId!=? AND deletedBy<=0 AND alias=?`).QueryRow(data.PageId, data.Alias)
		exists, err := row.Scan(&existingPageId)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't read from aliases", err)
		} else if exists {
			return pages.HandlerErrorFail(fmt.Sprintf("Alias '%s' is already in use by: %d", data.Alias, existingPageId), nil)
		}
	}

	// Standardize text
	data.Text = strings.Replace(data.Text, "\r\n", "\n", -1)
	data.Text, err = core.StandardizeLinks(db, data.Text)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't standardize links", err)
	}
	data.MetaText = strings.Replace(data.MetaText, "\r\n", "\n", -1)

	isMinorEditBool := data.IsMinorEditStr == "on"

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

		// Process parents. We'll store them encoded with the edit as a string for
		// keeping historical track, and we also need to update the pagePairs table.
		// TODO: de-duplicate parent ids?
		encodedParentIds := make([]string, 0, len(parentIds))
		pagePairValues := make([]interface{}, 0, len(parentIds)*2)
		for _, id := range parentIds {
			encodedParentIds = append(encodedParentIds, strconv.FormatInt(id, core.PageIdEncodeBase))
			pagePairValues = append(pagePairValues, id, data.PageId)
		}

		// Create a new edit.
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = newEditNum
		hashmap["prevEdit"] = data.PrevEdit
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
		hashmap["isMinorEdit"] = isMinorEditBool
		hashmap["hasVote"] = hasVote
		hashmap["voteType"] = data.VoteType
		hashmap["karmaLock"] = data.KarmaLock
		hashmap["isAutosave"] = data.IsAutosave
		hashmap["isSnapshot"] = data.IsSnapshot
		hashmap["type"] = data.Type
		hashmap["groupId"] = data.GroupId
		hashmap["parents"] = strings.Join(encodedParentIds, ",")
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

		// Update pagePairs tables.
		// TODO: check if parents are actually different from previously published version
		if isCurrentEdit {
			// Delete previous values.
			statement := tx.NewTxStatement(`
				DELETE FROM pagePairs
				WHERE childId=?`)
			_, err = statement.Exec(data.PageId)
			if err != nil {
				return "Couldn't delete old pagePair", err
			}

			if len(pagePairValues) > 0 {
				// Insert new pagePairs values.
				statement := tx.NewTxStatement(`
					INSERT INTO pagePairs (parentId,childId)
					VALUES ` + database.ArgsPlaceholder(len(pagePairValues), 2))
				if _, err = statement.Exec(pagePairValues...); err != nil {
					return "Couldn't insert new pagePairs", err
				}
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
		// Update elastic search index.
		doc := &elastic.Document{
			PageId:    data.PageId,
			Type:      data.Type,
			Title:     data.Title,
			Clickbait: data.Clickbait,
			Text:      data.Text,
			Alias:     data.Alias,
			GroupId:   data.GroupId,
			CreatorId: u.Id,
		}
		err = elastic.AddPageToIndex(c, doc)
		if err != nil {
			c.Errorf("failed to update index: %v", err)
		}

		// Generate updates for users who are subscribed to this page.
		if oldPage.WasPublished && !isMinorEditBool {
			// This is an edit.
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.UpdateType = core.PageEditUpdateType
			task.GroupByPageId = data.PageId
			task.SubscribedToPageId = data.PageId
			task.GoToPageId = data.PageId
			if data.Type == core.CommentPageType {
				// It's actually a comment, so redo some stuff.
				task.UpdateType = core.CommentEditUpdateType
				if commentParentId <= 0 {
					// It's a top level comment.
					task.GroupByPageId = commentPrimaryPageId
					task.SubscribedToPageId = commentPrimaryPageId
				} else {
					// It's actually a reply.
					task.GroupByPageId = commentParentId
					task.SubscribedToPageId = commentParentId
				}
			}
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate updates for users who are subscribed to the author.
		if !oldPage.WasPublished && data.Type != core.CommentPageType && !isMinorEditBool {
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

		if !oldPage.WasPublished && data.Type != core.CommentPageType && !isMinorEditBool {
			// Generate updates for users who are subscribed to the parent pages.
			for _, parentId := range parentIds {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.UpdateType = core.NewChildPageUpdateType
				task.GroupByPageId = parentId
				task.SubscribedToPageId = parentId
				task.GoToPageId = data.PageId
				if err := task.IsValid(); err != nil {
					c.Errorf("Invalid task created: %v", err)
				} else if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}

		if !oldPage.WasPublished && data.Type == core.CommentPageType && !isMinorEditBool {
			// This is a new comment
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

		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.PageId
		if err := task.IsValid(); err != nil {
			c.Errorf("Invalid task created: %v", err)
		} else if err := tasks.Enqueue(c, task, "propagateDomain"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return pages.StatusOK(nil)
}
