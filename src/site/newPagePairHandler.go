// newPagePairHandler.go handles creating and updating a page

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// newPagePairData contains the data we get in the request.
type newPagePairData struct {
	ParentID string
	ChildID  string
	Type     string
	Level    int
	IsStrong bool
}

var newPagePairHandler = siteHandler{
	URI:         "/newPagePair/",
	HandlerFunc: newPagePairHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newPagePairHandlerFunc handles requests for adding a new tag.
func newPagePairHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data newPagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	return newPagePairHandlerInternal(params.DB, params.U, &data)
}

func newPagePairHandlerInternal(db *database.DB, u *core.CurrentUser, data *newPagePairData) *pages.Result {
	c := db.C
	returnData := core.NewHandlerData(u)
	var err error

	// Error checking
	if !core.IsIDValid(data.ParentID) || !core.IsIDValid(data.ChildID) {
		return pages.Fail("ParentId and ChildId have to be set", nil).Status(http.StatusBadRequest)
	}
	if data.ParentID == data.ChildID &&
		data.Type != core.SubjectPagePairType &&
		data.Type != core.RequirementPagePairType {
		return pages.Fail("ParentId equals ChildId", nil).Status(http.StatusBadRequest)
	}
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.Fail("Incorrect type", err).Status(http.StatusBadRequest)
	}

	// Check if this page pair already exists
	var pagePair *core.PagePair
	queryPart := database.NewQuery(`WHERE pp.parentId=? AND pp.childId=? AND pp.type=?`, data.ParentID, data.ChildID, data.Type)
	err = core.LoadPagePairs(db, queryPart, func(db *database.DB, pp *core.PagePair) error {
		pagePair = pp
		return nil
	})
	if err != nil {
		return pages.Fail("Failed to check for existing page pair: %v", err)
	} else if pagePair != nil {
		return pages.Success(nil)
	}

	// Check if adding the relationship would create a parent-child cycle
	if data.Type == core.ParentPagePairType {
		childIsAncestor, err := core.IsAncestor(db, data.ChildID, data.ParentID)
		if err != nil {
			return pages.Fail("Failed to check for ancestor: %v", err)
		} else if childIsAncestor {
			return pages.Fail("Error: adding the relationship would create a cycle", nil)
		}
	}

	// Load pages
	pagePair = &core.PagePair{
		ParentID: data.ParentID,
		ChildID:  data.ChildID,
		Type:     data.Type,
	}
	parent, child, err := core.LoadFullEditsForPagePair(db, pagePair, u)
	if err != nil {
		return pages.Fail("Error loading pagePair pages", err)
	}

	// Check edit permissions
	permissionError, err := core.CanAffectRelationship(db.C, parent, child, data.Type)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Do it!
	var pagePairID string
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Create new page pair
		pagePairID, err := core.CreateNewPagePair(db, u, &core.CreateNewPagePairOptions{
			ParentID: data.ParentID,
			ChildID:  data.ChildID,
			Type:     data.Type,
			Level:    data.Level,
			IsStrong: data.IsStrong,
			Tx:       tx,
		})
		if err != nil {
			return sessions.NewError("Couldn't insert pagePair", err)
		}

		var task tasks.PublishPagePairTask
		task.UserID = u.ID
		task.PagePairID = pagePairID
		err = tasks.Enqueue(c, &task, nil)
		if err != nil {
			return sessions.NewError("Couldn't enqueue the task", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	returnData.ResultMap["pagePair"], err = core.LoadPagePair(db, pagePairID)
	if err != nil {
		return pages.Fail("Error loading the page pair", err)
	}
	return pages.Success(returnData)
}
