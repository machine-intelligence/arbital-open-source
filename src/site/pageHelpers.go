// pageHelpers.go contains the page struct as well as helpful functions.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
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

	// When encoding a page id into a compressed string, we use this base.
	pageIdEncodeBase = 36
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
	Group          group
	ParentsStr     string
	DeletedBy      int64 `json:",string"`
	IsAutosave     bool
	IsSnapshot     bool

	// Data loaded from other tables.
	LastVisit string

	// Computed values.
	InputCount   int //used?
	IsSubscribed bool
	HasChildren  bool // whether or not this page has children
	HasParents   bool // whether or not this page has parents
	LikeCount    int
	DislikeCount int
	MyLikeValue  int
	LikeScore    int // computed from LikeCount and DislikeCount
	VoteValue    sql.NullFloat64
	VoteCount    int
	MyVoteValue  sql.NullFloat64
	Comments     []*comment
	WasPublished bool // true iff there is an edit that has isCurrentEdit set for this page
	MaxEditEver  int  // highest edit number used for this page for all users
	Parents      []*pagePair
	Children     []*pagePair
	LinkedFrom   []string
}

// pagePair describes a parent child relationship, which are stored in pagePairs db table.
type pagePair struct {
	// From db.
	Id       int64 `json:",string"`
	ParentId int64 `json:",string"`
	ChildId  int64 `json:",string"`
}

// loadPageOptions describes options for loading page(s) from the db
type loadPageOptions struct {
	loadText    bool
	loadSummary bool
}

/*func (a pagesChronologically) Less(i, j int) bool {
	return a[i].Child.CreatedAt > a[j].Child.CreatedAt
}
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
func (a pagesByLikes) Less(i, j int) bool {
	return a[i].Child.LikeScore > a[j].Child.LikeScore
}*/

// processParents converts ParentsStr from this page to the Parents array, and
// populates the given pageMap with the parents.
func (p *page) processParents(c sessions.Context, pageMap map[int64]*page) error {
	if len(p.ParentsStr) <= 0 {
		return nil
	}
	parentIds := strings.Split(p.ParentsStr, ",")
	for _, idStr := range parentIds {
		id, err := strconv.ParseInt(idStr, pageIdEncodeBase, 64)
		if err != nil {
			return err
		}
		pair := pagePair{ParentId: id, ChildId: p.PageId}
		if pageMap != nil {
			newPage, ok := pageMap[pair.ParentId]
			if !ok {
				newPage = &page{PageId: pair.ParentId}
				pageMap[newPage.PageId] = newPage
			}
			newPage.Children = append(newPage.Children, &pair)
		}
		p.Parents = append(p.Parents, &pair)
		p.HasParents = true
	}
	return nil
}

// loadFullEdit loads and retuns the last edit for the given page id and user id,
// even if it's not live. It also loads all the auxillary data like tags.
// If the page couldn't be found, (nil, nil) will be returned.
func loadFullEdit(c sessions.Context, pageId, userId int64) (*page, error) {
	pagePtr, err := loadEdit(c, pageId, userId, loadEditOptions{loadNonliveEdit: true})
	if err != nil {
		return nil, err
	}
	if pagePtr == nil {
		return nil, nil
	}
	return pagePtr, nil
}

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If userId is given, the last edit of the given pageId will be returned. It
// might be an autosave or a snapshot, and thus not the current live page.
// If the page couldn't be found, (nil, nil) will be returned.
type loadEditOptions struct {
	// If true, the last edit will be loaded for the given user, even if it's an
	// autosave or a snapshot.
	loadNonliveEdit bool
}

func loadEdit(c sessions.Context, pageId, userId int64, options loadEditOptions) (*page, error) {
	var p page
	whereClause := "p.isCurrentEdit"
	if options.loadNonliveEdit {
		whereClause = fmt.Sprintf(`
			p.edit=(
				SELECT MAX(edit)
				FROM pages
				WHERE pageId=%d AND (creatorId=%d OR NOT (isSnapshot OR isAutosave))
			)`, pageId, userId)
	}
	// TODO: we often don't need maxEditEver
	query := fmt.Sprintf(`
		SELECT p.pageId,p.edit,p.type,p.title,p.text,p.summary,p.alias,p.sortChildrenBy,p.hasVote,
			p.createdAt,p.karmaLock,p.privacyKey,p.groupName,p.parents,p.deletedBy,p.isAutosave,p.isSnapshot,
			(SELECT max(isCurrentEdit) FROM pages WHERE pageId=%[1]d) AS wasPublished,
			(SELECT max(edit) FROM pages WHERE pageId=%[1]d) AS maxEditEver,
			u.id,u.firstName,u.lastName
		FROM pages AS p
		LEFT JOIN (
			SELECT id,firstName,lastName
			FROM users
		) AS u
		ON p.creatorId=u.Id
		WHERE p.pageId=%[1]d AND %[2]s AND
			(p.groupName="" OR p.groupName IN (SELECT groupName FROM groupMembers WHERE userId=%[3]d))`,
		pageId, whereClause, userId)
	exists, err := database.QueryRowSql(c, query, &p.PageId, &p.Edit,
		&p.Type, &p.Title, &p.Text, &p.Summary, &p.Alias, &p.SortChildrenBy,
		&p.HasVote, &p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.Group.Name,
		&p.ParentsStr, &p.DeletedBy, &p.IsAutosave, &p.IsSnapshot,
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
func loadPage(c sessions.Context, pageId int64, userId int64) (*page, error) {
	return loadEdit(c, pageId, userId, loadEditOptions{})
}

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If the page couldn't be found, (nil, nil) will be returned.
func loadPageByAlias(c sessions.Context, pageAlias string, userId int64) (*page, error) {
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
	return loadPage(c, pageId, userId)
}

// loadPageIds from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func loadPageIds(c sessions.Context, query string, pageMap map[int64]*page) ([]string, error) {
	ids := make([]string, 0, indexPanelLimit)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}

		p, ok := pageMap[pageId]
		if !ok {
			p = &page{PageId: pageId}
			pageMap[pageId] = p
		}
		ids = append(ids, fmt.Sprintf("%d", p.PageId))
		return nil
	})
	return ids, err
}

type loadChildrenIdsOptions struct {
	// If set, the children will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*page
	// Load whether or not each child has children of its own.
	LoadHasChildren bool
}

// loadChildrenIds loads the page ids for all the children of the pages in the given pageMap.
func loadChildrenIds(c sessions.Context, pageMap map[int64]*page, options loadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIdsStr := pageIdsStringFromMap(sourcePageMap)
	newPages := make(map[int64]*page)
	query := fmt.Sprintf(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE parentId IN (%s)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p pagePair
		err := rows.Scan(&p.ParentId, &p.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[p.ChildId]
		if !ok {
			newPage = &page{PageId: p.ChildId}
			pageMap[newPage.PageId] = newPage
			newPages[newPage.PageId] = newPage
		}
		newPage.Parents = append(newPage.Parents, &p)
		sourcePageMap[p.ParentId].Children = append(sourcePageMap[p.ParentId].Children, &p)
		sourcePageMap[p.ParentId].HasChildren = true
		return nil
	})
	if err != nil {
		return err
	}
	if options.LoadHasChildren && len(newPages) > 0 {
		pageIdsStr = pageIdsStringFromMap(newPages)
		query = fmt.Sprintf(`
			SELECT parentId,sum(1)
			FROM pagePairs
			WHERE parentId IN (%s)
			GROUP BY 1`, pageIdsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId int64
			var children int
			err := rows.Scan(&pageId, &children)
			if err != nil {
				return fmt.Errorf("failed to scan for grandchildren: %v", err)
			}
			pageMap[pageId].HasChildren = children > 0
			return nil
		})
	}
	return err
}

type loadParentsIdsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*page
	// Load whether or not each parent has parents of its own.
	LoadHasParents bool
}

// loadParentsIds loads the page ids for all the parents of the pages in the given pageMap.
func loadParentsIds(c sessions.Context, pageMap map[int64]*page, options loadParentsIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIdsStr := pageIdsStringFromMap(sourcePageMap)
	newPages := make(map[int64]*page)
	query := fmt.Sprintf(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE childId IN (%s)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p pagePair
		err := rows.Scan(&p.ParentId, &p.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[p.ParentId]
		if !ok {
			newPage = &page{PageId: p.ParentId}
			pageMap[newPage.PageId] = newPage
			newPages[newPage.PageId] = newPage
		}
		newPage.Children = append(newPage.Children, &p)
		sourcePageMap[p.ChildId].Parents = append(sourcePageMap[p.ChildId].Parents, &p)
		sourcePageMap[p.ChildId].HasParents = true
		return nil
	})
	if err != nil {
		return err
	}
	if options.LoadHasParents && len(newPages) > 0 {
		pageIdsStr = pageIdsStringFromMap(newPages)
		query := fmt.Sprintf(`
			SELECT childId,sum(1)
			FROM pagePairs
			WHERE childId IN (%s)
			GROUP BY 1`, pageIdsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId int64
			var parents int
			err := rows.Scan(&pageId, &parents)
			if err != nil {
				return fmt.Errorf("failed to scan for grandparents: %v", err)
			}
			pageMap[pageId].HasParents = parents > 0
			return nil
		})
	}
	return err
}

// loadPages loads the given pages.
func loadPages(c sessions.Context, pageMap map[int64]*page, userId int64, options loadPageOptions) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := pageIdsStringFromMap(pageMap)
	textSelect := "\"\""
	if options.loadText {
		textSelect = "text"
	}
	summarySelect := "\"\""
	if options.loadSummary {
		summarySelect = "text"
	}
	query := fmt.Sprintf(`
		SELECT pageId,edit,type,creatorId,createdAt,title,%s,karmaLock,privacyKey,
			deletedBy,hasVote,%s,alias,sortChildrenBy,groupName
		FROM pages
		WHERE isCurrentEdit AND deletedBy=0 AND pageId IN (%s) AND
			(groupName="" OR groupName IN (SELECT groupName FROM groupMembers WHERE userId=%d))`,
		textSelect, summarySelect, pageIds, userId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p page
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.Type, &p.Author.Id, &p.CreatedAt, &p.Title,
			&p.Text, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.HasVote,
			&p.Summary, &p.Alias, &p.SortChildrenBy, &p.Group.Name)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}
		if p.DeletedBy <= 0 {
			// We are reduced to this mokery of copying every variable because the page
			// in the pageMap might already have some variables populated.
			// TODO: definitely fix this somehow
			op := pageMap[p.PageId]
			op.Edit = p.Edit
			op.Type = p.Type
			op.Author.Id = p.Author.Id
			op.CreatedAt = p.CreatedAt
			op.Title = p.Title
			op.Text = p.Text
			op.KarmaLock = p.KarmaLock
			op.PrivacyKey = p.PrivacyKey
			op.DeletedBy = p.DeletedBy
			op.HasVote = p.HasVote
			op.Summary = p.Summary
			op.Alias = p.Alias
			op.SortChildrenBy = p.SortChildrenBy
			op.Group.Name = p.Group.Name
		}
		return nil
	})
	return err
}

// pageIdsStringFromMap returns a comma separated string of all pageIds in the given map.
func pageIdsStringFromMap(pageMap map[int64]*page) string {
	var buffer bytes.Buffer
	for id, _ := range pageMap {
		buffer.WriteString(fmt.Sprintf("%d,", id))
	}
	str := buffer.String()
	if len(str) >= 1 {
		str = str[0 : len(str)-1]
	}
	return str
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

// Check if the user can edit this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "blog" = can't perform action because this is a blog page the user doesn't own
// "###" = user doesn't have at least ### karma
func getEditLevel(p *page, u *user.User) string {
	if p.Type == blogPageType {
		if p.Author.Id == u.Id {
			return ""
		} else {
			return "blog"
		}
	}
	karmaReq := p.KarmaLock
	if karmaReq < editPageKarmaReq && p.WasPublished {
		karmaReq = editPageKarmaReq
	}
	if u.Karma < karmaReq {
		if u.IsAdmin {
			return "admin"
		}
		return fmt.Sprintf("%d", karmaReq)
	}
	return ""
}

// Check if the user can delete this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "blog" = can't perform action because this is a blog page the user doesn't own
// "###" = user doesn't have at least ### karma
func getDeleteLevel(p *page, u *user.User) string {
	if p.Type == blogPageType {
		if p.Author.Id == u.Id {
			return ""
		} else if u.IsAdmin {
			return "admin"
		} else {
			return "blog"
		}
	}
	karmaReq := p.KarmaLock
	if karmaReq < deletePageKarmaReq {
		karmaReq = deletePageKarmaReq
	}
	if u.Karma < karmaReq {
		if u.IsAdmin {
			return "admin"
		}
		return fmt.Sprintf("%d", karmaReq)
	}
	return ""
}
