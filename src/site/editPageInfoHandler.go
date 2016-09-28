// editPageInfoHandler.go contains the handler for editing pageInfo data.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// editPageInfoData contains parameters passed in.
type editPageInfoData struct {
	PageID                   string
	Type                     string
	HasVote                  bool
	VoteType                 string
	SeeGroupID               string
	EditGroupID              string
	Alias                    string // if empty, leave the current one
	SortChildrenBy           string
	IsRequisite              bool
	IndirectTeacher          bool
	IsEditorCommentIntention bool
}

var editPageInfoHandler = siteHandler{
	URI:         "/editPageInfo/",
	HandlerFunc: editPageInfoHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// editPageInfoHandlerFunc handles requests to create a new page.
func editPageInfoHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	// Decode data
	var data editPageInfoData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	if !core.IsIDValid(data.PageID) {
		return pages.Fail("No pageId specified", nil).Status(http.StatusBadRequest)
	}

	// Load the published page.
	editLoadOptions := &core.LoadEditOptions{
		LoadNonliveEdit: true,
		PreferLiveEdit:  true,
	}
	oldPage, err := core.LoadFullEdit(db, data.PageID, u, editLoadOptions)
	if err != nil {
		return pages.Fail("Couldn't load the old page", err)
	} else if oldPage == nil {
		return pages.Fail("Couldn't find the old page", err)
	}

	// Fix some data.
	if data.Type == core.CommentPageType {
		data.EditGroupID = u.ID
	}
	if oldPage.WasPublished {
		if (data.Type == core.WikiPageType || data.Type == core.QuestionPageType) &&
			(oldPage.Type == core.WikiPageType || oldPage.Type == core.QuestionPageType) {
			// Allow type changing from wiki <-> question
		} else {
			// Don't allow type changing
			data.Type = oldPage.Type
		}
	}

	// Error checking.
	// Check the group settings
	if oldPage.SeeGroupID != data.SeeGroupID && oldPage.WasPublished {
		return pages.Fail("Editing this page in incorrect private group", nil).Status(http.StatusBadRequest)
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	data.Type, err = core.CorrectPageType(data.Type)
	if err != nil {
		return pages.Fail(err.Error(), nil).Status(http.StatusBadRequest)
	}
	if data.SortChildrenBy != core.LikesChildSortingOption &&
		data.SortChildrenBy != core.RecentFirstChildSortingOption &&
		data.SortChildrenBy != core.OldestFirstChildSortingOption &&
		data.SortChildrenBy != core.AlphabeticalChildSortingOption {
		return pages.Fail("Invalid sort children value", nil).Status(http.StatusBadRequest)
	}
	if data.VoteType != "" && data.VoteType != core.ProbabilityVoteType && data.VoteType != core.ApprovalVoteType {
		return pages.Fail("Invalid vote type value", nil).Status(http.StatusBadRequest)
	}

	// Data correction. Rewrite the data structure so that we can just use it
	// in a straight-forward way to populate the database.
	// Can't change certain parameters after the page has been published.
	var hasVote bool
	if oldPage.WasPublished && oldPage.VoteType != "" {
		hasVote = data.HasVote
		data.VoteType = oldPage.VoteType
	} else {
		hasVote = data.VoteType != ""
	}
	// Enforce SortChildrenBy
	if data.Type == core.CommentPageType {
		data.SortChildrenBy = core.RecentFirstChildSortingOption
	} else if data.Type == core.QuestionPageType {
		data.SortChildrenBy = core.LikesChildSortingOption
	}
	// Check IsEditorCommentIntention
	if data.IsEditorCommentIntention && data.Type != core.CommentPageType {
		data.IsEditorCommentIntention = false
	}

	// Make sure alias is valid
	if strings.ToLower(data.Alias) == "www" {
		return pages.Fail("Alias can't be 'www'", nil).Status(http.StatusBadRequest)
	} else if data.Type == core.GroupPageType || data.Type == core.DomainPageType {
		data.Alias = oldPage.Alias
	} else if data.Alias == "" {
		data.Alias = data.PageID
	} else if data.Alias != data.PageID {
		// Check if the alias matches the strict regexp
		if !core.StrictAliasRegexp.MatchString(data.Alias) {
			return pages.Fail("Invalid alias. An aliases can only contain letters, digits, and underscores. It also cannot start with a digit.", nil)
		}

		// Prefix alias with the group alias, if appropriate
		if core.IsIDValid(data.SeeGroupID) && data.Type != core.GroupPageType && data.Type != core.DomainPageType {
			tempPageMap := map[string]*core.Page{data.SeeGroupID: core.NewPage(data.SeeGroupID)}
			err = core.LoadPages(db, u, tempPageMap)
			if err != nil {
				return pages.Fail("Couldn't load the see group", err)
			}
			data.Alias = fmt.Sprintf("%s.%s", tempPageMap[data.SeeGroupID].Alias, data.Alias)
		}

		// Check if another page is already using the alias
		var existingPageID string
		row := database.NewQuery(`
			SELECT pageId
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			WHERE pageId!=?`, data.PageID).Add(`
			AND alias=?`, data.Alias).ToStatement(db).QueryRow()
		exists, err := row.Scan(&existingPageID)
		if err != nil {
			return pages.Fail("Failed on looking for conflicting alias", err)
		} else if exists {
			return pages.Fail(fmt.Sprintf("Alias '%s' is already in use by: %s", data.Alias, existingPageID), nil)
		}
	}

	isEditorComment := oldPage.IsEditorComment
	if oldPage.Type == core.CommentPageType {
		// See if the user can affect isEditorComment's value
		if oldPage.Permissions.Comment.Has {
			isEditorComment = data.IsEditorCommentIntention
		}
	}

	// Check if something is actually different from live edit
	// NOTE: we do this as the last step before writing data, just so we can be sure
	// exactly what date we'll be writing
	if !oldPage.IsDeleted {
		if data.Alias == oldPage.Alias &&
			data.SortChildrenBy == oldPage.SortChildrenBy &&
			data.HasVote == oldPage.HasVote &&
			data.VoteType == oldPage.VoteType &&
			data.Type == oldPage.Type &&
			data.SeeGroupID == oldPage.SeeGroupID &&
			data.EditGroupID == oldPage.EditGroupID &&
			data.IsRequisite == oldPage.IsRequisite &&
			data.IndirectTeacher == oldPage.IndirectTeacher &&
			isEditorComment == oldPage.IsEditorComment &&
			data.IsEditorCommentIntention == oldPage.IsEditorCommentIntention {
			return pages.Success(nil)
		}
	}

	// Make sure the user has the right permissions to edit this page
	// NOTE: check permissions AFTER checking if any data will be changed, becase we
	// don't want to flag the user for not having correct permissions, when they are
	// not actually changing anything
	if !oldPage.Permissions.Edit.Has {
		return pages.Fail("Can't edit: "+oldPage.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}

	var changeLogIDs []int64

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update pageInfos
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageID
		hashmap["alias"] = data.Alias
		hashmap["sortChildrenBy"] = data.SortChildrenBy
		hashmap["hasVote"] = hasVote
		hashmap["voteType"] = data.VoteType
		hashmap["type"] = data.Type
		hashmap["seeGroupId"] = data.SeeGroupID
		hashmap["editGroupId"] = data.EditGroupID
		hashmap["isRequisite"] = data.IsRequisite
		hashmap["indirectTeacher"] = data.IndirectTeacher
		hashmap["isEditorComment"] = isEditorComment
		hashmap["isEditorCommentIntention"] = data.IsEditorCommentIntention
		statement := tx.DB.NewInsertStatement("pageInfos", hashmap, hashmap.GetKeys()...).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos", err)
		}

		// Update change logs
		if oldPage.WasPublished {
			updateChangeLog := func(changeType string, auxPageID string, oldSettingsValue string, newSettingsValue string) (int64, sessions.Error) {

				hashmap = make(database.InsertMap)
				hashmap["pageId"] = data.PageID
				hashmap["userId"] = u.ID
				hashmap["createdAt"] = database.Now()
				hashmap["type"] = changeType
				hashmap["auxPageId"] = auxPageID
				hashmap["oldSettingsValue"] = oldSettingsValue
				hashmap["newSettingsValue"] = newSettingsValue

				statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
				result, err := statement.Exec()
				if err != nil {
					return 0, sessions.NewError(fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err)
				}
				changeLogID, err := result.LastInsertId()
				if err != nil {
					return 0, sessions.NewError(fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err)
				}
				return changeLogID, nil
			}

			if data.Alias != oldPage.Alias {
				changeLogID, err2 := updateChangeLog(core.NewAliasChangeLog, "", oldPage.Alias, data.Alias)
				if err2 != nil {
					return err2
				}
				changeLogIDs = append(changeLogIDs, changeLogID)
			}
			if data.SortChildrenBy != oldPage.SortChildrenBy {
				changeLogID, err2 := updateChangeLog(core.NewSortChildrenByChangeLog, "", oldPage.SortChildrenBy, data.SortChildrenBy)
				if err2 != nil {
					return err2
				}
				changeLogIDs = append(changeLogIDs, changeLogID)
			}
			if hasVote != oldPage.HasVote {
				changeType := core.TurnOnVoteChangeLog
				if !hasVote {
					changeType = core.TurnOffVoteChangeLog
				}
				changeLogID, err2 := updateChangeLog(changeType, "", strconv.FormatBool(oldPage.HasVote), strconv.FormatBool(hasVote))
				if err2 != nil {
					return err2
				}
				changeLogIDs = append(changeLogIDs, changeLogID)
			}
			if data.VoteType != oldPage.VoteType {
				changeLogID, err2 := updateChangeLog(core.SetVoteTypeChangeLog, "", oldPage.VoteType, data.VoteType)
				if err2 != nil {
					return err2
				}
				changeLogIDs = append(changeLogIDs, changeLogID)
			}
			if data.EditGroupID != oldPage.EditGroupID {
				changeLogID, err2 := updateChangeLog(core.NewEditGroupChangeLog, data.EditGroupID, oldPage.EditGroupID, data.EditGroupID)
				if err2 != nil {
					return err2
				}
				changeLogIDs = append(changeLogIDs, changeLogID)
			}

		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// === Once the transaction has succeeded, we can't really fail on anything
	// else. So we print out errors, but don't return an error. ===

	// Update elastic search index.
	if oldPage.WasPublished {
		var task tasks.UpdateElasticPageTask
		task.PageID = data.PageID
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	// Generate "edit" update for users who are subscribed to this page.
	if oldPage.WasPublished {
		for _, changeLogID := range changeLogIDs {
			var task tasks.NewUpdateTask
			task.UserID = u.ID
			task.GoToPageID = data.PageID
			task.SubscribedToID = data.PageID
			task.UpdateType = core.ChangeLogUpdateType
			task.ChangeLogID = changeLogID
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.Success(nil)
}
