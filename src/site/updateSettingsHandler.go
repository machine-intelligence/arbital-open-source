// updateSettingsHandler.go updates the settings from the Settings page

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
type updateSettingsData struct {
	EmailFrequency         string `json:"emailFrequency"`
	EmailThreshold         int    `json:"emailThreshold"`
	IgnoreMathjax          bool   `json:"ignoreMathJax"`
	ShowAdvancedEditorMode bool   `json:"showAdvancedEditorMode"`
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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if data.EmailFrequency != core.DailyEmailFrequency &&
		data.EmailFrequency != core.WeeklyEmailFrequency &&
		data.EmailFrequency != core.NeverEmailFrequency &&
		data.EmailFrequency != core.ImmediatelyEmailFrequency {
		return pages.Fail("EmailFrequency value is invalid", nil).Status(http.StatusBadRequest)
	}
	if data.EmailThreshold <= 0 {
		return pages.Fail("Email Threshold has to be greater than 0", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		hashmap := make(database.InsertMap)
		hashmap["id"] = u.ID
		hashmap["emailFrequency"] = data.EmailFrequency
		hashmap["emailThreshold"] = data.EmailThreshold
		hashmap["showAdvancedEditorMode"] = data.ShowAdvancedEditorMode
		hashmap["ignoreMathjax"] = data.IgnoreMathjax
		statement := db.NewInsertStatement("users", hashmap, hashmap.GetKeys()...)
		_, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't update the mark", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
