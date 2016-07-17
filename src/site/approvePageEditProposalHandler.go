// approvePageEditProposalHandler.go contains the handler for editing pageInfo data.
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

// approvePageEditProposalData contains parameters passed in.
type approvePageEditProposalData struct {
	ChangeLogId string
	// If this is set, we are dismissing the proposal instead
	Dismiss bool
}

var approvePageEditProposalHandler = siteHandler{
	URI:         "/json/approvePageEditProposal/",
	HandlerFunc: approvePageEditProposalHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// approvePageEditProposalHandlerFunc handles requests to create a new page.
func approvePageEditProposalHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	// Decode data
	var data approvePageEditProposalData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Load the changelog
	changelogs, err := core.LoadChangeLogsByIds(db, []string{data.ChangeLogId}, core.NewEditProposalChangeLog)
	if err != nil {
		return pages.Fail("Couldn't load changelog", err)
	}
	changeLog, ok := changelogs[data.ChangeLogId]
	if !ok {
		return pages.Fail("Couldn't find changelog", nil).Status(http.StatusBadRequest)
	}

	// Load the published page.
	oldPage, err := core.LoadFullEdit(db, changeLog.PageId, u, nil)
	if err != nil {
		return pages.Fail("Couldn't load the old page", err)
	} else if oldPage == nil {
		return pages.Fail("Couldn't find the old page", err)
	}

	// Make sure the user has the right permissions to edit this page
	if !oldPage.Permissions.Edit.Has {
		return pages.Fail("Can't edit: "+oldPage.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}

	// Load the proposed edit
	proposedEdit, err := core.LoadFullEdit(db, changeLog.PageId, u, &core.LoadEditOptions{LoadSpecificEdit: changeLog.Edit})
	if err != nil {
		return pages.Fail("Couldn't load the proposed edit", err)
	} else if proposedEdit == nil {
		return pages.Fail("Couldn't find the proposed edit", err)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		if !data.Dismiss {
			// Update pageInfos
			hashmap := make(database.InsertMap)
			hashmap["pageId"] = proposedEdit.PageId
			hashmap["currentEdit"] = proposedEdit.Edit
			statement := tx.DB.NewInsertStatement("pageInfos", hashmap, "currentEdit").WithTx(tx)
			if _, err = statement.Exec(); err != nil {
				return sessions.NewError("Couldn't update pageInfos", err)
			}

			// Update pages
			statement = database.NewQuery(`
			UPDATE pages SET isLiveEdit=(edit=?)`, proposedEdit.Edit).Add(`
			WHERE pageId=?`, proposedEdit.PageId).ToTxStatement(tx)
			if _, err = statement.Exec(); err != nil {
				return sessions.NewError("Couldn't update pages", err)
			}
		}

		// Update change log's type
		hashmap := make(database.InsertMap)
		hashmap["id"] = data.ChangeLogId
		hashmap["type"] = core.NewEditChangeLog
		statement := tx.DB.NewInsertStatement("changeLogs", hashmap, "type").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update change log", err)
		}

		// Add an update for the user who submitted the edit
		if !data.Dismiss {
			if changeLog.UserId != u.Id {
				hashmap = make(map[string]interface{})
				hashmap["userId"] = changeLog.UserId
				hashmap["byUserId"] = u.Id
				hashmap["type"] = core.EditProposalAcceptedUpdateType
				hashmap["subscribedToId"] = proposedEdit.PageId
				hashmap["goToPageId"] = proposedEdit.PageId
				hashmap["changeLogId"] = data.ChangeLogId
				hashmap["createdAt"] = database.Now()
				statement = tx.DB.NewInsertStatement("updates", hashmap).WithTx(tx)
				if _, err := statement.Exec(); err != nil {
					return sessions.NewError("Couldn't create new update: %v", err)
				}
			}

			// Update the links table.
			err = core.UpdatePageLinks(tx, proposedEdit.PageId, proposedEdit.Text, sessions.GetDomain())
			if err != nil {
				return sessions.NewError("Couldn't update links", err)
			}
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Update elastic search index.
	if !data.Dismiss && proposedEdit.WasPublished {
		var task tasks.UpdateElasticPageTask
		task.PageId = proposedEdit.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return pages.Success(nil)
}
