// updateSettingsHandler.go updates the settings from the Settings page
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/user"
)

// updateSettingsData contains data given to us in the request.
type updateSettingsData struct {
	EmailFrequency string
	EmailThreshold int
	InviteCode     string
	IgnoreMathjax  bool
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
	if data.EmailFrequency != user.DailyEmailFrequency &&
		data.EmailFrequency != user.WeeklyEmailFrequency &&
		data.EmailFrequency != user.NeverEmailFrequency &&
		data.EmailFrequency != user.ImmediatelyEmailFrequency {
		return pages.HandlerBadRequestFail("EmailFrequency value is invalid", nil)
	}
	if data.EmailThreshold <= 0 {
		return pages.HandlerBadRequestFail("Email Threshold has to be greater than 0", nil)
	}

	row := db.NewStatement(`
		SELECT karma
		FROM users
		WHERE id=?`).QueryRow(u.Id)
	_, err = row.Scan(&u.Karma)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't retrieve a user", err)
	}

	if u.Karma < core.CorrectInviteKarma && data.InviteCode == core.CorrectInviteCode {
		u.Karma = core.CorrectInviteKarma
	}

	statement := db.NewStatement(`
		UPDATE users
		SET emailFrequency=?,emailThreshold=?,inviteCode=?,karma=?,ignoreMathjax=?
		WHERE id=?`)
	_, err = statement.Exec(data.EmailFrequency, data.EmailThreshold,
		data.InviteCode, u.Karma, data.IgnoreMathjax, u.Id)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't update settings", err)
	}

	return pages.StatusOK(nil)
}
