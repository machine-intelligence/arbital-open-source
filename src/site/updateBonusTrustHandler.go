// updateBonusTrustHandler.go adds new invites to db and auto-claims / sends invite emails
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
type updateBonusTrustData struct {
	UserId         string
	BonusEditTrust int
}

var updateBonusTrustHandler = siteHandler{
	URI:         "/json/updateBonusTrust/",
	HandlerFunc: updateBonusTrustHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func updateBonusTrustHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	var data updateBonusTrustData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		query := database.NewQuery(`
			UPDATE invites
			SET bonusEditTrust=?`, data.BonusEditTrust).Add(`
			WHERE domainId=?`, core.MathDomainId).Add(`
				AND toUserId=?`, data.UserId).ToTxStatement(tx)
		if _, err := query.Exec(); err != nil {
			return sessions.NewError("Couldn't update bonusEditTrust", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
