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
	if !u.TrustMap[data.DomainId].Permissions.DomainTrust.Has {
		return pages.Fail(u.TrustMap[data.DomainId].Permissions.DomainTrust.Reason, nil)
	}

	// Load the submission info
	submission, err := core.LoadPageToDomainSubmission(db, data.PageId, data.DomainId)
	if err != nil {
		return pages.Fail("Couldn't load submission", err)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return approvePageToDomainTx(tx, u, submission, oldPage.PageCreatorId)
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
			AND pp.childId=?`, submission.PageId).Add(`
			AND pdp.domainId=?`, submission.DomainId).ToStatement(db).QueryRow().Scan(&parentCount)
	if err != nil {
		return pages.Fail("Couldn't load parents", err)
	}

	// If no parent, add domain as a parent
	if parentCount <= 0 {
		handlerData := newPagePairData{
			ParentId: submission.DomainId,
			ChildId:  submission.PageId,
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
	core.AddPageToMap(data.PageId, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func approvePageToDomainTx(tx *database.Tx, approver *core.CurrentUser, submission *core.PageToDomainSubmission,
	pageCreatorId string) sessions.Error {

	// Approve the page
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = submission.PageId
	hashmap["domainId"] = submission.DomainId
	hashmap["approverId"] = approver.Id
	hashmap["approvedAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("pageToDomainSubmissions", hashmap, "approvedAt", "approverId").WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't add submission", err)
	}

	// Notify page creator and the person who submitted the page to domain
	err := insertPageToDomainAcceptedUpdate(tx, approver.Id, submission.SubmitterId, submission.PageId, submission.DomainId)
	if err != nil {
		return sessions.NewError("Couldn't insert update for submitter", err)
	}
	if submission.SubmitterId != pageCreatorId {
		err = insertPageToDomainAcceptedUpdate(tx, approver.Id, pageCreatorId, submission.PageId, submission.DomainId)
		if err != nil {
			return sessions.NewError("Couldn't insert update for creator", err)
		}
	}

	// Subscribe the approver as a maintainer
	serr := addSubscription(tx, approver.Id, submission.PageId, true)
	if serr != nil {
		return serr
	}

	return nil
}

// Add an update for the given user about a page being accepted into a domain
func insertPageToDomainAcceptedUpdate(tx *database.Tx, approverId, forUserId, pageId, domainId string) error {
	if approverId == forUserId {
		return nil
	}
	hashmap := make(map[string]interface{})
	hashmap["userId"] = forUserId
	hashmap["byUserId"] = approverId
	hashmap["type"] = core.PageToDomainAcceptedUpdateType
	hashmap["subscribedToId"] = domainId
	hashmap["goToPageId"] = pageId
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("updates", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't create new update: %v", err)
	}
	return nil
}
