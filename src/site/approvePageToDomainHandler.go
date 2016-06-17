// approvePageToDomainHandler.go approves a page that was submitted to a domain
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

// Contains data given to us in the request.
type approvePageToDomainData struct {
	PageId   string `json:"pageId"`
	DomainId string `json:"domainId"`
}

var approvePageToDomainHandler = siteHandler{
	URI:         "/json/approvePageToDomain/",
	HandlerFunc: approvePageToDomainHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func approvePageToDomainHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	var data approvePageToDomainData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	} else if !core.IsIdValid(data.DomainId) {
		return pages.Fail("Invalid domain id", nil).Status(http.StatusBadRequest)
	}

	// Load the page info
	oldPage, err := core.LoadFullEdit(db, data.PageId, u, nil)
	if err != nil {
		return pages.Fail("Error loading the page", err)
	} else if oldPage == nil {
		return pages.Fail("Couldn't find the page", nil)
	}

	// Check permissions
	if !u.TrustMap[data.DomainId].Permissions.DomainAccess.Has {
		return pages.Fail("You don't have access to the domain", nil)
	}

	// Load the submission info
	submission, err := core.LoadPageToDomainSubmission(db, data.PageId, data.DomainId)
	if err != nil {
		return pages.Fail("Couldn't load submission", err)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {

		// Approve the page
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["domainId"] = data.DomainId
		hashmap["approverId"] = u.Id
		hashmap["approvedAt"] = database.Now()
		statement := db.NewInsertStatement("pageToDomainSubmissions", hashmap, "approvedAt", "approverId").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't add submission", err)
		}

		// Check if the page already has a parent that's in the math domain
		var parentCount int
		_, err := database.NewQuery(`
			SELECT COUNT(*)
			FROM pagePairs AS pp
			JOIN pageDomainPairs AS pdp
			ON (pp.parentId=pdp.pageId)
			WHERE pp.type=?`, core.ParentPagePairType).Add(`
				AND pp.childId=?`, data.PageId).Add(`
				AND pdp.domainId=?`, data.DomainId).ToTxStatement(tx).QueryRow().Scan(&parentCount)
		if err != nil {
			return sessions.NewError("Couldn't load parents", err)
		}

		// If no parent, add domain as a parent
		if parentCount <= 0 {
			handlerData := newPagePairData{
				ParentId: data.DomainId,
				ChildId:  data.PageId,
				Type:     core.ParentPagePairType,
			}
			result := newPagePairHandlerInternal(params, &handlerData)
			if result.Err != nil {
				return result.Err
			}
		}

		// Notify page creator and the person who submitted the page to domain
		err = insertPageToDomainAcceptedUpdate(tx, u, submission.SubmitterId, data.PageId, data.DomainId)
		if err != nil {
			return sessions.NewError("Couldn't insert update for submitter", err)
		}
		if submission.SubmitterId != oldPage.PageCreatorId {
			err = insertPageToDomainAcceptedUpdate(tx, u, oldPage.PageCreatorId, data.PageId, data.DomainId)
			if err != nil {
				return sessions.NewError("Couldn't insert update for creator", err)
			}
		}

		// Subscribe the approver as a maintainer
		err2 := addSubscription(tx, u.Id, data.PageId, true)
		if err2 != nil {
			return err2
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Load the page with the new domain permissions
	loadOptions := (&core.PageLoadOptions{
		Parents:     true,
		SubmittedTo: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.PageId, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

// Add an update for the given user about a page being accepted into a domain
func insertPageToDomainAcceptedUpdate(tx *database.Tx, u *core.CurrentUser, forUserId, pageId, domainId string) error {
	if u.Id == forUserId {
		return nil
	}
	hashmap := make(map[string]interface{})
	hashmap["userId"] = forUserId
	hashmap["byUserId"] = u.Id
	hashmap["type"] = core.PageToDomainAcceptedUpdateType
	hashmap["groupByPageId"] = pageId
	hashmap["subscribedToId"] = domainId
	hashmap["goToPageId"] = pageId
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("updates", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't create new update: %v", err)
	}
	return nil
}
