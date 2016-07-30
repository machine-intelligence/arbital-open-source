// newContentRequestJsonHandler.go handles content requests

package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// contentRequestData is the data received from the request.
type contentRequestData struct {
	PageID      string
	RequestType core.ContentRequestType
}

var contentRequestHandler = siteHandler{
	URI:         "/json/contentRequest/",
	HandlerFunc: contentRequestJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func contentRequestJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data contentRequestData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}
	if !core.IsContentRequestTypeValid(data.RequestType) {
		return pages.Fail(fmt.Sprintf("Invalid content request type: %v", data.RequestType), nil).Status(http.StatusBadRequest)
	}

	// Add the request.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return plusOneToContentRequest(tx, u, data.PageID, data.RequestType)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

// Add a like to the content request for the given (page, type) pair.
func plusOneToContentRequest(tx *database.Tx, u *core.CurrentUser, pageID string, requestType core.ContentRequestType) sessions.Error {
	// Check to see if there's already a content request for this (page, type) pair.
	alreadyExists, id, err := _lookupContentRequest(tx.DB, u, pageID, requestType)
	if err != nil {
		return sessions.NewError("Error querying for an existing content request", err)
	}

	// If a content request doesn't exist, create a new one.
	if !alreadyExists {
		idInt64, serr := _createContentRequest(tx, u, pageID, requestType)
		if serr != nil {
			return sessions.NewError("Couldn't create a new content request", serr)
		}
		id = strconv.FormatInt(idInt64, 10)
	}

	// Finally, add a like for the request.
	return addNewLike(tx, u, 0, id, core.ContentRequestLikeableType, 1)
}

// Find the id of the content request for the given (page, type) pair.
func _lookupContentRequest(db *database.DB, u *core.CurrentUser, pageID string, requestType core.ContentRequestType) (bool, string, error) {
	var id string

	row := database.NewQuery(`
		SELECT id
		FROM contentRequests AS er
		WHERE er.pageId=?`, pageID).Add(`
			AND er.type=?`, string(requestType)).ToStatement(db).QueryRow()
	exists, err := row.Scan(&id)
	if err != nil {
		return false, "", fmt.Errorf("failed to scan a content request id: %v", err)
	}

	return exists, id, nil
}

// Insert a new content request row into the table.
func _createContentRequest(tx *database.Tx, u *core.CurrentUser, pageID string, requestType core.ContentRequestType) (int64, error) {
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageID
	hashmap["type"] = string(requestType)
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("contentRequests", hashmap)
	result, err := statement.WithTx(tx).Exec()
	if err != nil {
		return 0, fmt.Errorf("Couldn't insert new content request", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Couldn't retrieve id of the new content request", err)
	}
	return id, nil
}
