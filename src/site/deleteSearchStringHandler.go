// deleteSearchStringHandler.go adds a search string to a page
package site

import (
	"encoding/json"
	"net/http"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// deleteSearchStringData contains data given to us in the request.
type deleteSearchStringData struct {
	Id string
}

var deleteSearchStringHandler = siteHandler{
	URI:         "/deleteSearchString/",
	HandlerFunc: deleteSearchStringHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// deleteSearchStringHandlerFunc handles requests to create/update a like.
func deleteSearchStringHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	c := params.C
	db := params.DB

	var data deleteSearchStringData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(data.Id, 10, 64)
	if err != nil {
		return pages.Fail("Invalid id", err).Status(http.StatusBadRequest)
	}

	searchString, err := core.LoadSearchString(db, data.Id)
	if err != nil {
		return pages.Fail("Couldn't load the search string", err)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Delete the search string
		statement := database.NewQuery(`
			DELETE FROM searchStrings WHERE id=?`, id).ToStatement(db).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't delete from DB", err)
		}

		// Update change logs
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = searchString.PageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.SearchStringChangeChangeLog
		hashmap["oldSettingsValue"] = searchString.Text
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't add to changeLogs", err)
		}
		changeLogId, err := resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get changeLog id", err)
		}

		// Insert updates
		var task tasks.NewUpdateTask
		task.UserId = u.Id
		task.GoToPageId = searchString.PageId
		task.SubscribedToId = searchString.PageId
		task.UpdateType = core.ChangeLogUpdateType
		task.GroupByPageId = searchString.PageId
		task.ChangeLogId = changeLogId
		task.EditorsOnly = true
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task: %v", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
