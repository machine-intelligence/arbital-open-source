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
	PageID   string `json:"pageId"`
	DomainID string `json:"domainId"`
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
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	} else if !core.IsIDValid(data.DomainID) {
		return pages.Fail("Invalid domain id", nil).Status(http.StatusBadRequest)
	}

	// Load the page info
	oldPage, err := core.LoadFullEdit(db, data.PageID, u, nil)
	if err != nil {
		return pages.Fail("Error loading the page", err)
	} else if oldPage == nil {
		return pages.Fail("Couldn't find the page", nil)
	}

	// Check permissions
	if !core.RoleAtLeast(u.GetDomainMembershipRole(data.DomainID), core.ReviewerDomainRole) {
		return pages.Fail("You don't have permission to do this", nil)
	}

	// Load the submission info
	submission, err := core.LoadPageToDomainSubmission(db, data.PageID, data.DomainID)
	if err != nil {
		return pages.Fail("Couldn't load submission", err)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return approvePageToDomainTx(tx, u, submission, oldPage.PageCreatorID)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Check if the page already has a parent that's in the math domain
	var parentCount int
	_, err = database.NewQuery(`
		SELECT COUNT(*)
		FROM pagePairs AS pp
		JOIN pageDomainPairs AS pdp
		ON (pp.parentId=pdp.pageId)
		WHERE pp.type=?`, core.ParentPagePairType).Add(`
			AND pp.childId=?`, submission.PageID).Add(`
			AND pdp.domainId=?`, submission.DomainID).ToStatement(db).QueryRow().Scan(&parentCount)
	if err != nil {
		return pages.Fail("Couldn't load parents", err)
	}

	// If no parent, add domain as a parent
	if parentCount <= 0 {
		handlerData := newPagePairData{
			ParentID: submission.DomainID,
			ChildID:  submission.PageID,
			Type:     core.ParentPagePairType,
		}
		result := newPagePairHandlerInternal(db, u, &handlerData)
		if result.Err != nil {
			return pages.Fail("Couldn't add domain as parent", fmt.Errorf("Failed to add page pair: %v", result.Err))
		}
	}

	// Load the page with the new domain permissions
	loadOptions := (&core.PageLoadOptions{
		Parents:     true,
		SubmittedTo: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.PageID, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func approvePageToDomainTx(tx *database.Tx, approver *core.CurrentUser, submission *core.PageToDomainSubmission,
	pageCreatorID string) sessions.Error {

	// Approve the page
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = submission.PageID
	hashmap["domainId"] = submission.DomainID
	hashmap["approverId"] = approver.ID
	hashmap["approvedAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("pageToDomainSubmissions", hashmap, "approvedAt", "approverId").WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't add submission", err)
	}

	// Notify page creator and the person who submitted the page to domain
	err := insertPageToDomainAcceptedUpdate(tx, approver.ID, submission.SubmitterID, submission.PageID, submission.DomainID)
	if err != nil {
		return sessions.NewError("Couldn't insert update for submitter", err)
	}
	if submission.SubmitterID != pageCreatorID {
		err = insertPageToDomainAcceptedUpdate(tx, approver.ID, pageCreatorID, submission.PageID, submission.DomainID)
		if err != nil {
			return sessions.NewError("Couldn't insert update for creator", err)
		}
	}

	// Subscribe the approver as a maintainer
	serr := addSubscription(tx, approver.ID, submission.PageID, true)
	if serr != nil {
		return serr
	}

	return nil
}

// Add an update for the given user about a page being accepted into a domain
func insertPageToDomainAcceptedUpdate(tx *database.Tx, approverID, forUserID, pageID, domainID string) error {
	if approverID == forUserID {
		return nil
	}
	hashmap := make(map[string]interface{})
	hashmap["userId"] = forUserID
	hashmap["byUserId"] = approverID
	hashmap["type"] = core.PageToDomainAcceptedUpdateType
	hashmap["subscribedToId"] = domainID
	hashmap["goToPageId"] = pageID
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("updates", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't create new update: %v", err)
	}
	return nil
}
