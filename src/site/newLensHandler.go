// newLensHandler.go updates the name of a lens
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// newLensData contains the data we get in the request
type newLensData struct {
	PageId string
	LensId string
}

var newLensHandler = siteHandler{
	URI:         "/json/newLens/",
	HandlerFunc: newLensHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func newLensHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data newLensData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Check if this page is already a lens
	var lens *core.Lens
	queryPart := database.NewQuery(`
		WHERE l.lensId=?`, data.LensId)
	err = core.LoadLenses(db, queryPart, nil, func(db *database.DB, l *core.Lens) error {
		lens = l
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load the lens: %v", err)
	} else if lens != nil {
		return pages.Fail(fmt.Sprintf("This page is already a lens for %v", lens.PageId), nil).Status(http.StatusBadRequest)
	}

	// Check permissions
	pageIds := []string{data.PageId, data.LensId}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Begin the transaction.
	var id int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Compute the lens index
		lensIndex := 0
		row := database.NewQuery(`
			SELECT IFNULL(MAX(lensIndex)+1,0)
			FROM lenses
			WHERE pageId=?`, data.PageId).ToTxStatement(tx).QueryRow()
		_, err := row.Scan(&lensIndex)
		if err != nil {
			return sessions.NewError("Couldn't load lensIndex", err)
		}

		// Create the lens
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["lensId"] = data.LensId
		hashmap["lensIndex"] = lensIndex
		hashmap["lensName"] = fmt.Sprintf("Lens %d", lensIndex)
		hashmap["createdBy"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["updatedBy"] = u.Id
		hashmap["updatedAt"] = database.Now()
		statement := db.NewInsertStatement("lenses", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't update lenses", err)
		}

		id, err = result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get lens id", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Load the newly created lens
	queryPart = database.NewQuery(`WHERE l.id=?`, id)
	err = core.LoadLenses(db, queryPart, nil, func(db *database.DB, lens *core.Lens) error {
		returnData.ResultMap["lens"] = lens
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load the lens: %v", err)
	} else if _, ok := returnData.ResultMap["lens"]; !ok {
		return pages.Fail("Couldn't find the lens", nil).Status(http.StatusBadRequest)
	}
	return pages.Success(returnData)
}
