// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"appengine/search"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// editPageData contains parameters passed in to create a page.
type editPageData struct {
	PageId         int64 `json:",string"`
	Type           string
	Title          string
	Text           string
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
	parentIds := make([]int64, 0, 0)
	if data.ParentIds != "" {
		parentStrIds := strings.Split(data.ParentIds, ",")
		for _, parentStrId := range parentStrIds {
			parentId, err := strconv.ParseInt(parentStrId, 10, 64)
			if err != nil {
				return http.StatusBadRequest, fmt.Sprintf("Invalid parent id: %s", parentStrId)
			}
			parentIds = append(parentIds, parentId)
		}
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load user: %v", err)
	}
	if !u.IsLoggedIn {
		return http.StatusForbidden, ""
	}

	// Load user groups
	if err = loadUserGroups(c, u); err != nil {
		return http.StatusForbidden, fmt.Sprintf("Couldn't load user groups: %v", err)
	}

	// Load the published page.
	var oldPage *core.Page
	oldPage, err = loadFullEditWithOptions(c, data.PageId, u.Id, &loadEditOptions{})
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load the old page: %v", err)
	} else if oldPage == nil {
		oldPage = &core.Page{}
	}
	oldPage.ProcessParents(c, nil)

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
			query := fmt.Sprintf(`
				SELECT COUNT(DISTINCT pageId)
				FROM pages
				WHERE pageId IN (%s) AND (isCurrentEdit OR creatorId=%d) AND %s AND
					(groupId="" OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=%d))
				`, data.ParentIds, u.Id, typeConstraint, u.Id)
			_, err = database.QueryRowSql(c, query, &count)
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
				query := fmt.Sprintf(`
					SELECT pageId,type
					FROM pages
					WHERE pageId IN (%s) AND isCurrentEdit
					`, data.ParentIds)
				err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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

		if len(parentIds) <= 0 && data.Type == core.CommentPageType || data.Type == core.AnswerPageType {
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
		query := fmt.Sprintf(`
			SELECT max(groupId)
			FROM pages
			WHERE pageId IN (%s) AND type!="comment" AND isCurrentEdit`, data.ParentIds)
		_, err := database.QueryRowSql(c, query, &data.GroupId)
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

	// Try to extract the summary out of the text.
	re := regexp.MustCompile("(?ms)^ {0,3}Summary ?: *\n?(.+?)(\n$|\\z)")
	submatches := re.FindStringSubmatch(data.Text)
	summary := ""
	if len(submatches) > 0 {
		summary = strings.TrimSpace(submatches[1])
	} else {
		// If no summary tags, just extract the first line.
		re := regexp.MustCompile("^(.*)")
		submatches := re.FindStringSubmatch(data.Text)
		summary = strings.TrimSpace(submatches[1])
	}

	// Begin the transaction.
	tx, err := database.NewTransaction(c)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("failed to create a transaction: %v\n", err)
	}

	if isCurrentEdit {
		// Handle isCurrentEdit and clearing previous isCurrentEdit if necessary
		if oldPage.WasPublished {
			query := fmt.Sprintf("UPDATE pages SET isCurrentEdit=false WHERE pageId=%d AND isCurrentEdit", data.PageId)
			if _, err = tx.Exec(query); err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't update isCurrentEdit for old edits: %v", err)
			}
		}

		// Update aliases table.
		aliasRegexp := regexp.MustCompile("^[0-9A-Za-z_]*[A-Za-z_][0-9A-Za-z_]*$")
		if aliasRegexp.MatchString(data.Alias) {
			// The user might be trying to create a new alias.
			var maxSuffix int      // maximum suffix used with this alias
			var existingSuffix int // if this page already used this suffix, this will be set to it
			standardizedName := strings.Replace(strings.ToLower(data.Alias), "_", "", -1)
			query := fmt.Sprintf(`
				SELECT ifnull(max(suffix),0),ifnull(max(if(pageId=%d,suffix,-1)),-1)
				FROM aliases
				WHERE standardizedName="%s"`, data.PageId, standardizedName)
			_, err := database.QueryRowSql(c, query, &maxSuffix, &existingSuffix)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't read from aliases: %v", err)
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
					query = database.GetInsertSql("aliases", hashmap)
					if _, err = tx.Exec(query); err != nil {
						tx.Rollback()
						return http.StatusInternalServerError, fmt.Sprintf("Couldn't add an alias: %v", err)
					}
				}
			} else if existingSuffix > 0 {
				data.Alias = fmt.Sprintf("%s-%d", data.Alias, existingSuffix)
			}
		} else if data.Alias != fmt.Sprintf("%d", data.PageId) {
			// Check if we are simply reusing an existing alias.
			var ignore int
			query := fmt.Sprintf(`SELECT 1 FROM aliases WHERE pageId=%d AND fullName="%s"`, data.PageId, data.Alias)
			exists, err := database.QueryRowSql(c, query, &ignore)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't check existing alias: %v", err)
			} else if !exists {
				tx.Rollback()
				return http.StatusBadRequest, fmt.Sprintf("Invalid alias. Can only contain letters, underscores, and digits. It cannot be a number.")
			}
		}
	}

	// Create encoded string for parents, as well as a string for updating pagePairs.
	// TODO: de-duplicate parent ids
	encodedParentIds := make([]string, len(parentIds))
	pagePairValues := make([]string, len(parentIds))
	for i, id := range parentIds {
		encodedParentIds[i] = strconv.FormatInt(id, core.PageIdEncodeBase)
		pagePairValues[i] = fmt.Sprintf("(%d, %d)", id, data.PageId)
	}

	// Create a new edit.
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["creatorId"] = u.Id
	hashmap["title"] = data.Title
	hashmap["text"] = data.Text
	hashmap["summary"] = summary
	hashmap["alias"] = data.Alias
	hashmap["sortChildrenBy"] = data.SortChildrenBy
	hashmap["edit"] = newEditNum
	hashmap["isCurrentEdit"] = isCurrentEdit
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
	query := ""
	if overwritingEdit {
		query = database.GetReplaceSql("pages", hashmap)
	} else {
		query = database.GetInsertSql("pages", hashmap)
	}
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert a new page: %v", err)
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
		query = database.GetInsertSql("pageInfos", hashmap, "maxEdit", "currentEdit", "lockedUntil")
	} else {
		query = database.GetInsertSql("pageInfos", hashmap, "maxEdit")
	}
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't update pageInfos: %v", err)
	}

	// Update pagePairs tables.
	// TODO: check if parents are actually different from previously published version
	if isCurrentEdit {
		// Delete previous values.
		query := fmt.Sprintf(`
			DELETE FROM pagePairs
			WHERE childId=%d`, data.PageId)
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't delete old pagePair: %v", err)
		}

		if len(pagePairValues) > 0 {
			// Insert new pagePairs values.
			insertValuesStr := strings.Join(pagePairValues, ",")
			query := fmt.Sprintf(`
				INSERT INTO pagePairs (parentId,childId)
				VALUES %s`, insertValuesStr)
			if _, err = tx.Exec(query); err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert new pagePairs: %v", err)
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
		query = database.GetInsertSql("subscriptions", hashmap, "userId")
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't add a subscription: %v", err)
		}
	}

	// Update the links table.
	if isCurrentEdit {
		// Delete old links.
		if oldPage.WasPublished {
			query = fmt.Sprintf("DELETE FROM links WHERE parentId=%d", data.PageId)
			_, err = tx.Exec(query)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't delete old links: %v", err)
			}
		}
		// NOTE: these regexps are waaaay too simplistic and don't account for the
		// entire complexity of Markdown, like 4 spaces, backticks, and escaped
		// brackets / parens.
		aliasesAndIds := make([]string, 0, 0)
		extractLinks := func(exp *regexp.Regexp) {
			submatches := exp.FindAllStringSubmatch(data.Text, -1)
			for _, submatch := range submatches {
				aliasesAndIds = append(aliasesAndIds, submatch[1])
			}
		}
		// Find directly encoded urls
		extractLinks(regexp.MustCompile(regexp.QuoteMeta(getConfigAddress()) + "/pages/([0-9]+)"))
		// Find ids and aliases using [id/alias] syntax.
		extractLinks(regexp.MustCompile("\\[([A-Za-z0-9_-]+?)\\](?:[^(]|$)"))
		// Find ids and aliases using [text](id/alias) syntax.
		extractLinks(regexp.MustCompile("\\[.+?\\]\\(([A-Za-z0-9_-]+?)\\)"))
		if len(aliasesAndIds) > 0 {
			// Populate linkTuples
			linkMap := make(map[string]bool) // track which aliases we already added to the list
			linkTuples := make([]string, 0, 0)
			for _, alias := range aliasesAndIds {
				if linkMap[alias] {
					continue
				}
				insertValue := fmt.Sprintf("(%d, '%s')", data.PageId, alias)
				linkTuples = append(linkTuples, insertValue)
				linkMap[alias] = true
			}

			// Insert all the tuples into the links table.
			linkTuplesStr := strings.Join(linkTuples, ",")
			query = fmt.Sprintf(`
				INSERT INTO links (parentId,childAlias)
				VALUES %s`, linkTuplesStr)
			if _, err = tx.Exec(query); err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert links: %v", err)
			}
		}
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Error commit a transaction: %v\n", err)
	}

	// === Once the transaction has succeeded, we can't really fail on anything
	// else. So we print out errors, but don't return an error. ===

	if isCurrentEdit {

		// Update pages' index.
		p := &tasks.PageIndexDoc{}
		p.PageId = search.Atom(fmt.Sprintf("%d", data.PageId))
		p.Type = data.Type
		p.Title = data.Title
		p.Text = data.Text
		p.Alias = search.Atom(data.Alias)

		index, err := search.Open("pages")
		if err != nil {
			c.Errorf("failed to open index: %v", err)
		} else {
			_, err = index.Put(c, string(p.PageId), p)
			if err != nil {
				c.Errorf("failed to put page into index: %v", err)
			}
		}

		// Generate updates for users who are subscribed to this page.
		if oldPage.WasPublished {
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
		if !oldPage.WasPublished && data.Type != core.CommentPageType && privacyKey <= 0 {
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

		if !oldPage.WasPublished && data.Type != core.CommentPageType {
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

		if !oldPage.WasPublished && data.Type == core.CommentPageType {
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
	return 0, fmt.Sprintf("%d", privacyKey)
}
