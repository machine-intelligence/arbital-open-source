// group.go contains all the stuff about groups
package core

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/sessions"
)

type Member struct {
	UserId        string `json:"userId"`
	CanAddMembers bool   `json:"canAddMembers"`
	CanAdmin      bool   `json:"canAdmin"`
}

// LoadUserGroupIds loads all the group names this user belongs to.
func LoadUserGroupIds(db *database.DB, u *CurrentUser) error {
	u.GroupIds = make([]string, 0)
	rows := db.NewStatement(`
		SELECT groupId
		FROM groupMembers
		WHERE userId=?`).Query(u.Id)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var groupId string
		err := rows.Scan(&groupId)
		if err != nil {
			return fmt.Errorf("failed to scan for a member: %v", err)
		}
		u.GroupIds = append(u.GroupIds, groupId)
		return nil
	})
	return err
}

// AddUserGroupIdsToPageMap adds user's groups to the page map so we can load them.
func AddUserGroupIdsToPageMap(u *CurrentUser, pageMap map[string]*Page) {
	for _, pageId := range u.GroupIds {
		AddPageIdToMap(pageId, pageMap)
	}
}

// newInternalGroup creates a new group. For internal use only.
func newInternalGroup(tx *database.Tx, groupType string, groupId, userId string, title, alias, clickbait string, isPersonalGroup bool) sessions.Error {
	userGroupStr := "group"
	if isPersonalGroup {
		userGroupStr = "user"
	}
	url := GetEditPageFullUrl("", groupId)
	// NOTE: the tabbing/spacing is really important here since it gets preserved.
	// If we have 4 spaces, in Markdown that will start a code block.
	text := fmt.Sprintf(`
[summary: Nothing here yet.]

Automatically generated page for "%s" %s.
If you are the owner, click [here to edit](%s).`, title, userGroupStr, url)
	// Create new group for the user.
	hashmap := make(database.InsertMap)
	hashmap["pageId"] = groupId
	hashmap["edit"] = 1
	hashmap["title"] = title
	hashmap["text"] = text
	hashmap["clickbait"] = clickbait
	hashmap["creatorId"] = userId
	hashmap["createdAt"] = database.Now()
	hashmap["isLiveEdit"] = true
	statement := tx.DB.NewInsertStatement("pages", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't create a new page", err)
	}

	// Add new group to pageInfos.
	hashmap = make(database.InsertMap)
	hashmap["pageId"] = groupId
	hashmap["alias"] = alias
	hashmap["type"] = groupType
	hashmap["currentEdit"] = 1
	hashmap["maxEdit"] = 1
	hashmap["createdBy"] = userId
	hashmap["createdAt"] = database.Now()
	hashmap["sortChildrenBy"] = AlphabeticalChildSortingOption
	if groupType == GroupPageType {
		hashmap["editGroupId"] = groupId
	}
	statement = tx.DB.NewInsertStatement("pageInfos", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't create a new page", err)
	}

	// Add a summary for the page
	hashmap = make(database.InsertMap)
	hashmap["pageId"] = groupId
	hashmap["name"] = "Summary"
	hashmap["text"] = text
	statement = tx.DB.NewInsertStatement("pageSummaries", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't create a new page summary", err)
	}

	// Add user to the group.
	if groupType == GroupPageType {
		hashmap = make(database.InsertMap)
		hashmap["userId"] = userId
		hashmap["groupId"] = groupId
		hashmap["createdAt"] = database.Now()
		if !isPersonalGroup {
			hashmap["canAddMembers"] = true
			hashmap["canAdmin"] = true
		}
		statement = tx.DB.NewInsertStatement("groupMembers", hashmap).WithTx(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't add user to the group", err)
		}
	}

	// Update elastic search index.
	doc := &elastic.Document{
		PageId:    groupId,
		Type:      groupType,
		Title:     title,
		Clickbait: clickbait,
		Text:      text,
		Alias:     alias,
		CreatorId: userId,
	}
	err := elastic.AddPageToIndex(tx.DB.C, doc)
	if err != nil {
		return sessions.NewError("Failed to update index", err)
	}
	return nil
}

// NewGroup creates a new group and a corresponding page.
func NewGroup(tx *database.Tx, groupId, userId string, title, alias string) sessions.Error {
	return newInternalGroup(tx, GroupPageType, groupId, userId, title, alias, "", false)
}

// NewDomain create a new domain and a corresponding page.
func NewDomain(tx *database.Tx, domainId, userId string, fullName, alias string) sessions.Error {
	return newInternalGroup(tx, DomainPageType, domainId, userId, fullName, alias, "", false)
}

// NewUserGroup create a new person group for a user and the corresponding page.
func NewUserGroup(tx *database.Tx, userId string, fullName, alias string) sessions.Error {
	clickbait := "Automatically generated page for " + fullName
	return newInternalGroup(tx, GroupPageType, userId, userId, fullName, alias, clickbait, true)
}
