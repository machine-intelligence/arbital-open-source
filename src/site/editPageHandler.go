// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"bytes"
	"database/sql"
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
	// Optional page id. If it's passed in, we are editing a page, otherwise it's a new page.
	PageId  int64 `json:",string"`
	Type    string
	Title   string
	Text    string
	HasVote string
	// If <0, the user is turning the privacy key off. If zero, the user
	// wants to create a new privacy key. If >0, keep the old key.
	PrivacyKey int64 `json:",string"`
	KarmaLock  int
	TagIds     []int64
	IsDraft    bool
}

type pageDataTag struct {
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
	var oldPage page
	oldTagsMap := make(map[int64]*pageDataTag)
	if data.PageId > 0 {
		var pagePtr *page
		pagePtr, err = loadFullPage(c, data.PageId)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't load a page: %v", err)
		}

		oldPage = *pagePtr
		for _, t := range oldPage.Tags {
			oldTagsMap[t.Id] = &pageDataTag{tag: *t}
		}
	} else {
		data.PageId = rand.Int63()
	}

	// Error checking.
	data.Type = strings.ToLower(data.Type)
	if len(data.Title) <= 0 || len(data.Text) <= 0 {
		return http.StatusBadRequest, fmt.Sprintf("Need title and text")
	}
	if data.Type != blogPageType && data.Type != wikiPageType {
		return http.StatusBadRequest, fmt.Sprintf("Invalid page type.")
	}
	if data.KarmaLock < 0 || data.KarmaLock > getMaxKarmaLock(u.Karma) {
		return http.StatusBadRequest, fmt.Sprintf("Karma value out of bounds")
	}
	if oldPage.PageId > 0 {
		if oldPage.IsDraft {
			if u.Id != oldPage.Author.Id {
				return http.StatusBadRequest, fmt.Sprintf("You can't edit this draft because you didn't create it.")
			}
		} else {
			if data.IsDraft {
				return http.StatusBadRequest, fmt.Sprintf("Can't save a draft once the page has been published.")
			}
			if getEditLevel(&oldPage, u) < 0 {
				return http.StatusBadRequest, fmt.Sprintf("Not enough karma to edit this page.")
			}
		}
		if oldPage.PrivacyKey <= 0 && data.PrivacyKey >= 0 {
			return http.StatusBadRequest, fmt.Sprintf("Can't change a public page to private.")
		}
	}

	// Check that all the tag ids are valid.
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
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't check tags.", err)
		}
		if count != len(data.TagIds) {
			return http.StatusBadRequest, fmt.Sprintf("Some of the tags might be invalid.")
		}
	}

	// Data correction. Rewrite the data structure so that we can just use it
	// in a straight-forward way to populate the database.
	if data.IsDraft {
		data.PrivacyKey = oldPage.PrivacyKey
	}
	if data.Type == blogPageType {
		data.KarmaLock = 0
	}
	// We can't change page type after it has been published.
	if oldPage.PageId > 0 && !oldPage.IsDraft {
		data.Type = oldPage.Type
	}
	if data.PrivacyKey == 0 {
		data.PrivacyKey = rand.Int63()
	} else if data.PrivacyKey < 0 {
		data.PrivacyKey = 0
	}

	// Get database
	var db *sql.DB
	db, err = database.GetDB(c)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("failed to get DB: %v\n", err)
	}

	// Begin the transaction.
	tx, err := db.Begin()

	// To simplify our life significantly when we are trying to find the most recent edit,
	// we shift "edit" variable for all edits of this page up by one. Then the most recent
	// edit will always have edit=0.
	if oldPage.PageId > 0 {
		query := fmt.Sprintf("UPDATE pages SET edit=edit+1 WHERE pageId=%d ORDER BY edit DESC", data.PageId)
		if _, err = tx.Exec(query); err != nil {
			tx.Rollback()
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't update edit for old edits: %v", err)
		}
	}

	// Create a new edit.
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["creatorId"] = u.Id
	hashmap["createdAt"] = database.Now()
	hashmap["title"] = data.Title
	hashmap["text"] = data.Text
	hashmap["hasVote"] = data.HasVote == "on"
	hashmap["karmaLock"] = data.KarmaLock
	hashmap["isDraft"] = data.IsDraft
	hashmap["type"] = data.Type
	hashmap["privacyKey"] = data.PrivacyKey
	query := database.GetInsertSql("pages", hashmap)
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't insert a new page: %v", err)
	}

	// Insert newly added tags and mark the ones still in use.
	for _, tagId := range data.TagIds {
		if oldTagsMap[tagId] == nil {
			hashmap := make(map[string]interface{})
			hashmap["tagId"] = tagId
			hashmap["pageId"] = data.PageId
			hashmap["createdBy"] = u.Id
			hashmap["createdAt"] = database.Now()
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

	// Add subscription.
	if oldPage.PageId <= 0 {
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
	if oldPage.Text != data.Text {
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

	// Generate updates for users who are subscribed to this page.
	if oldPage.PageId > 0 {
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
	if oldPage.PageId <= 0 && data.Type == blogPageType {
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

	// Return id of the new page.
	privacyAddon := ""
	if data.PrivacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", data.PrivacyKey)
	}
	editPart := ""
	if data.IsDraft {
		editPart = "/edit"
	}
	return 0, fmt.Sprintf("/pages%s/%d%s", editPart, data.PageId, privacyAddon)
}
