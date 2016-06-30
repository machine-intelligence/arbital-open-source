// newPathPageHandler.go adds a page to a path
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// newPathPageData contains the data we get in the request
type newPathPageData struct {
	GuideId    string
	PathPageId string
}

var newPathPageHandler = siteHandler{
	URI:         "/json/newPathPage/",
	HandlerFunc: newPathPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func newPathPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data newPathPageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Check permissions
	pageIds := []string{data.GuideId}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Begin the transaction.
	var id int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Compute the path page index
		pathIndex := 0
		row := database.NewQuery(`
			SELECT IFNULL(MAX(pathIndex)+1,0)
			FROM pathPages
			WHERE guideId=?`, data.GuideId).ToTxStatement(tx).QueryRow()
		_, err := row.Scan(&pathIndex)
		if err != nil {
			return sessions.NewError("Couldn't load pathIndex", err)
		}

		// Create the path page
		hashmap := make(database.InsertMap)
		hashmap["guideId"] = data.GuideId
		hashmap["pathPageId"] = data.PathPageId
		hashmap["pathIndex"] = pathIndex
		hashmap["createdBy"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["updatedBy"] = u.Id
		hashmap["updatedAt"] = database.Now()
		statement := db.NewInsertStatement("pathPages", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't update pathPages", err)
		}

		id, err = result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get pathPage id", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Load the newly created pathPage
	returnData.ResultMap["pathPage"], err = core.LoadPathPage(db, fmt.Sprintf("%d", id))
	if err != nil {
		return pages.Fail("Couldn't load the pathPage: %v", err)
	}
	return pages.Success(returnData)
}