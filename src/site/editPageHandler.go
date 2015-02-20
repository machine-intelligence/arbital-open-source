// editPageHandler.go contains the handler for creating a new page / edit.
package site

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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
	PageId     int64 `json:",string"`
	Type       string
	Title      string
	Text       string
	PrivacyKey string // either the actual key or "on", which means we need to create one
	KarmaLock  int
	TagIds     []int64
	IsDraft    bool
	Answer1    string
	Answer2    string
}

type pageDataTag struct {
	tag

	StillInUse bool // true iff this tag is still in use (used in the code)
}

// editPageHandler handles requests to create a new page.
func editPageHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	rand.Seed(time.Now().UnixNano())

	// Decode data
	decoder := json.NewDecoder(r.Body)
	var data pageData
	err := decoder.Decode(&data)
	if err != nil {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUserFromDb(c)
	if err != nil {
		c.Inc("page_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Load old page
	var oldPage page
	oldTagsMap := make(map[int64]*pageDataTag)
	if data.PageId > 0 {
		var pagePtr *page
		pagePtr, err = loadFullPage(c, data.PageId)
		if err != nil {
			c.Errorf("Couldn't load a page: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
		c.Errorf("Need title and text")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.Type != questionPageType && data.Type != blogPageType && data.Type != infoPageType {
		c.Errorf("Invalid page type.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.KarmaLock < 0 || data.KarmaLock > getMaxKarmaLock(u.Karma) {
		c.Errorf("Karma value out of bounds")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.Type == questionPageType && (len(data.Answer1) <= 0 || len(data.Answer2) <= 0) {
		c.Errorf("Both answers need to be given for a question type page.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if oldPage.PageId > 0 {
		if getEditLevel(&oldPage, u) < 0 {
			c.Errorf("Not enough karma to edit this page.")
			w.WriteHeader(http.StatusBadRequest)
			return
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
			c.Errorf("Couldn't check tags.", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if count != len(data.TagIds) {
			c.Errorf("Some of the tags might be invalid.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Data correction.
	if data.IsDraft {
		data.PrivacyKey = "on"
	}
	if data.Type == blogPageType {
		data.KarmaLock = 0
	}

	// Get database
	var db *sql.DB
	db, err = database.GetDB(c)
	if err != nil {
		c.Inc("page_handler_fail")
		c.Errorf("failed to get DB: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Begin the transaction.
	//var result sql.Result
	tx, err := db.Begin()

	// Create a new edit.
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["createdAt"] = database.Now()
	hashmap["type"] = data.Type
	hashmap["title"] = data.Title
	hashmap["text"] = data.Text
	hashmap["karmaLock"] = data.KarmaLock
	if data.PrivacyKey == "on" {
		hashmap["privacyKey"] = rand.Int63()
		data.PrivacyKey = fmt.Sprintf("%d", hashmap["privacyKey"])
	} else if len(data.PrivacyKey) > 0 {
		hashmap["privacyKey"] = data.PrivacyKey
	}
	query := database.GetInsertSql("pages", hashmap)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("page_handler_fail")
		c.Errorf("Couldn't insert a new page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
				c.Inc("page_handler_fail")
				c.Errorf("Couldn't insert a new pageTagPair: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
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
			c.Inc("page_handler_fail")
			c.Errorf("Couldn't delete old pageTagPair: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Add/edit answers.
	if data.Type == questionPageType {
		hashmap = make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["indexId"] = 0
		hashmap["text"] = data.Answer1
		hashmap["createdAt"] = database.Now()
		hashmap["updatedAt"] = database.Now()
		if oldPage.PageId <= 0 || oldPage.Answers[0].Text != data.Answer1 {
			query = database.GetInsertSql("answers", hashmap, "text", "updatedAt")
			_, err = tx.Exec(query)
		}
		if err == nil {
			hashmap["indexId"] = 1
			hashmap["text"] = data.Answer2
			if oldPage.PageId <= 0 || oldPage.Answers[1].Text != data.Answer2 {
				query = database.GetInsertSql("answers", hashmap, "text", "updatedAt")
				_, err = tx.Exec(query)
			}
		}
		if err != nil {
			tx.Rollback()
			c.Inc("page_handler_fail")
			c.Errorf("Couldn't add an answer: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Add subscription.
	hashmap = make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["pageId"] = data.PageId
	hashmap["createdAt"] = database.Now()
	// If the subscription already exists, we just overwrite pageId, which does nothing, but prevents failure.
	query = database.GetInsertSql("subscriptions", hashmap, "pageId")
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("page_handler_fail")
		c.Errorf("Couldn't add a subscription: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	/*
		if data.ParentPageId > 0 {
			// Create new input.
			hashmap = make(map[string]interface{})
			hashmap["parentId"] = data.ParentPageId
			hashmap["childId"] = data.PageId
			hashmap["creatorId"] = u.Id
			hashmap["creatorName"] = u.FullName()
			hashmap["createdAt"] = database.Now()
			hashmap["updatedAt"] = database.Now()
			sql := database.GetInsertSql("inputs", hashmap)
			if _, err = tx.Exec(sql); err != nil {
				tx.Rollback()
				c.Inc("new_input_fail")
				c.Errorf("Couldn't new input: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	*/

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		c.Errorf("Error commit a transaction: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate updates for people who are subscribed to this page.
	var task tasks.NewUpdateTask
	task.UserId = u.Id
	task.PageId = data.PageId
	task.UpdateType = "pageEdit"
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	// Return id of the new page.
	privacyAddon := ""
	if len(data.PrivacyKey) > 0 {
		privacyAddon = fmt.Sprintf("/%s", data.PrivacyKey)
	}
	fmt.Fprintf(w, "/pages/%d%s", data.PageId, privacyAddon)
}
