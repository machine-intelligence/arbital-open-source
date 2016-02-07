// editPageInfoHandler.go contains the handler for editing pageInfo data.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// editPageInfoData contains parameters passed in.
type editPageInfoData struct {
	PageId          string `json:""`
	Type            string
	HasVote         bool
	VoteType        string
	SeeGroupId      string `json:""`
	EditGroupId     string `json:""`
	EditKarmaLock   int
	Alias           string // if empty, leave the current one
	SortChildrenBy  string
	IsRequisite     bool
	IndirectTeacher bool
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
	oldPage, err := core.LoadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the old page", err)
	} else if oldPage == nil {
		oldPage = &core.Page{}
	}

	// Error checking.
	// Check the page isn't locked by someone else
	if oldPage.LockedUntil > database.Now() && oldPage.LockedBy != u.Id {
		return pages.HandlerBadRequestFail("Can't change locked page", nil)
	}
	// Check the group settings
	if oldPage.SeeGroupId != data.SeeGroupId && oldPage.WasPublished {
		return pages.HandlerBadRequestFail("Editing this page in incorrect private group", nil)
	}
	if core.IsIdValid(data.SeeGroupId) && !u.IsMemberOfGroup(data.SeeGroupId) {
		return pages.HandlerBadRequestFail("Don't have group permission to EVEN SEE this page", nil)
	}
	if core.IsIdValid(oldPage.SeeGroupId) && !u.IsMemberOfGroup(oldPage.SeeGroupId) {
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
		return pages.HandlerBadRequestFail("Invalid sort children value.", nil)
	}
	if data.VoteType != "" && data.VoteType != core.ProbabilityVoteType && data.VoteType != core.ApprovalVoteType {
		return pages.HandlerBadRequestFail("Invalid vote type value.", nil)
	}
	if data.EditKarmaLock < 0 || data.EditKarmaLock > u.MaxKarmaLock {
		return pages.HandlerBadRequestFail("Karma value out of bounds", nil)
	}

	// Make sure the user has the right permissions to edit this page
	if oldPage.WasPublished {
		editLevel := core.GetEditLevel(oldPage, u)
		if editLevel != "" && editLevel != "admin" {
			return pages.HandlerBadRequestFail("Not enough karma to edit this page.", nil)
		}
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
	if data.Type == core.GroupPageType || data.Type == core.DomainPageType {
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
			WHERE currentEdit>0 AND pageId!=? AND alias=?`).QueryRow(data.PageId, data.Alias)
		exists, err := row.Scan(&existingPageId)
		if err != nil {
			return pages.HandlerErrorFail("Failed on looking for conflicting alias", err)
		} else if exists {
			return pages.HandlerErrorFail(fmt.Sprintf("Alias '%s' is already in use by: %s", data.Alias, existingPageId), nil)
		}
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Update pageInfos
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["alias"] = data.Alias
		hashmap["sortChildrenBy"] = data.SortChildrenBy
		hashmap["hasVote"] = hasVote
		hashmap["voteType"] = data.VoteType
		hashmap["editKarmaLock"] = data.EditKarmaLock
		hashmap["type"] = data.Type
		hashmap["seeGroupId"] = data.SeeGroupId
		hashmap["editGroupId"] = data.EditGroupId
		hashmap["isRequisite"] = data.IsRequisite
		hashmap["indirectTeacher"] = data.IndirectTeacher
		hashmap["lockedUntil"] = core.GetPageQuickLockedUntilTime()
		statement := tx.NewInsertTxStatement("pageInfos", hashmap, hashmap.GetKeys()...)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't update pageInfos", err
		}

		// Update change logs
		if oldPage.WasPublished {
			updateChangeLog := func(changeType string, auxPageId string) (string, error) {
				hashmap = make(database.InsertMap)
				hashmap["pageId"] = data.PageId
				hashmap["userId"] = u.Id
				hashmap["createdAt"] = database.Now()
				hashmap["type"] = changeType
				hashmap["auxPageId"] = auxPageId
				statement = tx.NewInsertTxStatement("changeLogs", hashmap)
				if _, err = statement.Exec(); err != nil {
					return fmt.Sprintf("Couldn't insert new child change log for %s", changeType), err
				}
				return "", nil
			}
			if data.Alias != oldPage.Alias {
				if errorMessage, err := updateChangeLog(core.NewAliasChangeLog, ""); errorMessage != "" {
					return errorMessage, err
				}
			}
			if data.SortChildrenBy != oldPage.SortChildrenBy {
				if errorMessage, err := updateChangeLog(core.NewSortChildrenByChangeLog, ""); errorMessage != "" {
					return errorMessage, err
				}
			}
			if hasVote != oldPage.HasVote {
				changeType := core.TurnOnVoteChangeLog
				if !hasVote {
					changeType = core.TurnOffVoteChangeLog
				}
				if errorMessage, err := updateChangeLog(changeType, ""); errorMessage != "" {
					return errorMessage, err
				}
			}
			if data.VoteType != oldPage.VoteType {
				if errorMessage, err := updateChangeLog(core.SetVoteTypeChangeLog, ""); errorMessage != "" {
					return errorMessage, err
				}
			}
			if data.EditKarmaLock != oldPage.EditKarmaLock {
				if errorMessage, err := updateChangeLog(core.NewEditKarmaLockChangeLog, ""); errorMessage != "" {
					return errorMessage, err
				}
			}
			if data.EditGroupId != oldPage.EditGroupId {
				if errorMessage, err := updateChangeLog(core.NewEditGroupChangeLog, data.EditGroupId); errorMessage != "" {
					return errorMessage, err
				}
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
		doc := &elastic.Document{
			PageId:     data.PageId,
			Type:       data.Type,
			Title:      oldPage.Title,
			Clickbait:  oldPage.Clickbait,
			Text:       oldPage.Text,
			Alias:      data.Alias,
			SeeGroupId: data.SeeGroupId,
			CreatorId:  u.Id,
		}
		err = elastic.AddPageToIndex(c, doc)
		if err != nil {
			c.Errorf("failed to update index: %v", err)
		}
	}

	// Generate "edit" update for users who are subscribed to this page.
	if oldPage.WasPublished {
		var task tasks.NewUpdateTask
		task.UserId = u.Id
		task.GoToPageId = data.PageId
		task.SubscribedToId = data.PageId
		task.UpdateType = core.PageInfoEditUpdateType
		task.GroupByPageId = data.PageId
		if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return pages.StatusOK(nil)
}
