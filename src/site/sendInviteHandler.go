// sendInviteHandler.go adds new invites to db, restores deleted ones ("undo delete") and sends invite emails
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// updateSettingsData contains data given to us in the request.
type sendInviteData struct {
	Type      string          `json:"type"`
	DomainId  string          `json:"domainId"`
	OldCode   string          `json:"oldCode"`
	IsUpdate  bool            `json:"isUpdate"`
	ClaimedAt string          `json:"claimedAt"`
	Invitees  []*core.Invitee `json:"invitees"`
}

var sendInviteHandler = siteHandler{
	URI:         "/sendInvite/",
	HandlerFunc: sendInviteHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin:   true,
		RequireTrusted: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func sendInviteHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	u := params.U
	returnData := core.NewHandlerData(params.U, false)

	var data sendInviteData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.DomainId) && data.DomainId != "" {
		return pages.HandlerBadRequestFail("DomainId is invalid", nil)
	}
	if len(data.Invitees) <= 0 {
		return pages.StatusOK(returnData)
	}

	var inviteCode string
	if data.OldCode == "" {
		// Generate a new, unique invite code
		inviteCode, err = GenerateUniqueInviteCode(db)
		if err != nil {
			c.Errorf("Error generating new, unique code %v\n", err)
		}
		returnData.ResultMap["newCode"] = inviteCode
	} else {
		// Undoing deletion of an invite; use the old code
		inviteCode = data.OldCode
	}

	// Insert new row into invites table
	if !data.IsUpdate {
		hashmap := make(map[string]interface{})
		hashmap["code"] = inviteCode
		hashmap["senderId"] = u.Id
		hashmap["type"] = data.Type
		hashmap["domainId"] = data.DomainId
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("invites", hashmap)
		if _, err = statement.Exec(); err != nil {
			return pages.HandlerErrorFail("Couldn't add row to invites table", err)
		}
	}

	// Insert a row into inviteEmailPairs table for each invitee
	var inviteeEmails []string
	hashmaps := make(database.InsertMaps, 0)
	// Insert one row for each invitee email address
	for _, v := range data.Invitees {
		hashmap := make(map[string]interface{})
		hashmap["code"] = inviteCode
		hashmap["email"] = v.Email
		hashmap["claimingUserId"] = v.ClaimingUserId
		// Note: data.ClaimedAt will either be a date (e.g. when undoing invite deletion),
		//   or "" if unclaimed
		hashmap["claimedAt"] = v.ClaimedAt
		// Gather invitee email addresses in a slice for sending emails
		hashmaps = append(hashmaps, hashmap)
		inviteeEmails = append(inviteeEmails, v.Email)
	}
	statement := db.NewMultipleInsertStatement("inviteEmailPairs", hashmaps, "email", "claimingUserId", "claimedAt")
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't add row(s) to inviteEmailPairs table", err)
	}

	// If it's a new invite code, send invitees an email
	if data.OldCode == "" {
		var task tasks.SendInviteEmailTask
		task.UserId = u.Id
		task.InviteeEmails = inviteeEmails
		task.Code = inviteCode
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return pages.StatusOK(returnData)
}
