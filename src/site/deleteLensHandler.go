// deleteLensHandler.go deletes the given lens, while keeping the page as a child
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// deleteLensData contains the data we get in the request
type deleteLensData struct {
	Id string
}

var deleteLensHandler = siteHandler{
	URI:         "/json/deleteLens/",
	HandlerFunc: deleteLensHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func deleteLensHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data deleteLensData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Check if this lens exists
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
		statement := database.NewQuery(`
			DELETE FROM lenses WHERE id=?`, data.Id).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't delete the lens", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
