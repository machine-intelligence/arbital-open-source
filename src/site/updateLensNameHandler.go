// updateLensNameHandler.go updates the name of a lens
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// updateLensNameData contains the data we get in the request
type updateLensNameData struct {
	Id   int64 `json:"id,string"`
	Name string
}

var updateLensNameHandler = siteHandler{
	URI:         "/json/updateLensName/",
	HandlerFunc: updateLensNameHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateLensNameHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updateLensNameData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if len(data.Name) <= 2 {
		return pages.Fail("Lens name is too short", nil).Status(http.StatusBadRequest)
	}

	// Load the lens
	var lens *core.Lens
	queryPart := database.NewQuery(`
		WHERE l.id=?`, data.Id)
	err = core.LoadLenses(db, queryPart, nil, func(db *database.DB, l *core.Lens) error {
		lens = l
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load the lens: %v", err)
	} else if lens == nil {
		return pages.Fail("Couldn't find the lens", nil).Status(http.StatusBadRequest)
	}

	// Check permissions
	pageIds := []string{lens.PageId, lens.LensId}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update the lens name
		hashmap := make(database.InsertMap)
		hashmap["id"] = lens.Id
		hashmap["lensName"] = data.Name
		hashmap["updatedBy"] = u.Id
		hashmap["updatedAt"] = database.Now()
		statement := db.NewInsertStatement("lenses", hashmap, "lensName", "updatedBy", "updatedAt").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update lenses", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
