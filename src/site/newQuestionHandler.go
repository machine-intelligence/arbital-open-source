// newQuestion.go creates a new question
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newQuestionData contains parameters passed in to create a new question
type newQuestionData struct {
	Text       string
	Answer1    string
	Answer2    string
	Input1Text string
	Input2Text string
	Private    string
	PriorVote  float32 `json:",string"`
}

// newQuestionHandler handles requests to create a new question.
func newQuestionHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newQuestionData
	err := decoder.Decode(&data)
	if err != nil || len(data.Text) <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.PriorVote < 0 || data.PriorVote > 100 {
		c.Errorf("Value has to be between 0 and 100 inclusive")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_question_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Get database
	var db *sql.DB
	db, err = database.GetDB(c)
	if err != nil {
		c.Inc("new_question_fail")
		c.Errorf("failed to get DB: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Begin the transaction.
	var result sql.Result
	tx, err := db.Begin()

	// Create new question.
	var privacyKey int64
	hashmap := make(map[string]interface{})
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["text"] = data.Text
	hashmap["answer1"] = data.Answer1
	hashmap["answer2"] = data.Answer2
	if data.Private == "on" {
		rand.Seed(time.Now().UnixNano())
		privacyKey = rand.Int63()
		hashmap["privacyKey"] = privacyKey
	}
	hashmap["createdAt"] = database.Now()
	query := database.GetInsertSql("questions", hashmap)
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't insert a new question: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var questionId int64
	questionId, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't get question id from result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add input for 1.
	hashmap = make(map[string]interface{})
	hashmap["questionId"] = questionId
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["text"] = data.Input1Text
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("inputs", hashmap)
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't insert a new input: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var inputId1 int64
	inputId1, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't get input id from result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add input for 2.
	hashmap["text"] = data.Input2Text
	query = database.GetInsertSql("inputs", hashmap)
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't insert a new input: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var inputId2 int64
	inputId2, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't get input id from result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update question.
	hashmap = make(map[string]interface{})
	hashmap["id"] = questionId
	hashmap["inputId1"] = inputId1
	hashmap["inputId2"] = inputId2
	query = database.GetInsertSql("questions", hashmap, "inputId1", "inputId2")
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("new_question_fail")
		c.Errorf("Couldn't update a new question: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Add new vote.
	hashmap = make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["questionId"] = questionId
	hashmap["createdAt"] = database.Now()
	hashmap["value"] = data.PriorVote
	query = database.GetInsertSql("priorVotes", hashmap)
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("update_question_fail")
		c.Errorf("Couldn't add a prior vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add subscription.
	hashmap = make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["questionId"] = questionId
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("subscriptions", hashmap)
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("update_question_fail")
		c.Errorf("Couldn't add a subscription: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		c.Errorf("Error commit a transaction: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return id of the new question.
	privacyAddon := ""
	if privacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", privacyKey)
	}
	fmt.Fprintf(w, "/questions/%d%s", questionId, privacyAddon)
}
