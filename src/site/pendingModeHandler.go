// Provide data for "pending" mode, like what pages have been submitted to a domain or
// what edits have be proposed but not approved yet.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type pendingModeData struct {
	NumPagesToLoad int
}

var pendingModeHandler = siteHandler{
	URI:         "/json/pending/",
	HandlerFunc: pendingModeHandlerFunc,
	Options:     pages.PageOptions{},
}

func pendingModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data pendingModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load new page submissions
	returnData.ResultMap["pageToDomainSubmissionRows"], err = loadPageToDomainSubmissionRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	// Load new edit proposals
	returnData.ResultMap["editProposalRows"], err = loadEditProposalRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

// Load pages that have been submitted to a domain, but haven't been approved yet
func loadPageToDomainSubmissionRows(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]*core.PageToDomainSubmission, error) {
	submissions := make([]*core.PageToDomainSubmission, 0)

	pageLoadOptions := (&core.PageLoadOptions{
		SubmittedTo: true,
	}).Add(core.TitlePlusLoadOptions)

	queryPart := database.NewQuery(`
		WHERE approverId=""
		ORDER BY createdAt DESC
		LIMIT ?`, limit)
	err := core.LoadPageToDomainSubmissions(db, queryPart, func(db *database.DB, submission *core.PageToDomainSubmission) error {
		core.AddPageToMap(submission.PageId, returnData.PageMap, pageLoadOptions)
		submissions = append(submissions, submission)
		return nil
	})
	return submissions, err
}

// Load edit proposals that have been submitted but not accepted yet
func loadEditProposalRows(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]*core.ChangeLog, error) {
	changeLogs := make([]*core.ChangeLog, 0)

	pageLoadOptions := (&core.PageLoadOptions{
		DomainsAndPermissions: true,
		EditHistory:           true,
	}).Add(core.EmptyLoadOptions)

	queryPart := database.NewQuery(`
		WHERE type=?`, core.NewEditProposalChangeLog).Add(`
		ORDER BY createdAt DESC
		LIMIT ?`, limit)
	err := core.LoadChangeLogs(db, queryPart, returnData, func(db *database.DB, changeLog *core.ChangeLog) error {
		core.AddPageToMap(changeLog.PageId, returnData.PageMap, pageLoadOptions)
		changeLogs = append(changeLogs, changeLog)
		return nil
	})
	return changeLogs, err
}
