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
	Type     string          `json:"type"`
	DomainId string          `json:"domainId"`
	OldCode  string          `json:"oldCode"`
	Invitees []*core.Invitee `json:"invitees"`
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

	// If the user is modifying an existing code, load the data for it
	inviteMap := make(map[string]*core.Invite)
	existingInvitees := make(map[string]*core.Invitee) // email -> Invitee
	if data.OldCode != "" {
		wherePart := database.NewQuery(`WHERE i.code=?`, data.OldCode)
		inviteMap, err = core.LoadInvitesWhere(db, wherePart)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't load existing invites", err)
		}

		existingInvite, ok := inviteMap[data.OldCode]
		if !ok {
			return pages.HandlerBadRequestFail("Couldn't find this invite code", nil)
		}
		if existingInvite.SenderId != u.Id {
			return pages.HandlerBadRequestFail("Can't modify invite you didn't create", nil)
		}
		data.DomainId = existingInvite.DomainId

		// Check if this user was already invited
		for _, invitee := range existingInvite.Invitees {
			existingInvitees[invitee.Email] = invitee
		}
	}

	var inviteCode string
	if data.OldCode == "" {
		// Generate a new, unique invite code
		inviteCode, err = GenerateUniqueInviteCode(db)
		if err != nil {
			return pages.HandlerErrorFail("Error generating new, unique code", err)
		}
		returnData.ResultMap["newCode"] = inviteCode
	} else {
		// Modifying an invite or undoing a delete
		inviteCode = data.OldCode
	}

	// Insert new row into invites table
	// Insert an invite, even if it already exists (in case we are undoing a delete)
	hashmap := make(map[string]interface{})
	hashmap["code"] = inviteCode
	hashmap["senderId"] = u.Id
	hashmap["type"] = data.Type
	hashmap["domainId"] = data.DomainId
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("invites", hashmap, "code")
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't add row to invites table", err)
	}

	// Insert a row into inviteEmailPairs table for each invitee
	hashmaps := make(database.InsertMaps, 0)
	// Insert one row for each invitee email address
	for _, invitee := range data.Invitees {
		// Check if this user was already invited
		if _, ok := existingInvitees[invitee.Email]; ok {
			continue
		}
		hashmap := make(map[string]interface{})
		hashmap["code"] = inviteCode
		hashmap["email"] = invitee.Email
		hashmap["claimingUserId"] = invitee.ClaimingUserId
		hashmap["claimedAt"] = database.Now()
		hashmaps = append(hashmaps, hashmap)
	}
	if len(hashmaps) > 0 {
		statement = db.NewMultipleInsertStatement("inviteEmailPairs", hashmaps)
		if _, err = statement.Exec(); err != nil {
			return pages.HandlerErrorFail("Couldn't add row(s) to inviteEmailPairs table", err)
		}
	}

	// Send out email invites to new invitees
	for _, invitee := range data.Invitees {
		// Check if this user was already invited
		if _, ok := existingInvitees[invitee.Email]; ok {
			continue
		}
		var task tasks.ProcessInviteTask
		task.UserId = u.Id
		task.EmailTo = invitee.Email
		task.Code = inviteCode
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}
	return pages.StatusOK(returnData)
}
