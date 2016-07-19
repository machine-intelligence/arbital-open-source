// resolveMarkHandler.go resolves an existing mark.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// resolveMarkData contains data given to us in the request.
type resolveMarkData struct {
	MarkID string

	// Resolve the mark, and this is the page id with the "answer"
	// If "", it means the mark was dismissed
	ResolvedPageID string
	// Text (optional, can only be set by the owner of the mark)
	Text string
}

var resolveMarkHandler = siteHandler{
	URI:         "/resolveMark/",
	HandlerFunc: resolveMarkHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// resolveMarkHandlerFunc handles requests to create/resolve a prior like.
func resolveMarkHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	loadData := core.NewHandlerData(u)

	var data resolveMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Load the mark
	mark := &core.Mark{ID: data.MarkID}
	loadData.MarkMap[data.MarkID] = mark
	err = core.LoadMarkData(db, loadData.PageMap, loadData.UserMap, loadData.MarkMap, u)
	if err != nil {
		return pages.Fail("Couldn't load the mark", err)
	} else if mark.Type == "" {
		return pages.Fail("No such mark", nil).Status(http.StatusBadRequest)
	}

	if mark.CreatorID != u.ID {
		data.Text = ""
	}

	// Update existing mark
	hashmap := make(database.InsertMap)
	hashmap["id"] = data.MarkID
	hashmap["resolvedPageId"] = data.ResolvedPageID
	hashmap["resolvedBy"] = u.ID
	hashmap["resolvedAt"] = database.Now()
	if data.Text != "" {
		hashmap["text"] = data.Text
	}
	if mark.CreatorID == u.ID {
		hashmap["isSubmitted"] = true
	}
	statement := db.NewInsertStatement("marks", hashmap, hashmap.GetKeys()...)
	_, err = statement.Exec()
	if err != nil {
		return pages.Fail("Couldn't update the mark", err)
	}

	// If the mark was resolved for the first time, update the user mark owner
	if mark.Type != core.QueryMarkType && mark.ResolvedBy == "" && data.ResolvedPageID != "" {
		hashmap := make(database.InsertMap)
		hashmap["userId"] = mark.CreatorID
		hashmap["type"] = core.ResolvedMarkUpdateType
		hashmap["goToPageId"] = data.ResolvedPageID
		hashmap["markId"] = data.MarkID
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("updates", hashmap)
		_, err = statement.Exec()
		if err != nil {
			return pages.Fail("Couldn't insert update", err)
		}
	}

	return pages.Success(nil)
}
