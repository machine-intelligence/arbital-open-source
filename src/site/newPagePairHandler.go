// newPagePairHandler.go handles repages for adding a new tag.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// newPagePairData contains the data we get in the request.
type newPagePairData struct {
	ParentId string
	ChildId  string
	Type     string
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

	return newPagePairHandlerInternal(params, &data)
}

func newPagePairHandlerInternal(params *pages.HandlerParams, data *newPagePairData) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U
	var err error

	// Error checking
	if !core.IsIdValid(data.ParentId) || !core.IsIdValid(data.ChildId) {
		return pages.Fail("ParentId and ChildId have to be set", nil).Status(http.StatusBadRequest)
	}
	if data.ParentId == data.ChildId &&
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
	queryPart := database.NewQuery(`WHERE parentId=? AND childId=? AND type=?`, data.ParentId, data.ChildId, data.Type)
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
		childIsAncestor, err := core.IsAncestor(db, data.ChildId, data.ParentId)
		if err != nil {
			return pages.Fail("Failed to check for ancestor: %v", err)
		} else if childIsAncestor {
			return pages.Fail("Error: adding the relationship would create a cycle", nil)
		}
	}

	// Load pages
	pagePair = &core.PagePair{
		ParentId: data.ParentId,
		ChildId:  data.ChildId,
		Type:     data.Type,
	}
	parent, child, err := core.LoadFullEditsForPagePair(db, pagePair, u)
	if err != nil {
		return pages.Fail("Error loading pagePair pages", err)
	}

	// Check edit permissions
	permissionError, err := core.CanAffectRelationship(c, parent, child, data.Type)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Do it!
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Create new page pair
		hashmap := make(database.InsertMap)
		hashmap["parentId"] = data.ParentId
		hashmap["childId"] = data.ChildId
		hashmap["type"] = data.Type
		hashmap["creatorId"] = u.Id
		hashmap["createdAt"] = database.Now()
		statement := tx.DB.NewInsertStatement("pagePairs", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert pagePair", err)
		}
		pagePairId, err := resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get page pair id", err)
		}

		// Go ahead and update the domains for the child page
		// (we'll handle its descendants in the PublishPagePairTask)
		if data.Type == core.ParentPagePairType {
			err = core.PropagateDomainsWithTx(tx, []string{data.ChildId})
			if err != nil {
				return sessions.NewError("Couldn't update domains for the child page", err)
			}
		}

		var task tasks.PublishPagePairTask
		task.PagePairId = fmt.Sprintf("%d", pagePairId)
		err = tasks.Enqueue(c, &task, nil)
		if err != nil {
			return sessions.NewError("Couldn't enqueue the task", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
