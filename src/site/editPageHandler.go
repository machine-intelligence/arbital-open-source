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

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// pageData contains parameters passed in to create a page.
type pageData struct {
	PageId         int64 `json:",string"`
	Type           string
	Title          string
	Text           string
	HasVoteStr     string
	PrivacyKey     int64 `json:",string"` // if the page is private, this proves that we can access it
	KeepPrivacyKey bool
	KarmaLock      int
	ParentIds      string
	Alias          string // if empty, leave the current one
	SortChildrenBy string
	IsAutosave     bool
	IsSnapshot     bool
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
	var data pageData
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

	// Load old page
	var oldPage *page
	oldPage, err = loadFullEdit(c, data.PageId, u.Id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load the old page: %v", err)
	} else if oldPage == nil {
		oldPage = &page{}
	}

	// Compute edit number for the new edit of this page
	newEditNum := 0
	if oldPage.PageId > 0 {
		if oldPage.IsAutosave {
			newEditNum = oldPage.Edit
		} else {
			newEditNum = oldPage.MaxEditEver + 1
		}
	}

	// Error checking.
	data.Type = strings.ToLower(data.Type)
	if data.IsAutosave && data.IsSnapshot {
		return http.StatusBadRequest, fmt.Sprintf("Can't be autosave and snapshot")
	}
	if !data.IsAutosave {
		if len(data.Title) <= 0 || len(data.Text) <= 0 {
			return http.StatusBadRequest, fmt.Sprintf("Need title and text")
		}
		if data.Type != blogPageType && data.Type != wikiPageType {
			return http.StatusBadRequest, fmt.Sprintf("Invalid page type.")
		}
		if data.SortChildrenBy != likesChildSortingOption &&
			data.SortChildrenBy != chronologicalChildSortingOption &&
			data.SortChildrenBy != alphabeticalChildSortingOption {
			return http.StatusBadRequest, fmt.Sprintf("Invalid sort children value.")
		}
		if data.KarmaLock < 0 || data.KarmaLock > getMaxKarmaLock(u.Karma) {
			return http.StatusBadRequest, fmt.Sprintf("Karma value out of bounds")
		}
		for _, parentId := range parentIds {
			if parentId == data.PageId {
				return http.StatusBadRequest, fmt.Sprintf("Can't set a page as its own parent")
			}
		}
	}
	if oldPage.PageId > 0 {
		if oldPage.PrivacyKey > 0 && oldPage.PrivacyKey != data.PrivacyKey {
			return http.StatusForbidden, fmt.Sprintf("Need to specify correct privacy key to edit that page")
		}
		editLevel := getEditLevel(oldPage, u)
		if editLevel != "" && editLevel != "admin" {
			return http.StatusBadRequest, fmt.Sprintf("Not enough karma to edit this page.")
		}
		if oldPage.WasPublished && oldPage.PrivacyKey <= 0 && data.KeepPrivacyKey {
			return http.StatusBadRequest, fmt.Sprintf("Can't change a public page to private.")
		}
	}

	// Check that all the parent ids are valid.
	// TODO: check that you can apply the given parent ids
	if len(parentIds) > 0 {
		count := 0
		query := fmt.Sprintf(`SELECT COUNT(*) FROM pages WHERE pageId IN (%s) AND isCurrentEdit`, data.ParentIds)
		_, err = database.QueryRowSql(c, query, &count)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't check parents: %v", err)
		}
		if count != len(parentIds) {
			return http.StatusBadRequest, fmt.Sprintf("Some of the parents are invalid: %v", data.ParentIds)
		}
	}

	// Data correction. Rewrite the data structure so that we can just use it
	// in a straight-forward way to populate the database.
	// Can't change page type or voting after it has been published.
	hasVote := data.HasVoteStr == "on"
	if oldPage.WasPublished {
		data.Type = oldPage.Type
		hasVote = oldPage.HasVote
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
	re := regexp.MustCompile("(?ms)^ {0,3}<summary> *\n(.+?)\n {0,3}</summary> *$")
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

	// Handle isCurrentEdit and clearing previous isCurrentEdit if necessary
	isCurrentEdit := !data.IsAutosave && !data.IsSnapshot
	if oldPage.PageId > 0 && isCurrentEdit {
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
	hashmap["karmaLock"] = data.KarmaLock
	hashmap["isAutosave"] = data.IsAutosave
	hashmap["isSnapshot"] = data.IsSnapshot
	hashmap["type"] = data.Type
	hashmap["privacyKey"] = privacyKey
	hashmap["createdAt"] = database.Now()
	query := ""
	overwritingEdit := oldPage.PageId > 0 && oldPage.Edit == newEditNum
	if overwritingEdit {
		query = database.GetReplaceSql("pages", hashmap)
	} else {
		query = database.GetInsertSql("pages", hashmap)
	}
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert a new page: %v", err)
	}

	// Create a map of relationships from the old version of the page.
	oldParentsMap := make(map[int64]*pagePair)
	for _, r := range oldPage.Parents {
		oldParentsMap[r.Parent.PageId] = r
	}

	// Insert newly added pages and mark the ones still in use.
	// TODO: make this more efficient by creating one sql command
	for _, parentId := range parentIds {
		oldPageParent := oldParentsMap[parentId]
		if oldPageParent == nil || !overwritingEdit {
			hashmap := make(map[string]interface{})
			hashmap["parentId"] = parentId
			hashmap["childId"] = data.PageId
			hashmap["childEdit"] = newEditNum
			query := database.GetInsertSql("pagePairs", hashmap)
			_, err = tx.Exec(query)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert a new pagePair: %v", err)
			}
		} else {
			oldParentsMap[parentId].StillInUse = true
		}
	}
	// Delete all parents that are not still in use.
	// TODO: make this more efficient by creating one sql command
	if overwritingEdit {
		for _, pair := range oldParentsMap {
			if pair.StillInUse {
				continue
			}
			query := fmt.Sprintf(`
				DELETE FROM pagePairs
				WHERE id=%d`, pair.Id)
			_, err = tx.Exec(query)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't delete old pagePair: %v", err)
			}
		}
	}

	// Add subscription.
	if isCurrentEdit && !oldPage.WasPublished {
		hashmap = make(map[string]interface{})
		hashmap["userId"] = u.Id
		hashmap["toPageId"] = data.PageId
		hashmap["createdAt"] = database.Now()
		query = database.GetInsertSql("subscriptions", hashmap)
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't add a subscription: %v", err)
		}
	}

	// Update the links table.
	if isCurrentEdit {
		// Delete old links.
		if oldPage.PageId > 0 {
			query = fmt.Sprintf("DELETE FROM links WHERE parentId=%d", data.PageId)
			_, err = tx.Exec(query)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't delete old links: %v", err)
			}
		}
		re := regexp.MustCompile(getConfigAddress() + "/pages/[0-9]+")
		links := re.FindAllString(data.Text, -1)
		insertValues := make([]string, 0, 0)
		if len(links) > 0 {
			for _, link := range links {
				pageIdStr := link[strings.LastIndex(link, "/")+1:]
				insertValue := fmt.Sprintf("(%d, %s, '%s')", data.PageId, pageIdStr, database.Now())
				insertValues = append(insertValues, insertValue)
			}
			insertValuesStr := strings.Join(insertValues, ",")
			query := fmt.Sprintf(`
				INSERT INTO links (parentId,childId,createdAt)
				VALUES %s
				ON DUPLICATE KEY UPDATE createdAt = VALUES(createdAt)`, insertValuesStr)
			if _, err = tx.Exec(query); err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert new visits: %v", err)
			}
		}
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Error commit a transaction: %v\n", err)
	}

	if isCurrentEdit {
		// Generate updates for users who are subscribed to this page.
		if oldPage.WasPublished {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.ContextPageId = data.PageId
			task.ToPageId = data.PageId
			task.UpdateType = pageEditUpdateType
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			}
			if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate updates for users who are subscribed to the author.
		if !oldPage.WasPublished && data.Type == blogPageType && privacyKey <= 0 {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.ContextPageId = data.PageId
			task.ToUserId = u.Id
			task.UpdateType = newPageByUserUpdateType
			if err := task.IsValid(); err != nil {
				c.Errorf("Invalid task created: %v", err)
			}
			if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}

		// Generate updates for users who are subscribed to the parent pages.
		if !oldPage.WasPublished {
			for _, parentId := range parentIds {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.ContextPageId = data.PageId
				task.ToPageId = parentId
				task.UpdateType = newChildPageUpdateType
				if err := task.IsValid(); err != nil {
					c.Errorf("Invalid task created: %v", err)
				}
				if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
					c.Errorf("Couldn't enqueue a task: %v", err)
				}
			}
		}
	}

	// Return the full page url if the submission was for the current edit.
	if isCurrentEdit {
		privacyAddon := ""
		if privacyKey > 0 {
			privacyAddon = fmt.Sprintf("/%d", privacyKey)
		}
		return 0, fmt.Sprintf("/pages/%d%s", data.PageId, privacyAddon)
	}
	// Return just the privacy key
	return 0, fmt.Sprintf("%d", privacyKey)
}
