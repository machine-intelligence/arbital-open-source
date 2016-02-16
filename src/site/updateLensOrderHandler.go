// updateLensOrderHandler.go handles reordering of lenses
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// updateLensOrderData contains the data we get in the request
type updateLensOrderData struct {
	// Id of the page the lenses are for
	PageId string
	// Lens id -> order index map
	OrderMap map[string]int
}

var updateLensOrderHandler = siteHandler{
	URI:         "/updateLensOrder/",
	HandlerFunc: updateLensOrderHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

func updateLensOrderHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updateLensOrderData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Page id isn't specified", err)
	}
	if len(data.OrderMap) <= 0 {
		return pages.StatusOK(nil)
	}

	// Computed which pages count as visited.
	lensIndexValues := make([]interface{}, 0)
	for pageId, index := range data.OrderMap {
		lensIndexValues = append(lensIndexValues, pageId, index)
	}

	// Add a visit to pages for which we loaded text.
	if len(lensIndexValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO pageInfos (pageId, lensIndex)
			VALUES ` + database.ArgsPlaceholder(len(lensIndexValues), 2) + `
			ON DUPLICATE KEY UPDATE lensIndex=VALUES(lensIndex)`)
		if _, err = statement.Exec(lensIndexValues...); err != nil {
			return pages.HandlerErrorFail("Couldn't update visits", err)
		}
	}

	// Generate updates for users who are subscribed to the primary page
	var task tasks.NewUpdateTask
	task.UpdateType = core.PageInfoEditUpdateType
	task.UserId = u.Id
	task.GroupByPageId = data.PageId
	task.SubscribedToId = data.PageId
	task.GoToPageId = data.PageId
	if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.StatusOK(nil)
}
