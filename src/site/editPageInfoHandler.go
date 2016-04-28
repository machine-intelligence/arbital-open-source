// editPageInfoHandler.go contains the handler for editing pageInfo data.
package site

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// editPageInfoData contains parameters passed in.
type editPageInfoData struct {
	PageId          string
	Type            string
	HasVote         bool
	VoteType        string
	SeeGroupId      string
	EditGroupId     string
	Alias           string // if empty, leave the current one
	SortChildrenBy  string
	IsRequisite     bool
	IndirectTeacher bool
	IsEditorComment bool
}

var editPageInfoHandler = siteHandler{
	URI:         "/editPageInfo/",
	HandlerFunc: editPageInfoHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
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
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("No pageId specified", nil)
	}

	// Load the published page.
	oldPage, err := core.LoadFullEdit(db, data.PageId, u, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the old page", err)
	} else if oldPage == nil {
		oldPage = &core.Page{}
	}

	// Fix some data.
	if data.Type == core.CommentPageType {
		data.EditGroupId = u.Id
	}

	// Error checking.
	// Check the group settings
	if oldPage.SeeGroupId != data.SeeGroupId && oldPage.WasPublished {
		return pages.HandlerBadRequestFail("Editing this page in incorrect private group", nil)
	}
	if core.IsIdValid(data.SeeGroupId) && !u.IsMemberOfGroup(data.SeeGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
	}
	if core.IsIdValid(oldPage.EditGroupId) && !u.IsMemberOfGroup(oldPage.EditGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to edit this page", nil)
	}
	// Check validity of most options. (We are super permissive with autosaves.)
	data.Type, err = core.CorrectPageType(data.Type)
	if err != nil {
		return pages.HandlerBadRequestFail(err.Error(), nil)
	}
	if data.SortChildrenBy != core.LikesChildSortingOption &&
		data.SortChildrenBy != core.RecentFirstChildSortingOption &&
		data.SortChildrenBy != core.OldestFirstChildSortingOption &&
		data.SortChildrenBy != core.AlphabeticalChildSortingOption {
		return pages.HandlerBadRequestFail("Invalid sort children value", nil)
	}
	if data.VoteType != "" && data.VoteType != core.ProbabilityVoteType && data.VoteType != core.ApprovalVoteType {
		return pages.HandlerBadRequestFail("Invalid vote type value", nil)
	}

	// Make sure the user has the right permissions to edit this page
	if !oldPage.CanEdit {
		return pages.HandlerBadRequestFail(oldPage.CantEditMessage, nil)
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
		return pages.HandlerBadRequestFail("Alias can't be 'www'", nil)
	} else if data.Type == core.GroupPageType || data.Type == core.DomainPageType {
		data.Alias = oldPage.Alias
	} else if data.Alias == "" {
		data.Alias = data.PageId
	} else if data.Alias != data.PageId {
		// Check if the alias matches the strict regexp
		if !core.StrictAliasRegexp.MatchString(data.Alias) {
			return pages.HandlerErrorFail("Invalid alias. Can only contain letters and digits. It cannot be a number.", nil)
		}

		// Prefix alias with the group alias, if appropriate
		if core.IsIdValid(data.SeeGroupId) && data.Type != core.GroupPageType && data.Type != core.DomainPageType {
			tempPageMap := map[string]*core.Page{data.SeeGroupId: core.NewPage(data.SeeGroupId)}
			err = core.LoadPages(db, u, tempPageMap)
			if err != nil {
				return pages.HandlerErrorFail("Couldn't load the see group", err)
			}
			data.Alias = fmt.Sprintf("%s.%s", tempPageMap[data.SeeGroupId].Alias, data.Alias)
		}

		// Check if another page is already using the alias
		var existingPageId string
		row := db.NewStatement(`
			SELECT pageId
			FROM pageInfos
			WHERE currentEdit>0 AND NOT isDeleted AND pageId!=? AND alias=?`).QueryRow(data.PageId, data.Alias)
		exists, err := row.Scan(&existingPageId)
		if err != nil {
			return pages.HandlerErrorFail("Failed on looking for conflicting alias", err)
		} else if exists {
			return pages.HandlerErrorFail(fmt.Sprintf("Alias '%s' is already in use by: %s", data.Alias, existingPageId), nil)
		}
	}

	var changeLogIds []int64

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
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
		hashmap["isEditorComment"] = data.IsEditorComment
		statement := tx.DB.NewInsertStatement("pageInfos", hashmap, hashmap.GetKeys()...).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't update pageInfos", err
		}

		// Update change logs
		if oldPage.WasPublished {
			updateChangeLog := func(changeType string, auxPageId string, oldSettingsValue string, newSettingsValue string) (int64, string, error) {

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
					return 0, fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err
				}
				changeLogId, err := result.LastInsertId()
				if err != nil {
					return 0, fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err
				}
				return changeLogId, "", nil
			}

			if data.Alias != oldPage.Alias {
				changeLogId, errorMessage, err := updateChangeLog(core.NewAliasChangeLog, "", oldPage.Alias, data.Alias)
				if errorMessage != "" {
					return errorMessage, err
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if data.SortChildrenBy != oldPage.SortChildrenBy {
				changeLogId, errorMessage, err := updateChangeLog(core.NewSortChildrenByChangeLog, "", oldPage.SortChildrenBy, data.SortChildrenBy)
				if errorMessage != "" {
					return errorMessage, err
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if hasVote != oldPage.HasVote {
				changeType := core.TurnOnVoteChangeLog
				if !hasVote {
					changeType = core.TurnOffVoteChangeLog
				}
				changeLogId, errorMessage, err := updateChangeLog(changeType, "", strconv.FormatBool(oldPage.HasVote), strconv.FormatBool(hasVote))
				if errorMessage != "" {
					return errorMessage, err
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if data.VoteType != oldPage.VoteType {
				changeLogId, errorMessage, err := updateChangeLog(core.SetVoteTypeChangeLog, "", oldPage.VoteType, data.VoteType)
				if errorMessage != "" {
					return errorMessage, err
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}
			if data.EditGroupId != oldPage.EditGroupId {
				changeLogId, errorMessage, err := updateChangeLog(core.NewEditGroupChangeLog, data.EditGroupId, oldPage.EditGroupId, data.EditGroupId)
				if errorMessage != "" {
					return errorMessage, err
				}
				changeLogIds = append(changeLogIds, changeLogId)
			}

		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(fmt.Sprintf("Transaction failed: %s", errMessage), err)
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
			task.UpdateType = core.PageInfoEditUpdateType
			task.GroupByPageId = data.PageId
			task.ChangeLogId = changeLogId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.StatusOK(nil)
}
