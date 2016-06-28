// newMarkHandler.go creates a new mark.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

const (
	markAutoProcessDelay = 5 * 60 // seconds
)

// newMarkData contains data given to us in the request.
type newMarkData struct {
	PageId        string
	Type          string
	Edit          int
	Text          string
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
	returnData := core.NewHandlerData(u)
	now := database.Now()

	var data newMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	}
	if data.Type != core.QueryMarkType && data.Type != core.TypoMarkType && data.Type != core.ConfusionMarkType {
		return pages.Fail("Invalid mark type", nil).Status(http.StatusBadRequest)
	}
	if data.Type != core.QueryMarkType && data.AnchorContext == "" {
		return pages.Fail("No anchor context is set", nil).Status(http.StatusBadRequest)
	}

	// Load what requirements the user has met
	masteryMap := make(map[string]*core.Mastery)
	err = core.LoadMasteries(db, u, masteryMap)
	if err != nil {
		return pages.Fail("Load masteries failed: %v", err)
	}

	var markId int64
	// Set to true if this mark is automatically processed within a few minutes
	autoProcessed := data.Type == core.TypoMarkType || data.Type == core.ConfusionMarkType

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Compute snapshot id we can use
		var requisiteSnapshotId int64
		if data.Type != core.TypoMarkType {
			row := tx.DB.NewStatement(`
			SELECT IFNULL(max(id),0)
			FROM userRequisitePairSnapshots`).WithTx(tx).QueryRow()
			_, err = row.Scan(&requisiteSnapshotId)
			if err != nil {
				return sessions.NewError("Couldn't load max snapshot id", err)
			}
			requisiteSnapshotId++
		}

		// Create a new mark
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["type"] = data.Type
		hashmap["edit"] = data.Edit
		hashmap["text"] = data.Text
		hashmap["creatorId"] = u.Id
		hashmap["createdAt"] = now
		hashmap["anchorContext"] = data.AnchorContext
		hashmap["anchorText"] = data.AnchorText
		hashmap["anchorOffset"] = data.AnchorOffset
		hashmap["requisiteSnapshotId"] = requisiteSnapshotId
		hashmap["isSubmitted"] = autoProcessed
		statement := tx.DB.NewInsertStatement("marks", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert an new mark", err)
		}
		markId, err = resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get inserted id", err)
		}

		// Snapshot user's requisites
		if data.Type != core.TypoMarkType {
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
					return sessions.NewError("Couldn't insert into userRequisitePairSnapshots", err)
				}
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}
	markIdStr := fmt.Sprintf("%d", markId)

	// Enqueue a task that will create relevant updates for this mark event
	if autoProcessed {
		err = EnqueueNewMarkUpdateTask(params, markIdStr, data.PageId, markAutoProcessDelay)
		if err != nil {
			return pages.Fail("Couldn't enqueue an updateTask", err)
		}
	}

	// Load mark to return it
	core.AddMarkToMap(markIdStr, returnData.MarkMap)
	core.AddPageToMap("370", returnData.PageMap, core.TitlePlusLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["markId"] = markIdStr

	return pages.Success(returnData)
}

func EnqueueNewMarkUpdateTask(params *pages.HandlerParams, markId string, pageId string, delay int) error {
	var updateTask tasks.NewUpdateTask
	updateTask.UserId = params.U.Id
	updateTask.GoToPageId = pageId
	updateTask.SubscribedToId = pageId
	updateTask.UpdateType = core.NewMarkUpdateType
	updateTask.GroupByPageId = pageId
	updateTask.MarkId = markId
	options := &tasks.TaskOptions{Delay: delay}
	if err := tasks.Enqueue(params.C, &updateTask, options); err != nil {
		return fmt.Errorf("Couldn't enqueue an updateTask: %v", err)
	}
	return nil
}
