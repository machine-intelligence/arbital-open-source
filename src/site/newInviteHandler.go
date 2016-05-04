// newInviteHandler.go adds new invites to db and auto-claims / sends invite emails
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// updateSettingsData contains data given to us in the request.
type newInviteData struct {
	DomainId string `json:"domainId"`
	ToEmail  string `json:"toEmail"`
}

var newInviteHandler = siteHandler{
	URI:         "/newInvite/",
	HandlerFunc: newInviteHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin:   true,
		RequireTrusted: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func newInviteHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	u := params.U
	returnData := core.NewHandlerData(u)

	var data newInviteData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.DomainId) && data.DomainId != "" {
		return pages.HandlerBadRequestFail("DomainId is invalid", nil)
	}
	if data.ToEmail == "" {
		return pages.HandlerBadRequestFail("No invite email given", nil)
	}

	invite := &core.Invite{
		FromUserId: u.Id,
		DomainId:   data.DomainId,
		ToEmail:    data.ToEmail,
		CreatedAt:  database.Now(),
	}
	returnData.ResultMap["invite"] = invite

	// Check if this invite already exists
	wherePart := database.NewQuery(`WHERE fromUserId=? AND domainId=? AND toEmail=?`, u.Id, data.DomainId, data.ToEmail)
	invites, err := core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load sent invites", err)
	}
	if len(invites) > 0 {
		return pages.StatusOK(returnData)
	}

	// Check to see if the invitee is already a user in our DB
	row := db.NewStatement(`
		SELECT id
		FROM users
		WHERE email=?`).QueryRow(data.ToEmail)
	_, err = row.Scan(&invite.ToUserId)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't retrieve a user", err)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		// Create new invite
		hashmap := make(map[string]interface{})
		hashmap["fromUserId"] = u.Id
		hashmap["domainId"] = data.DomainId
		hashmap["toEmail"] = data.ToEmail
		hashmap["createdAt"] = database.Now()
		if invite.ToUserId != "" {
			hashmap["toUserId"] = invite.ToUserId
			hashmap["claimedAt"] = database.Now()
			invite.ClaimedAt = database.Now()
		}
		statement := db.NewInsertStatement("invites", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't add row to invites table", err
		}

		// If there is no user, invite them via email
		if invite.ToUserId == "" {
			var task tasks.ProcessInviteTask
			task.FromUserId = u.Id
			task.ToEmail = data.ToEmail
			task.DomainId = data.DomainId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				return "Couldn't enqueue a task", err
			}
		} else {
			// Add an update to the invitee
			hashmap := make(map[string]interface{})
			hashmap["userId"] = invite.ToUserId
			hashmap["type"] = core.InviteReceivedUpdateType
			hashmap["createdAt"] = database.Now()
			hashmap["groupByUserId"] = u.Id
			hashmap["subscribedToId"] = u.Id
			hashmap["goToPageId"] = data.DomainId
			hashmap["byUserId"] = u.Id
			hashmap["unseen"] = true
			statement := db.NewInsertStatement("updates", hashmap).WithTx(tx)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't add a new update for the invitee", err
			}
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	return pages.StatusOK(returnData)
}
