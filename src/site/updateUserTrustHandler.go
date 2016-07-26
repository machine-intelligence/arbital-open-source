// updateUserTrustHandler.go updates user's trust values

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// updateSettingsData contains data given to us in the request.
type updateUserTrustData struct {
	UserID    string
	DomainID  string
	EditTrust int
}

var updateUserTrustHandler = siteHandler{
	URI:         "/json/updateUserTrust/",
	HandlerFunc: updateUserTrustHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func updateUserTrustHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data updateUserTrustData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	var existingEditTrust int
	row := database.NewQuery(`
		SELECT editTrust
		FROM userTrust
		WHERE userId=?`, data.UserID).Add(`
			AND domainId=?`, data.DomainID).ToStatement(db).QueryRow()
	hasExistingEditTrust, err := row.Scan(&existingEditTrust)
	if err != nil {
		return pages.Fail("Couldn't query for existing edit trust", err)
	}
	if hasExistingEditTrust && data.EditTrust == existingEditTrust {
		// If the edit trust for this user and domain hasn't changed, we're done.
		return pages.Success(nil)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Create/update user trust.
		hashmap := make(map[string]interface{})
		hashmap["userId"] = data.UserID
		hashmap["domainId"] = data.DomainID
		hashmap["editTrust"] = data.EditTrust
		statement := db.NewInsertStatement("userTrust", hashmap, "editTrust")
		if _, err := statement.WithTx(tx).Exec(); err != nil {
			return sessions.NewError("Couldn't update/create userTrust row", err)
		}

		// Send them an update, iff they were promoted to Reviewer.
		if (!hasExistingEditTrust || existingEditTrust < core.ReviewerKarmaLevel) && core.ReviewerKarmaLevel <= data.EditTrust {
			hashmap = make(map[string]interface{})
			hashmap["userId"] = data.UserID
			hashmap["type"] = core.UserTrustUpdateType
			hashmap["createdAt"] = database.Now()
			hashmap["subscribedToId"] = data.DomainID
			hashmap["goToPageId"] = data.DomainID
			hashmap["byUserId"] = u.ID
			statement = db.NewInsertStatement("updates", hashmap).WithTx(tx)
			if _, err = statement.Exec(); err != nil {
				return sessions.NewError("Couldn't add a new update", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
