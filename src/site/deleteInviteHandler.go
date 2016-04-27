// deleteInviteHandler.go deletes invites
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// deleteInviteData contains data given to us in the request.
type deleteInviteData struct {
	Code string
}

var deleteInviteHandler = siteHandler{
	URI:         "/deleteInvite/",
	HandlerFunc: deleteInviteHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin:   true,
		RequireTrusted: true,
	},
}

// deleteInviteHandlerFunc handles deleting invites
func deleteInviteHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U, false)

	var data deleteInviteData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	// Attempt to delete invite from DB and return number of rows affected
	statement := db.NewStatement(`
		DELETE FROM invites, inviteEmailPairs
		USING invites
		INNER JOIN inviteEmailPairs
		WHERE invites.code = inviteEmailPairs.code AND invites.code=?`)
	if result, err := statement.Exec(data.Code); err != nil {
		return pages.HandlerErrorFail("Couldn't delete invite code", err)
	} else {
		returnData.ResultMap["deletionStatus"], _ = result.RowsAffected()
	}

	return pages.StatusOK(returnData)
}
