// startPathHandler.go starts the user on the given path
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

var startPathHandler = siteHandler{
	URI:         "/json/startPath/",
	HandlerFunc: startPathHandlerFunc,
}

type startPathData struct {
	GuideId string
}

func startPathHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data startPathData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.GuideId) {
		return pages.Fail("Invalid guideId", nil).Status(http.StatusBadRequest)
	}

	// Load path pages
	pathPageIds := []string{data.GuideId}
	queryPart := database.NewQuery(`
		WHERE pathp.guideId=?`, data.GuideId).Add(`
		ORDER BY pathp.pathIndex`)
	err = core.LoadPathPages(db, queryPart, nil, func(db *database.DB, pathPage *core.PathPage) error {
		pathPageIds = append(pathPageIds, pathPage.PathPageId)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load the path pages: %v", err)
	} else if len(pathPageIds) <= 0 {
		return pages.Fail("No path pages found for this guide", nil).Status(http.StatusBadRequest)
	}

	// Create the sourcePageIds
	sourcePageIds := make([]string, 0)
	for range pathPageIds {
		sourcePageIds = append(sourcePageIds, data.GuideId)
	}

	// Begin the transaction.
	var id int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Start the path
		hashmap := make(database.InsertMap)
		hashmap["userId"] = u.ID
		hashmap["guideId"] = data.GuideId
		hashmap["pageIds"] = strings.Join(pathPageIds, ",")
		hashmap["sourcePageIds"] = strings.Join(sourcePageIds, ",")
		hashmap["progress"] = 1
		hashmap["createdAt"] = database.Now()
		hashmap["updatedAt"] = database.Now()
		statement := db.NewInsertStatement("pathInstances", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert pathInstance", err)
		}

		id, err = result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get lens id", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	returnData.ResultMap["path"], err = core.LoadPathInstance(db, fmt.Sprintf("%d", id), u)
	if err != nil {
		return pages.Fail("Couldn't load the path: %v", err)
	}
	return pages.Success(returnData)
}
