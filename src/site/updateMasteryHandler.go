// updateMastery.go handles request to add/delete mastery
package site

import (
	"encoding/json"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updateMasteryData contains the data we get in the request.
type updateMasteryData struct {
	MasteryId int64 `json:",string"`
	Has       bool
}

var updateMasteryHandler = siteHandler{
	URI:         "/updateMastery/",
	HandlerFunc: updateMasteryHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updateMasteryHandlerFunc handles requests for adding a new subscription.
func updateMasteryHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updateMasteryData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.MasteryId == 0 {
		return pages.HandlerBadRequestFail("Mastery id has to be set", err)
	}

	hashmap := make(map[string]interface{})
	hashmap["masteryId"] = data.MasteryId
	hashmap["userId"] = u.Id
	hashmap["has"] = data.Has
	hashmap["isManuallySet"] = true
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	statement := db.NewInsertStatement("userMasteryPairs", hashmap, "has", "isManuallySet", "updatedAt")
	_, err = statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't create new subscription", err)
	}
	return pages.StatusOK(nil)
}
