// newInviteHandler.go adds new invites to db and auto-claims / sends invite emails

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
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
		RequireLogin: true,
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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	for _, domainID := range data.DomainIds {
		if !core.IsIDValid(domainID) {
			return pages.Fail("One of the domainIds is invalid", nil).Status(http.StatusBadRequest)
		}
	}
	if data.ToEmail == "" {
		return pages.Fail("No invite email given", nil).Status(http.StatusBadRequest)
	}

	// Check to make sure user has permissions for all the domains
	for _, domainID := range data.DomainIds {
		if u.TrustMap[domainID].Level < core.ArbiterTrustLevel {
			return pages.Fail("Don't have permissions to invite to one of the domains", nil).Status(http.StatusBadRequest)
		}
	}

	// Check to see if the invitee is already a user in our DB
	var inviteeUserID string
	row := db.NewStatement(`
		SELECT id
		FROM users
		WHERE email=?`).QueryRow(data.ToEmail)
	_, err = row.Scan(&inviteeUserID)
	if err != nil {
		return pages.Fail("Couldn't retrieve a user", err)
	}

	// Create invite map
	inviteMap := make(map[string]*core.Invite) // key: domainId
	for _, domainID := range data.DomainIds {
		inviteMap[domainID] = &core.Invite{
			FromUserID: u.ID,
			DomainID:   domainID,
			ToEmail:    data.ToEmail,
			ToUserID:   inviteeUserID,
			CreatedAt:  database.Now(),
		}
	}
	returnData.ResultMap["inviteMap"] = inviteMap

	// Check if this invite already exists
	wherePart := database.NewQuery(`WHERE fromUserId=?`, u.ID).Add(`
		AND domainId IN`).AddArgsGroupStr(data.DomainIds).Add(`
		AND toEmail=?`, data.ToEmail)
	existingInvites, err := core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return pages.Fail("Couldn't load sent invites", err)
	}
	for _, existingInvite := range existingInvites {
		delete(inviteMap, existingInvite.DomainID)
	}
	if len(inviteMap) <= 0 {
		return pages.Success(returnData)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {

		inviteDomainIds := make([]string, 0)
		for domainID, invite := range inviteMap {
			if domainID != "" {
				inviteDomainIds = append(inviteDomainIds, domainID)
			}

			// Create new invite
			hashmap := make(map[string]interface{})
			hashmap["fromUserId"] = u.ID
			hashmap["domainId"] = domainID
			hashmap["toEmail"] = data.ToEmail
			hashmap["createdAt"] = database.Now()
			if inviteeUserID != "" {
				hashmap["toUserId"] = inviteeUserID
				hashmap["claimedAt"] = database.Now()
				invite.ClaimedAt = database.Now()
			}
			statement := db.NewInsertStatement("invites", hashmap).WithTx(tx)
			if _, err = statement.Exec(); err != nil {
				return sessions.NewError("Couldn't add row to invites table", err)
			}

			// If the user already exists, send them an update
			if inviteeUserID != "" {
				hashmap := make(map[string]interface{})
				hashmap["userId"] = invite.ToUserID
				hashmap["type"] = core.InviteReceivedUpdateType
				hashmap["createdAt"] = database.Now()
				hashmap["subscribedToId"] = u.ID
				hashmap["goToPageId"] = domainID
				hashmap["byUserId"] = u.ID
				statement := db.NewInsertStatement("updates", hashmap).WithTx(tx)
				if _, err = statement.Exec(); err != nil {
					return sessions.NewError("Couldn't add a new update for the invitee", err)
				}

				// Create/update user trust
				hashmap = make(map[string]interface{})
				hashmap["userId"] = inviteeUserID
				hashmap["domainId"] = domainID
				hashmap["editTrust"] = core.BasicKarmaLevel
				statement = db.NewInsertStatement("userTrust", hashmap, "editTrust")
				if _, err := statement.WithTx(tx).Exec(); err != nil {
					return sessions.NewError("Couldn't update/create userTrust row", err)
				}
			}
		}

		// If the user doesn't exist, send them an invite
		if inviteeUserID == "" {
			var task tasks.SendInviteTask
			task.FromUserID = u.ID
			task.ToEmail = data.ToEmail
			task.DomainIds = inviteDomainIds
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				return sessions.NewError("Couldn't enqueue a task", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(returnData)
}
