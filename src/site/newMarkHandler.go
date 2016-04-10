// newMarkHandler.go creates a new mark.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newMarkData contains data given to us in the request.
type newMarkData struct {
	PageId        string
	Edit          int
	AnchorContext string
	AnchorText    string
	AnchorOffset  int
}

var newMarkHandler = siteHandler{
	URI:         "/newMark/",
	HandlerFunc: newMarkHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newMarkHandlerFunc handles requests to create/update a prior like.
func newMarkHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(params.U, true)

	var data newMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}
	if data.AnchorContext == "" {
		return pages.HandlerBadRequestFail("No anchor context is set", nil)
	}

	// Load what requirements the user has met
	masteryMap := make(map[string]*core.Mastery)
	err = core.LoadMasteries(db, u, masteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Load masteries failed: %v", err)
	}

	now := database.Now()
	var lastInsertId int64

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Compute snapshot id we can use
		var requisiteSnapshotId int64
		row := tx.NewTxStatement(`
			SELECT IFNULL(max(id),0)
			FROM userRequisitePairSnapshots
			`).QueryRow()
		_, err = row.Scan(&requisiteSnapshotId)
		if err != nil {
			return "Couldn't load max snapshot id", err
		}
		requisiteSnapshotId++

		// Create a new mark
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = data.Edit
		hashmap["creatorId"] = u.Id
		hashmap["createdAt"] = now
		hashmap["requisiteSnapshotId"] = requisiteSnapshotId
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		statement := tx.NewInsertTxStatement("marks", hashmap)
		resp, err := statement.Exec()
		if err != nil {
			return "Couldn't insert an new mark", err
		}

		lastInsertId, err = resp.LastInsertId()
		if err != nil {
			return "Couldn't get inserted id", err
		}

		snapshotValues := make([]interface{}, 0)
		for _, req := range masteryMap {
			if req.Has || req.Wants {
				snapshotValues = append(snapshotValues, requisiteSnapshotId, u.Id, req.PageId, now, req.Has, req.Wants)
			}
		}

		statement = tx.NewTxStatement(`
				INSERT INTO userRequisitePairSnapshots (id,userId,requisiteId,createdAt,has,wants)
				VALUES ` + database.ArgsPlaceholder(len(snapshotValues), 6))
		if _, err := statement.Exec(snapshotValues...); err != nil {
			return "Couldn't insert into userRequisitePairSnapshots", err
		}

		return "", nil
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	returnData.ResultMap["markId"] = fmt.Sprintf("%d", lastInsertId)

	return pages.StatusOK(returnData.ToJson())
}
