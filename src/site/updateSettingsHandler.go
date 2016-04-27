// updateSettingsHandler.go updates the settings from the Settings page
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
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

	// Update user model from settings form
	statement := db.NewStatement(`
		UPDATE users
		SET emailFrequency=?,emailThreshold=?,ignoreMathjax=?
		WHERE id=?`)
	_, err = statement.Exec(data.EmailFrequency, data.EmailThreshold, data.IgnoreMathjax, u.Id)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't update settings", err)
	}

	// If there's no newInviteCode submitted, return here
	if data.NewInviteCodeClaimed != "" {
		// Check if it matches a general code or domain code. Then process.
		// Search for code in inviteeEmailPairs (joined with invites, to get domainId)
		match, err := core.MatchInvite(db, data.NewInviteCodeClaimed, u.Email)
		if err != nil {
			return pages.HandlerErrorFail("Unable to match invite code", err)
		}
		// If code is not in db, return as bad
		if !match.CodeMatch {
			return pages.HandlerBadRequestFail("Not an invite code", nil)
		}
		// If code is in claimed in db and is personal, can't be used again
		if match.Invitee.ClaimingUserId != "" && match.Invite.Type == core.PersonalInviteType {
			return pages.HandlerBadRequestFail("Single-use invite already claimed", nil)
		}

		// Claim code for user, and send invite to UI
		_, err = core.ClaimCode(db, data.NewInviteCodeClaimed, match.Invite.DomainId, u)
		if err != nil {
			return pages.HandlerBadRequestFail("Couldn't claim code", err)
		}
		returnData.ResultMap["invite"] = match.Invite
	}

	return pages.StatusOK(returnData)
}
