// updateSettingsHandler.go updates the settings from the Settings page
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updateSettingsData contains data given to us in the request.
type updateSettingsData struct {
	EmailFrequency       string `json:"emailFrequency"`
	EmailThreshold       int    `json:"emailThreshold"`
	NewInviteCodeClaimed string `json:"newInviteCodeClaimed"`
	IgnoreMathjax        bool   `json:"ignoreMathJax"`
}

var updateSettingsHandler = siteHandler{
	URI:         "/updateSettings/",
	HandlerFunc: updateSettingsHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func updateSettingsHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(params.U, true)

	var data updateSettingsData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.EmailFrequency != core.DailyEmailFrequency &&
		data.EmailFrequency != core.WeeklyEmailFrequency &&
		data.EmailFrequency != core.NeverEmailFrequency &&
		data.EmailFrequency != core.ImmediatelyEmailFrequency {
		return pages.HandlerBadRequestFail("EmailFrequency value is invalid", nil)
	}
	if data.EmailThreshold <= 0 {
		return pages.HandlerBadRequestFail("Email Threshold has to be greater than 0", nil)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Update user model from settings form
		statement := db.NewStatement(`
		UPDATE users
		SET emailFrequency=?,emailThreshold=?,ignoreMathjax=?
		WHERE id=?`).WithTx(tx)
		_, err = statement.Exec(data.EmailFrequency, data.EmailThreshold, data.IgnoreMathjax, u.Id)
		if err != nil {
			return "Couldn't update settings", err
		}

		// If there's no newInviteCode submitted, return here
		if data.NewInviteCodeClaimed != "" {
			// Claim code for user, and send invite to UI
			invite, err := core.ClaimCode(tx, data.NewInviteCodeClaimed, u.Email, u.Id)
			if err != nil {
				return "Couldn't claim code", err
			}
			returnData.ResultMap["invite"] = invite
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	return pages.StatusOK(returnData)
}
