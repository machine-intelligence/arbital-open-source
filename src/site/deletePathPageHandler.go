// deletePathPageHandler.go deletes the given lens, while keeping the page as a child
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// deletePathPageData contains the data we get in the request
type deletePathPageData struct {
	Id string
}

var deletePathPageHandler = siteHandler{
	URI:         "/json/deletePathPage/",
	HandlerFunc: deletePathPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func deletePathPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data deletePathPageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Check if this lens exists
	pathPage, err := core.LoadPathPage(db, data.Id)
	if err != nil {
		return pages.Fail("Couldn't load the path page: %v", err)
	} else if pathPage == nil {
		return pages.Fail("Couldn't find the pathPage", nil).Status(http.StatusBadRequest)
	}

	// Check permissions
	pageIds := []string{pathPage.GuideId}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		statement := database.NewQuery(`
			DELETE FROM pathPages WHERE id=?`, data.Id).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't delete the pathPage", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
