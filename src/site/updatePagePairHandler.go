// updatePagePairHandler.go handles updating a page pair
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

// updatePagePairData contains the data we get in the request.
type updatePagePairData struct {
	ID       string
	Level    int
	IsStrong bool
}

var updatePagePairHandler = siteHandler{
	URI:         "/updatePagePair/",
	HandlerFunc: updatePagePairHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updatePagePairHandlerFunc handles requests for adding a update tag.
func updatePagePairHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data updatePagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	return updatePagePairHandlerInternal(params.DB, params.U, &data)
}

func updatePagePairHandlerInternal(db *database.DB, u *core.CurrentUser, data *updatePagePairData) *pages.Result {
	returnData := core.NewHandlerData(u)
	var err error

	// Load the existing page pair
	var pagePair *core.PagePair
	queryPart := database.NewQuery(`WHERE pp.id=?`, data.ID)
	err = core.LoadPagePairs(db, queryPart, func(db *database.DB, pp *core.PagePair) error {
		pagePair = pp
		return nil
	})
	if err != nil {
		return pages.Fail("Failed to check for existing page pair: %v", err)
	} else if pagePair == nil {
		return pages.Fail("Couldn't find the page pair", nil).Status(http.StatusBadRequest)
	}

	// Load pages
	pagePair = &core.PagePair{
		ParentId: pagePair.ParentId,
		ChildId:  pagePair.ChildId,
		Type:     pagePair.Type,
	}
	parent, child, err := core.LoadFullEditsForPagePair(db, pagePair, u)
	if err != nil {
		return pages.Fail("Error loading pagePair pages", err)
	}

	// Check edit permissions
	permissionError, err := core.CanAffectRelationship(db.C, parent, child, pagePair.Type)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Do it!
	var pagePairId int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update page pair
		hashmap := make(database.InsertMap)
		hashmap["id"] = data.ID
		hashmap["level"] = data.Level
		hashmap["isStrong"] = data.IsStrong
		statement := tx.DB.NewInsertStatement("pagePairs", hashmap, "level", "isStrong").WithTx(tx)
		_, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert pagePair", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	returnData.ResultMap["pagePair"], err = core.LoadPagePair(db, fmt.Sprintf("%d", pagePairId))
	if err != nil {
		return pages.Fail("Error loading the page pair", err)
	}
	return pages.Success(returnData)
}
