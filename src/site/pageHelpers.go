// pageHelpers.go contains the page struct as well as helpful functions.
package site

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

const (
	// Various page types we have in our system.
	blogPageType = "blog"
	wikiPageType = "wiki"

	// Various types of updates a user can get.
	topLevelCommentUpdateType = "topLevelComment"
	replyUpdateType           = "reply"
	pageEditUpdateType        = "pageEdit"
	newPageByUserUpdateType   = "newPageByUser"
	newPageWithTagUpdateType  = "newPageWithTag"
	addedTagUpdateType        = "addedTag"

	// Highest karma lock a user can create is equal to their karma * this constant.
	maxKarmaLockFraction = 0.8
)

type page struct {
	// Data loaded from pages table
	PageId     int64 `json:",string"`
	Edit       int
	Type       string
	Title      string
	Text       string
	Summary    string
	Alias      string
	HasVote    bool
	Author     dbUser
	CreatedAt  string
	KarmaLock  int
	PrivacyKey int64 `json:",string"`
	DeletedBy  int64 `json:",string"`
	IsAutosave bool
	IsSnapshot bool

	// Additional data.
	WasPublished bool // true iff there is an edit that has isCurrentEdit set for this page
	Tags         []*tag
}

// loadFullEdit loads and retuns the last edit for the given page id and user id,
// even if it's not live. It also loads all the auxillary data like tags.
// If the page couldn't be found, (nil, nil) will be returned.
func loadFullEdit(c sessions.Context, pageId, userId int64) (*page, error) {
	pagePtr, err := loadEdit(c, pageId, userId)
	if err != nil {
		return nil, err
	}
	if pagePtr == nil {
		return nil, nil
	}
	err = pagePtr.loadTags(c)
	if err != nil {
		return nil, err
	}
	return pagePtr, nil
}

// loadFullPage loads and retuns a page. It also loads all the auxillary data like tags.
// If the page couldn't be found, (nil, nil) will be returned.
func loadFullPage(c sessions.Context, pageId int64) (*page, error) {
	return loadFullEdit(c, pageId, -1)
}

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If userId is given, the last edit of the given pageId will be returned. It
// might be an autosave or a snapshot, and thus not the current live page.
// If the page couldn't be found, (nil, nil) will be returned.
func loadEdit(c sessions.Context, pageId, userId int64) (*page, error) {
	var p page
	whereClause := "p.isCurrentEdit"
	if userId > 0 {
		whereClause = fmt.Sprintf(`
			p.edit=(
				SELECT MAX(edit)
				FROM pages
				WHERE pageId=%d AND (creatorId=%d OR NOT (isSnapshot OR isAutosave))
			)`, pageId, userId)
	}
	// TODO: we often don't need hasCurrentEdit
	query := fmt.Sprintf(`
		SELECT p.pageId,p.edit,p.type,p.title,p.text,p.summary,p.alias,p.hasVote,
			p.createdAt,p.karmaLock,p.privacyKey,p.deletedBy,p.isAutosave,p.isSnapshot,
			(SELECT MAX(isCurrentEdit) FROM pages WHERE pageId=%[1]d) AS wasPublished,
			u.id,u.firstName,u.lastName
		FROM pages AS p
		LEFT JOIN (
			SELECT id,firstName,lastName
			FROM users
		) AS u
		ON p.creatorId=u.Id
		WHERE p.pageId=%[1]d AND %[2]s`, pageId, whereClause)
	exists, err := database.QueryRowSql(c, query, &p.PageId, &p.Edit,
		&p.Type, &p.Title, &p.Text, &p.Summary, &p.Alias, &p.HasVote, &p.CreatedAt,
		&p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.IsAutosave, &p.IsSnapshot,
		&p.WasPublished, &p.Author.Id, &p.Author.FirstName, &p.Author.LastName)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}
	if p.DeletedBy > 0 {
		return &page{PageId: p.PageId, DeletedBy: p.DeletedBy}, nil
	}
	return &p, nil
}

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If the page couldn't be found, (nil, nil) will be returned.
func loadPage(c sessions.Context, pageId int64) (*page, error) {
	return loadEdit(c, pageId, -1)
}

// loadTags loads tags corresponding to this page.
func (p *page) loadTags(c sessions.Context) error {
	pageMap := make(map[int64]*richPage)
	pageMap[p.PageId] = &richPage{page: *p}
	err := loadTags(c, pageMap)
	p.Tags = pageMap[p.PageId].Tags
	return err
}

// getMaxKarmaLock returns the highest possible karma lock a user with the
// given amount of karma can create.
func getMaxKarmaLock(karma int) int {
	return int(float32(karma) * maxKarmaLockFraction)
}

// getPageUrl returns the domain relative url for accessing the given page.
func getPageUrl(p *page) string {
	privacyAddon := ""
	if p.PrivacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey)
	}
	return fmt.Sprintf("/pages/%d%s", p.PageId, privacyAddon)
}

// getEditPageUrl returns the domain relative url for editing the given page.
func getEditPageUrl(p *page) string {
	var privacyAddon string
	if p.PrivacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey)
	}
	return fmt.Sprintf("/pages/edit/%d%s", p.PageId, privacyAddon)
}

// Check if the user can edit this page. -1 = no, 0 = only as admin, 1 = yes
func getEditLevel(p *page, u *user.User) int {
	if p.Type == blogPageType {
		if p.Author.Id == u.Id {
			return 1
		} else {
			return -1
		}
	}
	if u.Karma >= p.KarmaLock {
		return 1
	} else if u.IsAdmin {
		return 0
	}
	return -1
}

// Check if the user can delete this page. -1 = no, 0 = only as admin, 1 = yes
func getDeleteLevel(p *page, u *user.User) int {
	if p.Type == blogPageType {
		if p.Author.Id == u.Id {
			return 1
		} else if u.IsAdmin {
			return 0
		} else {
			return -1
		}
	}
	if u.Karma >= p.KarmaLock {
		return 1
	} else if u.IsAdmin {
		return 0
	}
	return -1
}
