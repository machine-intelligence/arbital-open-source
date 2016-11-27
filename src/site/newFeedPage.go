// newFeedPageHandler.go creates and returns a new page

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

const (
	AssumedFeedPageDomainID = "2069"
)

var newFeedPageHandler = siteHandler{
	URI:         "/newFeedPage/",
	HandlerFunc: newFeedPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newFeedPageData contains parameters passed in via the request.
type newFeedPageData struct {
	// Either page id has to be set...
	PageID string

	// ... or url & title have to be set
	Url   string
	Title string
}

// newFeedPageHandlerFunc handles the request.
func newFeedPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	// Decode data
	var data newFeedPageData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) && (data.Url == "" || data.Title == "") {
		return pages.Fail("Url & title have to be set if pageId isn't given", nil).Status(http.StatusBadRequest)
	}

	newFeedRow := &FeedRow{
		DomainID:    AssumedFeedPageDomainID,
		PageID:      data.PageID,
		SubmitterID: u.ID,
		CreatedAt:   database.Now(),
	}

	if !u.GetDomainMembershipRole(newFeedRow.DomainID).AtLeast(core.TrustedDomainRole) {
		return pages.Fail("You don't have permissions to submit a link to this domain", nil).Status(http.StatusBadRequest)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {

		// Create a new page for the external resource
		if !core.IsIDValid(data.PageID) {
			newFeedRow.PageID, err = core.CreateNewPage(db, u, &core.CreateNewPageOptions{
				EditDomainID: newFeedRow.DomainID,
				Title:        data.Title,
				Text:         fmt.Sprintf("Automatically generated page for an external resource: %s", data.Url),
				IsPublished:  true,
				Tx:           tx,
			})
			if err != nil {
				return sessions.NewError("Couldn't create a new page", err)
			}
		}

		hashmap := make(map[string]interface{})
		hashmap["domainId"] = newFeedRow.DomainID
		hashmap["pageId"] = newFeedRow.PageID
		hashmap["submitterId"] = newFeedRow.SubmitterID
		hashmap["createdAt"] = newFeedRow.CreatedAt
		statement := db.NewInsertStatement("feedPages", hashmap)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't insert into feedPages", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Load data
	core.AddPageIDToMap(newFeedRow.PageID, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["newFeedRow"] = newFeedRow
	return pages.Success(returnData)
}
