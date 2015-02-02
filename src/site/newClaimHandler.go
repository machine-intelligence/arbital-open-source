// newClaim.go creates a new claim
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

// newClaimData contains parameters passed in to create a new claim
type newClaimData struct {
	Text    string
	Private string
	TagId   int64 `json:",string"`
}

// newClaimHandler handles requests to create a new claim.
func newClaimHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newClaimData
	err := decoder.Decode(&data)
	if err != nil || len(data.Text) <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_claim_fail")
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
		c.Inc("new_claim_fail")
		c.Errorf("failed to get DB: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Begin the transaction.
	var result sql.Result
	tx, err := db.Begin()

	// Create new claim.
	var privacyKey int64
	hashmap := make(map[string]interface{})
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["text"] = data.Text
	if data.Private == "on" {
		rand.Seed(time.Now().UnixNano())
		privacyKey = rand.Int63()
		hashmap["privacyKey"] = privacyKey
	}
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	query := database.GetInsertSql("claims", hashmap)
	result, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("new_claim_fail")
		c.Errorf("Couldn't insert a new claim: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var claimId int64
	claimId, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		c.Inc("new_claim_fail")
		c.Errorf("Couldn't get claim id from result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add claim tag.
	if data.TagId > 0 {
		hashmap = make(map[string]interface{})
		hashmap["claimId"] = claimId
		hashmap["tagId"] = data.TagId
		hashmap["createdBy"] = u.Id
		hashmap["createdAt"] = database.Now()
		query = database.GetInsertSql("claimTagPairs", hashmap)
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			c.Inc("new_claim_fail")
			c.Errorf("Couldn't add a new claimTagPair: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Add subscription.
	hashmap = make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["claimId"] = claimId
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("subscriptions", hashmap)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		c.Inc("update_claim_fail")
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

	// Return id of the new claim.
	privacyAddon := ""
	if privacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", privacyKey)
	}
	fmt.Fprintf(w, "/claims/%d%s", claimId, privacyAddon)
}
