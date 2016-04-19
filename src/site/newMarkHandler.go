// newMarkHandler.go creates a new mark.
package site

import (
	"encoding/json"
	"fmt"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

const (
	processMarkDelay = time.Duration(5*60) * time.Second
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
	c := params.C
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(params.U, false)

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
	var markId int64

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Compute snapshot id we can use
		var requisiteSnapshotId int64
		row := tx.DB.NewStatement(`
			SELECT IFNULL(max(id),0)
			FROM userRequisitePairSnapshots
			`).WithTx(tx).QueryRow()
		_, err = row.Scan(&requisiteSnapshotId)
		if err != nil {
			return "Couldn't load max snapshot id", err
		}
		requisiteSnapshotId++

		// Create a new mark
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["edit"] = data.Edit
		hashmap["creatorId"] = u.Id
		hashmap["createdAt"] = now
		hashmap["requisiteSnapshotId"] = requisiteSnapshotId
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		statement := tx.DB.NewInsertStatement("marks", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return "Couldn't insert an new mark", err
		}
		markId, err = resp.LastInsertId()
		if err != nil {
			return "Couldn't get inserted id", err
		}

		// Snapshot user's requisites
		hashmaps := make(database.InsertMaps, 0)
		for _, req := range masteryMap {
			if req.Has || req.Wants {
				hashmap := make(database.InsertMap)
				hashmap["id"] = requisiteSnapshotId
				hashmap["userId"] = u.Id
				hashmap["requisiteId"] = req.PageId
				hashmap["has"] = req.Has
				hashmap["wants"] = req.Wants
				hashmap["createdAt"] = now
				hashmaps = append(hashmaps, hashmap)
			}
		}
		if len(hashmaps) > 0 {
			statement = tx.DB.NewMultipleInsertStatement("userRequisitePairSnapshots", hashmaps)
			if _, err := statement.WithTx(tx).Exec(); err != nil {
				return "Couldn't insert into userRequisitePairSnapshots", err
			}
		}
		return "", nil
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Enqueue a task that will create relevant updates for this mark event
	var task tasks.ProcessMarkTask
	task.Id = markId
	options := &tasks.TaskOptions{Delay: processMarkDelay}
	if !sessions.Live {
		options.Delay = 0
	}
	if err := tasks.Enqueue(c, &task, options); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	// Load mark to return it
	markIdStr := fmt.Sprintf("%d", markId)
	returnData.AddMark(markIdStr)
	core.AddPageToMap("370", returnData.PageMap, core.TitlePlusLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["markId"] = markIdStr

	return pages.StatusOK(returnData)
}