// pageHelpers.go contains the page struct as well as helpful functions.
package site

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

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
	newChildPageUpdateType    = "newChildPage"

	// Options for sorting page's children.
	chronologicalChildSortingOption = "chronological"
	alphabeticalChildSortingOption  = "alphabetical"
	likesChildSortingOption         = "likes"

	// Highest karma lock a user can create is equal to their karma * this constant.
	maxKarmaLockFraction = 0.8
)

type page struct {
	// Data loaded from pages table
	PageId         int64 `json:",string"`
	Edit           int
	Type           string
	Title          string
	Text           string
	Summary        string
	Alias          string
	SortChildrenBy string
	HasVote        bool
	Author         dbUser
	CreatedAt      string
	KarmaLock      int
	PrivacyKey     int64 `json:",string"`
	DeletedBy      int64 `json:",string"`
	IsAutosave     bool
	IsSnapshot     bool

	// Data loaded from other tables.
	LastVisit string

	// Computed values.
	InputCount   int //used?
	IsSubscribed bool
	LikeCount    int
	DislikeCount int
	MyLikeValue  int
	LikeScore    int // computed from LikeCount and DislikeCount
	VoteValue    sql.NullFloat64
	VoteCount    int
	MyVoteValue  sql.NullFloat64
	Contexts     []*page //used?
	Links        []*page //used?
	Comments     []*comment
	WasPublished bool // true iff there is an edit that has isCurrentEdit set for this page
	MaxEditEver  int  // highest edit number used for this page for all users
	Parents      []*pagePair
	Children     []*pagePair
	LinkedFrom   []*page
}

// pagePair describes a parent child relationship, which are stored in pagePairs db table.
type pagePair struct {
	// From db.
	Id     int64
	Parent *page
	Child  *page
	UserId int64

	// Populated by the code.
	StillInUse bool // true iff this relationship is still in use
}

// Helpers for sorting page pairs chronologically.
type pagesChronologically []*pagePair

func (a pagesChronologically) Len() int      { return len(a) }
func (a pagesChronologically) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a pagesChronologically) Less(i, j int) bool {
	return a[i].Child.CreatedAt > a[j].Child.CreatedAt
}

// Helpers for sorting page pairs alphabetically.
type pagesAlphabetically []*pagePair

func (a pagesAlphabetically) Len() int      { return len(a) }
func (a pagesAlphabetically) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a pagesAlphabetically) Less(i, j int) bool {
	// Usual string comparison doesn't work well with numbers, e.g. "10abc" comes
	// before "2abc". We fix this by padding all numbers to 20 characters.
	re := regexp.MustCompile("[0-9]+")
	iTitle := re.ReplaceAllStringFunc(a[i].Child.Title, padNumber)
	jTitle := re.ReplaceAllStringFunc(a[j].Child.Title, padNumber)
	return iTitle < jTitle
}
func padNumber(s string) string {
	if len(s) >= 20 {
		return s
	}
	return strings.Repeat("0", 20-len(s)) + s
}

// Helpers for sorting page pairs by votes.
type pagesByLikes []*pagePair

func (a pagesByLikes) Len() int      { return len(a) }
func (a pagesByLikes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a pagesByLikes) Less(i, j int) bool {
	return a[i].Child.LikeScore > a[j].Child.LikeScore
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
	err = pagePtr.loadParents(c, userId)
	if err != nil {
		return nil, err
	}
	return pagePtr, nil
}

// loadFullPage loads and retuns a page. It also loads all the auxillary data like tags.
// If the page couldn't be found, (nil, nil) will be returned.
func loadFullPage(c sessions.Context, pageId int64, userId int64) (*page, error) {
	return loadFullEdit(c, pageId, userId)
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
		SELECT p.pageId,p.edit,p.type,p.title,p.text,p.summary,p.alias,p.sortChildrenBy,p.hasVote,
			p.createdAt,p.karmaLock,p.privacyKey,p.deletedBy,p.isAutosave,p.isSnapshot,
			(SELECT MAX(isCurrentEdit) FROM pages WHERE pageId=%[1]d) AS wasPublished,
			(SELECT max(edit) FROM pages WHERE pageId=%[1]d) AS maxEditEver,
			u.id,u.firstName,u.lastName
		FROM pages AS p
		LEFT JOIN (
			SELECT id,firstName,lastName
			FROM users
		) AS u
		ON p.creatorId=u.Id
		WHERE p.pageId=%[1]d AND %[2]s`, pageId, whereClause)
	exists, err := database.QueryRowSql(c, query, &p.PageId, &p.Edit,
		&p.Type, &p.Title, &p.Text, &p.Summary, &p.Alias, &p.SortChildrenBy, &p.HasVote,
		&p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.IsAutosave, &p.IsSnapshot,
		&p.WasPublished, &p.MaxEditEver, &p.Author.Id, &p.Author.FirstName, &p.Author.LastName)
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

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If the page couldn't be found, (nil, nil) will be returned.
func loadPageByAlias(c sessions.Context, pageAlias string) (*page, error) {
	pageId, err := strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		query := fmt.Sprintf(`
			SELECT pageId
			FROM aliases
			WHERE fullName="%s"`, pageAlias)
		exists, err := database.QueryRowSql(c, query, &pageId)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load an alias: %v", err)
		} else if !exists {
			return nil, nil
		}
	}
	return loadPage(c, pageId)
}

// loadParents loads parents corresponding to this page.
func (p *page) loadParents(c sessions.Context, userId int64) error {
	pageMap := make(map[int64]*page)
	pageMap[p.PageId] = p
	err := loadParents(c, pageMap, userId)
	return err
}

// loadChildren loads children corresponding to this page.
func (p *page) loadChildren(c sessions.Context, userId int64) error {
	pageMap := make(map[int64]*page)
	pageMap[p.PageId] = p
	err := loadChildren(c, pageMap, userId)
	return err
}

// sortChildren sorts page's children.
func (p *page) sortChildren(c sessions.Context) {
	if len(p.Children) <= 0 {
		return
	}
	if p.SortChildrenBy == chronologicalChildSortingOption {
		sort.Sort(pagesChronologically(p.Children))
		c.Debugf("========== %+v", p.Children[0].Child.CreatedAt)
		c.Debugf("========== %+v", p.Children[1].Child.CreatedAt)
		c.Debugf("========== %+v", p.Children[2].Child.CreatedAt)
	} else if p.SortChildrenBy == alphabeticalChildSortingOption {
		sort.Sort(pagesAlphabetically(p.Children))
	} else {
		sort.Sort(pagesByLikes(p.Children))
	}
}

// loadParents loads parents for the given pages.
func loadParents(c sessions.Context, pageMap map[int64]*page, userId int64) error {
	if len(pageMap) <= 0 {
		return nil
	}
	whereClause := "FALSE"
	for id, p := range pageMap {
		whereClause += fmt.Sprintf(" OR (pp.childId=%d AND (pp.childEdit=%d OR pp.userId=%d))", id, p.Edit, userId)
	}
	query := fmt.Sprintf(`
		SELECT pp.id,pp.parentId,pp.childId,pp.userId,p.title,p.alias
		FROM pagePairs AS pp
		LEFT JOIN pages AS p
		ON (p.pageId=pp.parentId AND p.isCurrentEdit)
		WHERE %s`, whereClause)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p pagePair
		p.Parent = &page{}
		p.Child = &page{}
		err := rows.Scan(&p.Id, &p.Parent.PageId, &p.Child.PageId, &p.UserId, &p.Parent.Title, &p.Parent.Alias)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		p.Child = pageMap[p.Child.PageId]
		pageMap[p.Child.PageId].Parents = append(pageMap[p.Child.PageId].Parents, &p)
		return nil
	})
	return err
}

// loadChildren loads children for the given pages.
func loadChildren(c sessions.Context, pageMap map[int64]*page, userId int64) error {
	if len(pageMap) <= 0 {
		return nil
	}
	whereClause := "FALSE"
	for id, _ := range pageMap {
		whereClause += fmt.Sprintf(" OR (pp.parentId=%d)", id)
	}
	query := fmt.Sprintf(`
		SELECT pp.id,pp.parentId,pp.childId,pp.userId,p.title,p.alias,p.createdAt
		FROM pagePairs AS pp
		JOIN pages AS p
		ON (p.pageId=pp.childId AND p.edit=pp.childEdit AND p.isCurrentEdit)
		WHERE %s`, whereClause)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p pagePair
		p.Parent = &page{}
		p.Child = &page{}
		err := rows.Scan(&p.Id, &p.Parent.PageId, &p.Child.PageId, &p.UserId,
			&p.Child.Title, &p.Child.Alias, &p.Child.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		p.Parent = pageMap[p.Parent.PageId]
		pageMap[p.Parent.PageId].Children = append(pageMap[p.Parent.PageId].Children, &p)
		return nil
	})
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
	return fmt.Sprintf("/pages/%s%s", p.Alias, privacyAddon)
}

// getEditPageUrl returns the domain relative url for editing the given page.
func getEditPageUrl(p *page) string {
	var privacyAddon string
	if p.PrivacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey)
	}
	return fmt.Sprintf("/edit/%d%s", p.PageId, privacyAddon)
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
