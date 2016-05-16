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
	EmailFrequency string `json:"emailFrequency"`
	EmailThreshold int    `json:"emailThreshold"`
	IgnoreMathjax  bool   `json:"ignoreMathJax"`
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

		return "", nil
	})
	if errMessage != "" {
		return pages.Fail(errMessage, err)
	}

	return pages.Success(nil)
}
