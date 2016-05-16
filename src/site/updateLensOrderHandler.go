// updateLensOrderHandler.go handles reordering of lenses
package site

import (
	"encoding/json"
	"net/http"

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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Page id isn't specified", err).Status(http.StatusBadRequest)
	}
	if len(data.OrderMap) <= 0 {
		return pages.Success(nil)
	}

	// Check permissions
	pageIds := []string{data.PageId}
	for pageId, _ := range data.OrderMap {
		pageIds = append(pageIds, pageId)
	}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err).Status(http.StatusForbidden)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Computed which pages count as visited.
	lensIndexValues := make([]interface{}, 0)
	for pageId, index := range data.OrderMap {
		lensIndexValues = append(lensIndexValues, pageId, index)
	}

	// Update the lens index.
	if len(lensIndexValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO pageInfos (pageId, lensIndex)
			VALUES ` + database.ArgsPlaceholder(len(lensIndexValues), 2) + `
			ON DUPLICATE KEY UPDATE lensIndex=VALUES(lensIndex)`)
		if _, err = statement.Exec(lensIndexValues...); err != nil {
			return pages.Fail("Couldn't update a lens index", err)
		}
	}

	// Generate updates for users who are subscribed to the primary page
	var task tasks.NewUpdateTask
	task.UpdateType = core.PageInfoEditUpdateType
	task.UserId = u.Id
	task.GroupByPageId = data.PageId
	task.SubscribedToId = data.PageId
	task.GoToPageId = data.PageId
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
