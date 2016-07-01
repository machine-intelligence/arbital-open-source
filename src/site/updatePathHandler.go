// updatePathHandler.go updates the given path instance
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

var updatePathHandler = siteHandler{
	URI:         "/json/updatePath/",
	HandlerFunc: updatePathHandlerFunc,
}

type updatePathData struct {
	Id       string
	Progress int
}

func updatePathHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	// Decode data
	var data updatePathData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Load the path instance
	instance, err := core.LoadPathInstance(db, data.Id, u)
	if err != nil {
		return pages.Fail("Couldn't load the path instance: %v", err)
	} else if instance == nil {
		return pages.Fail("Couldn't find the path instance", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update the path
		hashmap := make(database.InsertMap)
		hashmap["id"] = data.Id
		hashmap["userId"] = u.Id
		hashmap["progress"] = data.Progress
		if data.Progress >= len(instance.PageIds)-1 {
			hashmap["isFinished"] = true
		}
		hashmap["updatedAt"] = database.Now()
		statement := db.NewInsertStatement("pathInstances", hashmap, hashmap.GetKeys()...).WithTx(tx)
		_, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert pathInstance", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
