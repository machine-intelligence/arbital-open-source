// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// editPageData contains parameters passed in to create a page.
type editPageData struct {
	PageId         int64 `json:",string"`
	PrevEdit       int   `json:",string"`
	Type           string
	Title          string
	Clickbait      string
	Text           string
	IsMinorEditStr string
	HasVoteStr     string
	VoteType       string
	PrivacyKey     int64 `json:",string"` // if the page is private, this proves that we can access it
	KeepPrivacyKey bool
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
func editPageHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	header, str := editPageProcessor(w, r)
	if header > 0 {
		if header == http.StatusInternalServerError {
			c.Inc(strings.Trim(r.URL.Path, "/") + "Fail")
		}
		c.Errorf("%s", str)
		w.WriteHeader(header)
	}
	if len(str) > 0 {
		fmt.Fprintf(w, "%s", str)
	}
}

func editPageProcessor(w http.ResponseWriter, r *http.Request) (int, string) {
	c := sessions.NewContext(r)
	rand.Seed(time.Now().UnixNano())

	// Decode data
	decoder := json.NewDecoder(r.Body)
	var data editPageData
	err := decoder.Decode(&data)
	if err != nil {
		return http.StatusBadRequest, fmt.Sprintf("Couldn't decode json: %v", err)
	}
	if data.PageId <= 0 {
		return http.StatusBadRequest, fmt.Sprintf("No pageId specified")
	}
	parentIds := make([]int64, 0)
	parentIdArgs := make([]interface{}, 0)
	if data.ParentIds != "" {
		parentStrIds := strings.Split(data.ParentIds, ",")
		for _, parentStrId := range parentStrIds {
			parentId, err := strconv.ParseInt(parentStrId, 10, 64)
			if err != nil {
				return http.StatusBadRequest, fmt.Sprintf("Invalid parent id: %s", parentStrId)
			}
			parentIds = append(parentIds, parentId)
			parentIdArgs = append(parentIdArgs, parentId)
		}
	}

	db, err := database.GetDB(c)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("%v", err)
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load user: %v", err)
	}
	if !u.IsLoggedIn {
		return http.StatusForbidden, ""
	}

	// Load user groups
	if err = loadUserGroups(db, u); err != nil {
		return http.StatusForbidden, fmt.Sprintf("Couldn't load user groups: %v", err)
	}

	// Load the published page.
	var oldPage *core.Page
	oldPage, err = loadFullEdit(db, data.PageId, u.Id, &loadEditOptions{})
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load the old page: %v", err)
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
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load additional page info: %v", err)
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
		return http.StatusBadRequest, fmt.Sprintf("Can't be autosave and snapshot")
	}
	// Check that we have the lock.
	if oldPage.LockedUntil > database.Now() && oldPage.LockedBy != u.Id {
		return http.StatusBadRequest, fmt.Sprintf("Can't change locked page")
	}
	// Check the group settings
	if oldPage.GroupId > 0 {
		if !u.IsMemberOfGroup(oldPage.GroupId) {
			return http.StatusBadRequest, fmt.Sprintf("Don't have group permissions to edit this page")
		}
	}
	// Check PrevEdit number.
	if data.PrevEdit < 0 {
		return http.StatusBadRequest, fmt.Sprintf("PrevEdit number is not valid")
	}
	// TODO: check that this user has access to that edit
	// Check validity of most options. (We are super permissive with autosaves.)
	if !data.IsAutosave {
		if len(data.Title) <= 0 && data.Type != core.CommentPageType {
			return http.StatusBadRequest, fmt.Sprintf("Need title")
		}
		if data.Type != core.WikiPageType &&
			data.Type != core.LensPageType &&
			data.Type != core.QuestionPageType &&
			data.Type != core.AnswerPageType &&
			data.Type != core.CommentPageType {
			return http.StatusBadRequest, fmt.Sprintf("Invalid page type.")
		}
		if data.SortChildrenBy != core.LikesChildSortingOption &&
			data.SortChildrenBy != core.ChronologicalChildSortingOption &&
			data.SortChildrenBy != core.AlphabeticalChildSortingOption {
			return http.StatusBadRequest, fmt.Sprintf("Invalid sort children value.")
		}
		if data.VoteType != "" && data.VoteType != core.ProbabilityVoteType && data.VoteType != core.ApprovalVoteType {
			return http.StatusBadRequest, fmt.Sprintf("Invalid vote type value.")
		}
		if data.KarmaLock < 0 || data.KarmaLock > getMaxKarmaLock(u.Karma) {
			return http.StatusBadRequest, fmt.Sprintf("Karma value out of bounds")
		}
		if data.AnchorContext == "" && data.AnchorText != "" {
			return http.StatusBadRequest, fmt.Sprintf("Anchor context isn't set")
		}
		if data.AnchorContext != "" && data.AnchorText == "" {
			return http.StatusBadRequest, fmt.Sprintf("Anchor text isn't set")
		}
		if data.AnchorOffset < 0 || data.AnchorOffset > len(data.AnchorContext) {
			return http.StatusBadRequest, fmt.Sprintf("Anchor offset out of bounds")
		}
		for _, parentId := range parentIds {
			if parentId == data.PageId {
				return http.StatusBadRequest, fmt.Sprintf("Can't set a page as its own parent")
			}
		}
	}
	if oldPage.WasPublished {
		if oldPage.PrivacyKey > 0 && oldPage.PrivacyKey != data.PrivacyKey {
			return http.StatusForbidden, fmt.Sprintf("Need to specify correct privacy key to edit that page")
		}
		editLevel := getEditLevel(oldPage, u)
		if editLevel != "" && editLevel != "admin" {
			if editLevel == core.CommentPageType {
				return http.StatusBadRequest, fmt.Sprintf("Can't edit a comment page you didn't create.")
			}
			return http.StatusBadRequest, fmt.Sprintf("Not enough karma to edit this page.")
		}
		if oldPage.PrivacyKey <= 0 && data.KeepPrivacyKey {
			return http.StatusBadRequest, fmt.Sprintf("Can't change a public page to private.")
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
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't check parents: %v", err)
			}
			if count != len(parentIds) {
				if data.Type == core.AnswerPageType {
					return http.StatusBadRequest, fmt.Sprintf("Some of the parents are invalid: %v. Perhaps because one of them is not a question.", data.ParentIds)
				}
				return http.StatusBadRequest, fmt.Sprintf("Some of the parents are invalid: %v. Perhaps one of them doesn't exist or is owned by a group you are a not a part of?", data.ParentIds)
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
					return http.StatusBadRequest, fmt.Sprintf("Couldn't load comment's parents: %v", err)
				}
				if commentPrimaryPageId <= 0 {
					return http.StatusBadRequest, fmt.Sprintf("Comment pages need at least one normal page parent")
				}
			}
		}

		if len(parentIds) <= 0 && (data.Type == core.CommentPageType || data.Type == core.AnswerPageType) {
			return http.StatusBadRequest, fmt.Sprintf("%s pages need to have a parent", data.Type)
		}

		// Lens pages need to have exactly one parent.
		if data.Type == core.LensPageType && len(parentIds) != 1 {
			return http.StatusBadRequest, fmt.Sprintf("Lens pages need to have exactly one parent")
		}

		// We can only change the group from to a more relaxed group:
		// personal group -> any group -> no group
		if oldPage.WasPublished && data.GroupId != oldPage.GroupId {
			if oldPage.GroupId != u.Id && data.GroupId != 0 {
				return http.StatusBadRequest, fmt.Sprintf("Can't change group to a more restrictive one")
			}
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
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't get primary page's group name: %v", err)
		}
	}
	// Enforce SortChildrenBy
	if data.Type == core.CommentPageType {
		data.SortChildrenBy = core.ChronologicalChildSortingOption
	} else if data.Type == core.QuestionPageType {
		data.SortChildrenBy = core.LikesChildSortingOption
	}
	// Can't turn on privacy after the page has been published.
	var privacyKey int64
	data.KeepPrivacyKey = false // FOR NOW
	if data.KeepPrivacyKey {
		if oldPage.PrivacyKey > 0 {
			privacyKey = oldPage.PrivacyKey
		} else {
			privacyKey = rand.Int63()
		}
	}
	if data.Alias == "" {
		data.Alias = fmt.Sprintf("%d", data.PageId)
	}
	data.Text = strings.Replace(data.Text, "\r\n", "\n", -1)

	isMinorEditBool := data.IsMinorEditStr == "on"

	// Begin the transaction.
	err = db.Transaction(func(tx *database.Tx) error {
		if isCurrentEdit {
			// Handle isCurrentEdit and clearing previous isCurrentEdit if necessary
			if oldPage.WasPublished {
				statement := tx.NewTxStatement("UPDATE pages SET isCurrentEdit=false WHERE pageId=? AND isCurrentEdit")
				if _, err = statement.Exec(data.PageId); err != nil {
					return fmt.Errorf("Couldn't update isCurrentEdit for old edits: %v", err)
				}
			}

			// Update aliases table.
			aliasRegexp := regexp.MustCompile("^[0-9A-Za-z_]*[A-Za-z_][0-9A-Za-z_]*$")
			if aliasRegexp.MatchString(data.Alias) {
				// The user might be trying to create a new alias.
				var maxSuffix int      // maximum suffix used with this alias
				var existingSuffix int // if this page already used this suffix, this will be set to it
				standardizedName := strings.Replace(strings.ToLower(data.Alias), "_", "", -1)
				row := tx.NewTxStatement(`
					SELECT ifnull(max(suffix),0),ifnull(max(if(pageId=?,suffix,-1)),-1)
					FROM aliases
					WHERE standardizedName=?`).QueryRow(data.PageId, standardizedName)
				_, err := row.Scan(&maxSuffix, &existingSuffix)
				if err != nil {
					return fmt.Errorf("Couldn't read from aliases: %v", err)
				}
				if existingSuffix < 0 {
					suffix := maxSuffix + 1
					data.Alias = fmt.Sprintf("%s-%d", data.Alias, suffix)
					if data.Type == core.QuestionPageType {
						data.Alias = fmt.Sprintf("Q-%s", data.Alias)
					} else if data.Type == core.QuestionPageType {
						data.Alias = fmt.Sprintf("A-%s", data.Alias)
					}
					if isCurrentEdit {
						hashmap := make(map[string]interface{})
						hashmap["fullName"] = data.Alias
						hashmap["standardizedName"] = standardizedName
						hashmap["suffix"] = suffix
						hashmap["pageId"] = data.PageId
						hashmap["creatorId"] = u.Id
						hashmap["createdAt"] = database.Now()
						statement := tx.NewInsertTxStatement("aliases", hashmap)
						if _, err = statement.Exec(); err != nil {
							return fmt.Errorf("Couldn't add an alias: %v", err)
						}
					}
				} else if existingSuffix > 0 {
					data.Alias = fmt.Sprintf("%s-%d", data.Alias, existingSuffix)
				}
			} else if data.Alias != fmt.Sprintf("%d", data.PageId) {
				// Check if we are simply reusing an existing alias.
				var ignore int
				row := tx.NewTxStatement(`
					SELECT 1 FROM aliases
					WHERE pageId=? AND fullName=?`).QueryRow(data.PageId, data.Alias)
				exists, err := row.Scan(&ignore)
				if err != nil {
					return fmt.Errorf("Couldn't check existing alias: %v", err)
				} else if !exists {
					return fmt.Errorf("Invalid alias. Can only contain letters, underscores, and digits. It cannot be a number.")
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
		hashmap["privacyKey"] = privacyKey
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
			return fmt.Errorf("Couldn't insert a new page: %v", err)
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
			return fmt.Errorf("Couldn't update pageInfos: %v", err)
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
				return fmt.Errorf("Couldn't delete old pagePair: %v", err)
			}

			if len(pagePairValues) > 0 {
				// Insert new pagePairs values.
				statement := tx.NewTxStatement(`
					INSERT INTO pagePairs (parentId,childId)
					VALUES ` + database.ArgsPlaceholder(len(pagePairValues), 2))
				if _, err = statement.Exec(pagePairValues...); err != nil {
					return fmt.Errorf("Couldn't insert new pagePairs: %v", err)
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
				return fmt.Errorf("Couldn't add a subscription: %v", err)
			}
		}

		// Update the links table.
		if isCurrentEdit {
			err = core.UpdatePageLinks(tx, data.PageId, data.Text, sessions.GetDomain())
			if err != nil {
				return fmt.Errorf("Couldn't update links: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Error commit a transaction: %v\n", err)
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
			}
			if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate updates for users who are subscribed to the author.
		if !oldPage.WasPublished && data.Type != core.CommentPageType && privacyKey <= 0 && !isMinorEditBool {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.UpdateType = core.NewPageByUserUpdateType
			task.GroupByUserId = u.Id
			task.SubscribedToUserId = u.Id
			task.GoToPageId = data.PageId
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			}
			if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
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
				}
				if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
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
			}
			if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.PageId
		if err := task.IsValid(); err != nil {
			c.Errorf("Invalid task created: %v", err)
		}
		if err := tasks.Enqueue(c, task, "propagateDomain"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	// Return the full page url if the submission was for the current edit.
	if isCurrentEdit {
		privacyAddon := ""
		if privacyKey > 0 {
			privacyAddon = fmt.Sprintf("/%d", privacyKey)
		}
		return 0, fmt.Sprintf("/pages/%s%s", data.Alias, privacyAddon)
	}
	// Return just the privacy key
	return 0, fmt.Sprintf("%d", newEditNum)
}
