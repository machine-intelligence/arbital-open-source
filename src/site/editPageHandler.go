// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
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
	TagIds         []int64
	IsAutosave     bool
	IsSnapshot     bool
}

type pageTagPair struct {
	tag

	StillInUse bool // true iff this tag is still in use (used in the code)
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
			newEditNum = oldPage.Edit + 1
		}
	}

	// Error checking.
	data.Type = strings.ToLower(data.Type)
	if !data.IsAutosave {
		if len(data.Title) <= 0 || len(data.Text) <= 0 {
			return http.StatusBadRequest, fmt.Sprintf("Need title and text")
		}
		if data.Type != blogPageType && data.Type != wikiPageType {
			return http.StatusBadRequest, fmt.Sprintf("Invalid page type.")
		}
		if data.KarmaLock < 0 || data.KarmaLock > getMaxKarmaLock(u.Karma) {
			return http.StatusBadRequest, fmt.Sprintf("Karma value out of bounds")
		}
	}
	if data.IsAutosave && data.IsSnapshot {
		return http.StatusBadRequest, fmt.Sprintf("Can't be autosave and snapshot")
	}
	if oldPage.PageId > 0 {
		if oldPage.PrivacyKey > 0 && oldPage.PrivacyKey != data.PrivacyKey {
			return http.StatusForbidden, fmt.Sprintf("Need to specify correct privacy key to edit that page")
		}
		if getEditLevel(oldPage, u) < 0 {
			return http.StatusBadRequest, fmt.Sprintf("Not enough karma to edit this page.")
		}
		if oldPage.WasPublished && oldPage.PrivacyKey <= 0 && data.KeepPrivacyKey {
			return http.StatusBadRequest, fmt.Sprintf("Can't change a public page to private.")
		}
	}

	// Check that all the tag ids are valid.
	// TODO: check that you can apply the given tags
	if len(data.TagIds) > 0 {
		var buffer bytes.Buffer
		for _, id := range data.TagIds {
			buffer.WriteString(fmt.Sprintf("%d", id))
			buffer.WriteString(",")
		}
		tagIds := strings.TrimRight(buffer.String(), ",")
		count := 0
		query := fmt.Sprintf(`SELECT COUNT(*) FROM tags WHERE id IN (%s)`, tagIds)
		_, err = database.QueryRowSql(c, query, &count)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't check tags: %v", err)
		}
		if count != len(data.TagIds) {
			return http.StatusBadRequest, fmt.Sprintf("Some of the tags are be invalid: %v", tagIds)
		}
	}

	// Data correction. Rewrite the data structure so that we can just use it
	// in a straight-forward way to populate the database.
	if data.Type == blogPageType {
		data.KarmaLock = 0
	}
	// We can't change page type or voting after it has been published. Also can't
	// turn on privacy.
	hasVote := data.HasVoteStr == "on"
	if oldPage.WasPublished {
		data.Type = oldPage.Type
		hasVote = oldPage.HasVote
	}
	var privacyKey int64
	if data.KeepPrivacyKey {
		if oldPage.PrivacyKey > 0 {
			privacyKey = oldPage.PrivacyKey
		} else {
			privacyKey = rand.Int63()
		}
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

	// Create a new edit.
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["creatorId"] = u.Id
	hashmap["createdAt"] = database.Now()
	hashmap["title"] = data.Title
	hashmap["text"] = data.Text
	hashmap["summary"] = summary
	hashmap["edit"] = newEditNum
	hashmap["isCurrentEdit"] = isCurrentEdit
	hashmap["hasVote"] = hasVote
	hashmap["karmaLock"] = data.KarmaLock
	hashmap["isAutosave"] = data.IsAutosave
	hashmap["isSnapshot"] = data.IsSnapshot
	hashmap["type"] = data.Type
	hashmap["privacyKey"] = privacyKey
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

	// Create a map of tags from the old version of the page.
	oldTagsMap := make(map[int64]*pageTagPair)
	for _, t := range oldPage.Tags {
		oldTagsMap[t.Id] = &pageTagPair{tag: *t}
	}

	// Insert newly added tags and mark the ones still in use.
	// TODO: make this more efficient by creating one sql command
	for _, tagId := range data.TagIds {
		oldPageTag := oldTagsMap[tagId]
		if oldPageTag == nil || !overwritingEdit {
			hashmap := make(map[string]interface{})
			hashmap["tagId"] = tagId
			hashmap["pageId"] = data.PageId
			hashmap["edit"] = newEditNum
			if oldPageTag == nil {
				hashmap["createdBy"] = u.Id
				hashmap["createdAt"] = database.Now()
			} else {
				hashmap["createdBy"] = oldPageTag.PairCreatedBy
				hashmap["createdAt"] = oldPageTag.PairCreatedAt
			}
			query := database.GetInsertSql("pageTagPairs", hashmap)
			_, err = tx.Exec(query)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert a new pageTagPair: %v", err)
			}
		} else {
			oldTagsMap[tagId].StillInUse = true
		}
	}
	// Delete all tags that are not still in use.
	if overwritingEdit {
		for _, pair := range oldTagsMap {
			if pair.StillInUse {
				continue
			}
			query := fmt.Sprintf(`
				DELETE FROM pageTagPairs
				WHERE tagId=%d AND pageId=%d`, pair.Id, data.PageId)
			_, err = tx.Exec(query)
			if err != nil {
				tx.Rollback()
				return http.StatusInternalServerError, fmt.Sprintf("Couldn't delete old pageTagPair: %v", err)
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
			sql := fmt.Sprintf(`
				INSERT INTO links (parentId,childId,createdAt)
				VALUES %s
				ON DUPLICATE KEY UPDATE createdAt = VALUES(createdAt)`, insertValuesStr)
			if _, err = tx.Exec(sql); err != nil {
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

		// Generate updates for users who are subscribed to the tags.
		for _, tagId := range data.TagIds {
			_, isOldTag := oldTagsMap[tagId]
			if oldPage.PageId <= 0 || !isOldTag {
				var task tasks.NewUpdateTask
				task.UserId = u.Id
				task.ContextPageId = data.PageId
				task.ToTagId = tagId
				if oldPage.PageId <= 0 {
					task.UpdateType = newPageWithTagUpdateType
				} else {
					task.UpdateType = addedTagUpdateType
				}
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
