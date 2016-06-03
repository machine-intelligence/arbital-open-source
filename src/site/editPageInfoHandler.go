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
	PageId                   string
	Type                     string
	HasVote                  bool
	VoteType                 string
	SeeGroupId               string
	EditGroupId              string
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

	if !core.IsIdValid(data.PageId) {
		return pages.Fail("No pageId specified", nil).Status(http.StatusBadRequest)
	}

	// Load the published page.
	oldPage, err := core.LoadFullEdit(db, data.PageId, u, nil)
	if err != nil {
		return pages.Fail("Couldn't load the old page", err)
	} else if oldPage == nil {
		// Likely the page hasn't been published yet, so let's load the unpublished version.
		oldPage, err = core.LoadFullEdit(db, data.PageId, u, &core.LoadEditOptions{LoadNonliveEdit: true})
		if err != nil || oldPage == nil {
			return pages.Fail("Couldn't load the old page2", err)
		}
	}

	// Fix some data.
	if data.Type == core.CommentPageType {
		data.EditGroupId = u.Id
	}

	// Error checking.
	// Check the group settings
	if oldPage.SeeGroupId != data.SeeGroupId && oldPage.WasPublished {
		return pages.Fail("Editing this page in incorrect private group", nil).Status(http.StatusBadRequest)
	}
	if core.IsIdValid(data.SeeGroupId) && !u.IsMemberOfGroup(data.SeeGroupId) {
		return pages.Fail("Don't have group permission to EVEN SEE this page", nil).Status(http.StatusBadRequest)
	}
	if core.IsIdValid(oldPage.EditGroupId) && !u.IsMemberOfGroup(oldPage.EditGroupId) {
		return pages.Fail("Don't have group permission to edit this page", nil).Status(http.StatusBadRequest)
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
	if data.IsEditorCommentIntention && data.Type != core.CommentPageType {
		return pages.Fail("Can't set editor-comment for non-comments", nil).Status(http.StatusBadRequest)
	}

	// Make sure the user has the right permissions to edit this page
	if !oldPage.Permissions.Edit.Has {
		return pages.Fail("Can't edit: "+oldPage.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
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
	if oldPage.WasPublished {
		data.Type = oldPage.Type
	}
	// Enforce SortChildrenBy
	if data.Type == core.CommentPageType {
		data.SortChildrenBy = core.RecentFirstChildSortingOption
	} else if data.Type == core.QuestionPageType {
		data.SortChildrenBy = core.LikesChildSortingOption
	}

	// Make sure alias is valid
	if strings.ToLower(data.Alias) == "www" {
		return pages.Fail("Alias can't be 'www'", nil).Status(http.StatusBadRequest)
	} else if data.Type == core.GroupPageType || data.Type == core.DomainPageType {
		data.Alias = oldPage.Alias
	} else if data.Alias == "" {
		data.Alias = data.PageId
	} else if data.Alias != data.PageId {
		// Check if the alias matches the strict regexp
		if !core.StrictAliasRegexp.MatchString(data.Alias) {
			return pages.Fail("Invalid alias. Can only contain letters, digits, and underscores. It also cannot start with a digit.", nil)
		}

		// Prefix alias with the group alias, if appropriate
		if core.IsIdValid(data.SeeGroupId) && data.Type != core.GroupPageType && data.Type != core.DomainPageType {
			tempPageMap := map[string]*core.Page{data.SeeGroupId: core.NewPage(data.SeeGroupId)}
			err = core.LoadPages(db, u, tempPageMap)
			if err != nil {
				return pages.Fail("Couldn't load the see group", err)
			}
			data.Alias = fmt.Sprintf("%s.%s", tempPageMap[data.SeeGroupId].Alias, data.Alias)
		}

		// Check if another page is already using the alias
		var existingPageId string
		row := database.NewQuery(`
			SELECT pageId
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			WHERE pageId!=?`, data.PageId).Add(`
			AND alias=?`, data.Alias).ToStatement(db).QueryRow()
		exists, err := row.Scan(&existingPageId)
		if err != nil {
			return pages.Fail("Failed on looking for conflicting alias", err)
		} else if exists {
			return pages.Fail(fmt.Sprintf("Alias '%s' is already in use by: %s", data.Alias, existingPageId), nil)
		}
	}

	isEditorComment := oldPage.IsEditorComment
	if oldPage.Type == core.CommentPageType {
		// See if the user can affect isEditorComment's value
		if oldPage.Permissions.DomainAccess.Has {
			isEditorComment = data.IsEditorCommentIntention
		}
	}

	var changeLogIds []int64

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update pageInfos
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["alias"] = data.Alias
		hashmap["sortChildrenBy"] = data.SortChildrenBy
		hashmap["hasVote"] = hasVote
		hashmap["voteType"] = data.VoteType
		hashmap["type"] = data.Type
		hashmap["seeGroupId"] = data.SeeGroupId
		hashmap["editGroupId"] = data.EditGroupId
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
			updateChangeLog := func(changeType string, auxPageId string, oldSettingsValue string, newSettingsValue string) (int64, sessions.Error) {

				hashmap = make(database.InsertMap)
				hashmap["pageId"] = data.PageId
				hashmap["userId"] = u.Id
				hashmap["createdAt"] = database.Now()
				hashmap["type"] = changeType
				hashmap["auxPageId"] = auxPageId
				hashmap["oldSettingsValue"] = oldSettingsValue
				hashmap["newSettingsValue"] = newSettingsValue

				statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
				result, err := statement.Exec()
				if err != nil {
					return 0, sessions.NewError(fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err)
				}
				changeLogId, err := result.LastInsertId()
				if err != nil {
					return 0, sessions.NewError(fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err)
				}
				return changeLogId, nil
			}

			if data.Alias != oldPage.Alias {
				changeLogId, err2 := updateChangeLog(core.NewAliasChangeLog, "", oldPage.Alias, data.Alias)
				if err2 != nil {
					return err2
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if data.SortChildrenBy != oldPage.SortChildrenBy {
				changeLogId, err2 := updateChangeLog(core.NewSortChildrenByChangeLog, "", oldPage.SortChildrenBy, data.SortChildrenBy)
				if err2 != nil {
					return err2
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if hasVote != oldPage.HasVote {
				changeType := core.TurnOnVoteChangeLog
				if !hasVote {
					changeType = core.TurnOffVoteChangeLog
				}
				changeLogId, err2 := updateChangeLog(changeType, "", strconv.FormatBool(oldPage.HasVote), strconv.FormatBool(hasVote))
				if err2 != nil {
					return err2
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if data.VoteType != oldPage.VoteType {
				changeLogId, err2 := updateChangeLog(core.SetVoteTypeChangeLog, "", oldPage.VoteType, data.VoteType)
				if err2 != nil {
					return err2
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if data.EditGroupId != oldPage.EditGroupId {
				changeLogId, err2 := updateChangeLog(core.NewEditGroupChangeLog, data.EditGroupId, oldPage.EditGroupId, data.EditGroupId)
				if err2 != nil {
					return err2
				}
				changeLogIds = append(changeLogIds, changeLogId)
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
		task.PageId = data.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	// Generate "edit" update for users who are subscribed to this page.
	if oldPage.WasPublished {
		for _, changeLogId := range changeLogIds {
			var task tasks.NewUpdateTask
			task.UserId = u.Id
			task.GoToPageId = data.PageId
			task.SubscribedToId = data.PageId
			task.UpdateType = core.ChangeLogUpdateType
			task.GroupByPageId = data.PageId
			task.ChangeLogId = changeLogId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.Success(nil)
}
