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
	DomainIds []string `json:"domainIds"`
	ToEmail   string   `json:"toEmail"`
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
	for _, domainId := range data.DomainIds {
		if !core.IsIdValid(domainId) {
			return pages.HandlerBadRequestFail("One of the domainIds is invalid", nil)
		}
	}
	if data.ToEmail == "" {
		return pages.HandlerBadRequestFail("No invite email given", nil)
	}

	// Always send an invite to the general domain
	data.DomainIds = append(data.DomainIds, "")

	// Check to see if the invitee is already a user in our DB
	var inviteeUserId string
	row := db.NewStatement(`
		SELECT id
		FROM users
		WHERE email=?`).QueryRow(data.ToEmail)
	_, err = row.Scan(&inviteeUserId)
	if err != nil {
		return pages.Fail("Couldn't retrieve a user", err)
	}

	// Create invite map
	inviteMap := make(map[string]*core.Invite) // key: domainId
	for _, domainId := range data.DomainIds {
		inviteMap[domainId] = &core.Invite{
			FromUserId: u.Id,
			DomainId:   domainId,
			ToEmail:    data.ToEmail,
			ToUserId:   inviteeUserId,
			CreatedAt:  database.Now(),
		}
	}
	returnData.ResultMap["inviteMap"] = inviteMap

	// Check if this invite already exists
	wherePart := database.NewQuery(`WHERE fromUserId=?`, u.Id).Add(`
		AND domainId IN`).AddArgsGroupStr(data.DomainIds).Add(`
		AND toEmail=?`, data.ToEmail)
	existingInvites, err := core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return pages.Fail("Couldn't load sent invites", err)
	}
	for _, existingInvite := range existingInvites {
		delete(inviteMap, existingInvite.DomainId)
	}
	if len(inviteMap) <= 0 {
		return pages.Success(returnData)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		inviteDomainIds := make([]string, 0)
		for domainId, invite := range inviteMap {
			if domainId != "" {
				inviteDomainIds = append(inviteDomainIds, domainId)
			}

			// Create new invite
			hashmap := make(map[string]interface{})
			hashmap["fromUserId"] = u.Id
			hashmap["domainId"] = domainId
			hashmap["toEmail"] = data.ToEmail
			hashmap["createdAt"] = database.Now()
			if inviteeUserId != "" {
				hashmap["toUserId"] = inviteeUserId
				hashmap["claimedAt"] = database.Now()
				invite.ClaimedAt = database.Now()
			}
			statement := db.NewInsertStatement("invites", hashmap).WithTx(tx)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't add row to invites table", err
			}

			// If the user already exists, send them an update
			if inviteeUserId != "" {
				hashmap := make(map[string]interface{})
				hashmap["userId"] = invite.ToUserId
				hashmap["type"] = core.InviteReceivedUpdateType
				hashmap["createdAt"] = database.Now()
				hashmap["groupByUserId"] = u.Id
				hashmap["subscribedToId"] = u.Id
				hashmap["goToPageId"] = domainId
				hashmap["byUserId"] = u.Id
				hashmap["unseen"] = true
				statement := db.NewInsertStatement("updates", hashmap).WithTx(tx)
				if _, err = statement.Exec(); err != nil {
					return "Couldn't add a new update for the invitee", err
				}
			}
		}

		// If the user doesn't exist, send them an invite
		if inviteeUserId == "" {
			var task tasks.SendInviteTask
			task.FromUserId = u.Id
			task.ToEmail = data.ToEmail
			task.DomainIds = inviteDomainIds
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				return "Couldn't enqueue a task", err
			}
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.Fail(errMessage, err)
	}

	return pages.Success(returnData)
}
