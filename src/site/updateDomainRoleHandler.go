// Handles requests to update a user's role in a domain

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type updateDomainRoleData struct {
	UserID   string `json:"userID"`
	DomainID string `json:"domainID"`
	Role     string `json:role"`
}

var updateDomainRoleHandler = siteHandler{
	URI:         "/updateDomainRole/",
	HandlerFunc: updateDomainRoleHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateDomainRoleHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data updateDomainRoleData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.UserID) {
		return pages.Fail("ToId is incorrect", nil).Status(http.StatusBadRequest)
	}
	if !core.IsIntIDValid(data.DomainID) {
		return pages.Fail("DomainID is incorrect", nil).Status(http.StatusBadRequest)
	}
	if !core.IsDomainRoleValid(data.Role) {
		return pages.Fail("Role is incorrect", nil).Status(http.StatusBadRequest)
	}

	if !core.CanCurrentUserGiveRole(u, data.DomainID, data.Role) {
		return pages.Fail("Don't have permissions to give this role in this domain", nil)
	}

	hashmap := make(map[string]interface{})
	hashmap["domainId"] = data.DomainID
	hashmap["userId"] = data.UserID
	hashmap["createdAt"] = database.Now()
	hashmap["role"] = data.Role
	statement := db.NewInsertStatement("domainMembers", hashmap, "role")
	if _, err := statement.Exec(); err != nil {
		return pages.Fail("Couldn't update domainMembers", err)
	}

	return pages.Success(nil)
}
