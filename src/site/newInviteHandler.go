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
	DomainID string `json:"domainId"`
	ToEmail  string `json:"toEmail"`
	Role     string `json:"role"`
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
	if !core.IsIntIDValid(data.DomainID) {
		return pages.Fail("DomainIds is invalid", nil).Status(http.StatusBadRequest)
	}
	if data.ToEmail == "" {
		return pages.Fail("No invite email given", nil).Status(http.StatusBadRequest)
	}

	// Check to make sure user has permissions for all the domains
	if !core.CanCurrentUserGiveRole(u, data.DomainID, data.Role) {
		return pages.Fail("Don't have permissions to give this role in this domain", nil)
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
	} else if inviteeUserID != "" {
		return pages.Fail("This user is already on Arbital. Invite them direction via 'Add Member'", err).Status(http.StatusBadRequest)
	}

	invite := &core.Invite{
		FromUserID: u.ID,
		DomainID:   data.DomainID,
		Role:       data.Role,
		ToEmail:    data.ToEmail,
		ToUserID:   inviteeUserID,
		CreatedAt:  database.Now(),
	}
	returnData.ResultMap["newInvite"] = invite

	// Check if this invite already exists
	wherePart := database.NewQuery(`WHERE fromUserId=?`, u.ID).Add(`
		AND domainId=?`, data.DomainID).Add(`
		AND toEmail=?`, data.ToEmail)
	existingInvites, err := core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return pages.Fail("Couldn't load sent invites", err)
	} else if len(existingInvites) > 0 {
		return pages.Fail("You already sent this invite.", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {

		// Create new invite
		hashmap := make(map[string]interface{})
		hashmap["fromUserId"] = invite.FromUserID
		hashmap["domainId"] = invite.DomainID
		hashmap["role"] = invite.Role
		hashmap["toEmail"] = invite.ToEmail
		hashmap["createdAt"] = invite.CreatedAt
		hashmap["toUserId"] = invite.ToUserID
		hashmap["claimedAt"] = invite.ClaimedAt
		statement := db.NewInsertStatement("invites", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't add row to invites table", err)
		}

		// If the user doesn't exist, send them an invite
		var task tasks.SendInviteTask
		task.FromUserID = invite.FromUserID
		task.ToEmail = invite.ToEmail
		task.DomainID = invite.DomainID
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(returnData)
}
