// Handles requests to update a user's role in a domain

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type updateDomainData struct {
	DomainID               string `json:"domainID"`
	CanUsersProposeComment bool   `json:canUsersProposeComment"`
}

var updateDomainHandler = siteHandler{
	URI:         "/updateDomain/",
	HandlerFunc: updateDomainHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateDomainHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data updateDomainData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIntIDValid(data.DomainID) {
		return pages.Fail("DomainID is incorrect", nil).Status(http.StatusBadRequest)
	}

	if !core.RoleAtLeast(u.GetDomainMembershipRole(data.DomainID), core.ArbitratorDomainRole) {
		return pages.Fail("Don't have permissions to change settings of this domain", nil)
	}

	hashmap := make(map[string]interface{})
	hashmap["id"] = data.DomainID
	hashmap["canUsersProposeComment"] = data.CanUsersProposeComment
	statement := db.NewInsertStatement("domains", hashmap, "canUsersProposeComment")
	if _, err := statement.Exec(); err != nil {
		return pages.Fail("Couldn't update domains", err)
	}

	return pages.Success(nil)
}
