// newExplanationRequestJsonHandler.go handles explantion requests
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

// explanationRequestData is the data received from the request.
type explanationRequestData struct {
	PageID string
	Type   string
}

var explanationRequestHandler = siteHandler{
	URI:         "/json/explanationRequest/",
	HandlerFunc: explanationRequestJSONHandler,
	Options:     pages.PageOptions{},
}

func explanationRequestJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data explanationRequestData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Add the request.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return plusOneToExplanationRequest(tx, u, data.PageID, data.Type)
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

// Add a like to the explanation request for the given (page, type) pair.
func plusOneToExplanationRequest(tx *database.Tx, u *core.CurrentUser, pageID string, requestType string) sessions.Error {
	// Check to see if there's already an explanation request for this (page, type) pair.
	alreadyExists, id, err := _lookupExplanationRequest(tx.DB, u, pageID, requestType)
	if err != nil {
		return sessions.NewError("Error querying for an existing explanation request", err)
	}

	// If an explanation request doesn't exist, create a new one.
	if !alreadyExists {
		idInt64, serr := _createExplanationRequest(tx, u, pageID, requestType)
		if serr != nil {
			return sessions.NewError("Couldn't create a new explanation request", serr)
		}
		id = strconv.FormatInt(idInt64, 10)
	}

	// Finally, add a like for the request.
	return addNewLike(tx, u, 0, id, core.ContentRequestLikeableType, 1)
}

// Find the id of the explanation request for the given (page, type) pair.
func _lookupExplanationRequest(db *database.DB, u *core.CurrentUser, pageID string, requestType string) (bool, string, error) {
	var id string

	row := database.NewQuery(`
		SELECT id
		FROM contentRequests AS er
		WHERE er.pageId=?`, pageID).Add(`
			AND er.type=?`, requestType).ToStatement(db).QueryRow()
	exists, err := row.Scan(&id)
	if err != nil {
		return false, "", fmt.Errorf("failed to scan an explanation request id: %v", err)
	}

	return exists, id, nil
}

// Insert a new explanation request row into the table.
func _createExplanationRequest(tx *database.Tx, u *core.CurrentUser, pageID string, requestType string) (int64, error) {
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageID
	hashmap["type"] = requestType
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("contentRequests", hashmap)
	result, err := statement.WithTx(tx).Exec()
	if err != nil {
		return 0, fmt.Errorf("Couldn't insert new explanation request", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Couldn't retrieve id of the new explanation request", err)
	}
	return id, nil
}
