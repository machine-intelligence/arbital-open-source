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
	blogPageType     = "blog"
	wikiPageType     = "wiki"
	commentPageType  = "comment"
	questionPageType = "question"
	answerPageType   = "answer"

	// Various types of updates a user can get.
	topLevelCommentUpdateType = "topLevelComment"
	replyUpdateType           = "reply"
	pageEditUpdateType        = "pageEdit"
	commentEditUpdateType     = "commentEdit"
	newPageByUserUpdateType   = "newPageByUser"
	newChildPageUpdateType    = "newChildPage"

	// Options for sorting page's children.
	chronologicalChildSortingOption = "chronological"
	alphabeticalChildSortingOption  = "alphabetical"
	likesChildSortingOption         = "likes"

	// Options for vote types
	probabilityVoteType = "probability"
	approvalVoteType    = "approval"

	// Highest karma lock a user can create is equal to their karma * this constant.
	maxKarmaLockFraction = 0.8

	// When encoding a page id into a compressed string, we use this base.
	pageIdEncodeBase = 36
)

type vote struct {
	Value     int    `json:"value"`
	UserId    int64  `json:"userId,string"`
	CreatedAt string `json:"createdAt"`
}

type page struct {
	// Data loaded from pages table
	PageId            int64  `json:"pageId,string"`
	Edit              int    `json:"edit"`
	Type              string `json:"type"`
	Title             string `json:"title"`
	Text              string `json:"text"`
	Summary           string `json:"summary"`
	Alias             string `json:"alias"`
	SortChildrenBy    string `json:"sortChildrenBy"`
	HasVote           bool   `json:"hasVote"`
	VoteType          string `json:"voteType"`
	CreatorId         int64  `json:"creatorId,string"`
	CreatedAt         string `json:"createdAt"`
	OriginalCreatedAt string `json:"originalCreatedAt"`
	KarmaLock         int    `json:"karmaLock"`
	PrivacyKey        int64  `json:"privacyKey,string"`
	Group             group  `json:"group"`
	ParentsStr        string `json:"parentsStr"`
	DeletedBy         int64  `json:"deletedBy,string"`
	IsAutosave        bool   `json:"isAutosave"`
	IsSnapshot        bool   `json:"isSnapshot"`
	IsCurrentEdit     bool   `json:"isCurrentEdit"`
	AnchorContext     string `json:"anchorContext"`
	AnchorText        string `json:"anchorText"`
	AnchorOffset      int    `json:"anchorOffset"`

	// Data loaded from other tables.
	LastVisit string `json:"lastVisit"`

	// ===== Computed values. =====
	IsSubscribed bool `json:"isSubscribed"`
	// Whether or not this page has children
	HasChildren bool `json:"hasChildren"`
	// Whether or not this page has parents
	HasParents bool `json:"hasParents"`
	// True iff the user has a work-in-progress draft for this page
	HasDraft     bool `json:"hasDraft"`
	LikeCount    int  `json:"likeCount"`
	DislikeCount int  `json:"dislikeCount"`
	MyLikeValue  int  `json:"myLikeValue"`
	// Computed from LikeCount and DislikeCount
	LikeScore int     `json:"likeScore"`
	Votes     []*vote `json:"votes"`
	// True iff there is an edit that has isCurrentEdit set for this page
	WasPublished bool `json:"wasPublished"`
	// Highest edit number used for this page for all users
	MaxEditEver int `json:"maxEditEver"`
	// We don't allow users to change the vote type once a page has been published
	// with a voteType!="" even once. If it has, this is the vote type it shall
	// always have.
	LockedVoteType string      `json:"lockedVoteType"`
	Parents        []*pagePair `json:"parents"`
	Children       []*pagePair `json:"children"`
	CommentIds     []string    `json:"commentIds"`
	// Page alias/id for a link -> true iff the page is published
	Links        map[string]bool `json:"links"`
	LinkedFrom   []string        `json:"linkedFrom"`
	RedLinkCount int             `json:"redLinkCount"`
	// Set to pageId corresponding to the question the user started creating for this page
	ChildDraftId int64 `json:"childDraftId,string"`
}

// pagePair describes a parent child relationship, which are stored in pagePairs db table.
type pagePair struct {
	// From db.
	Id       int64 `json:"id,string"`
	ParentId int64 `json:"parentId,string"`
	ChildId  int64 `json:"childId,string"`
}

// loadPageOptions describes options for loading page(s) from the db
type loadPageOptions struct {
	loadText    bool
	loadSummary bool
	// If set to true, load snapshots and autosaves, not only current edits
	allowUnpublished bool
}

// processParents converts ParentsStr from this page to the Parents array, and
// populates the given pageMap with the parents.
// pageMap can be nil.
func (p *page) processParents(c sessions.Context, pageMap map[int64]*page) error {
	if len(p.ParentsStr) <= 0 {
		return nil
	}
	p.Parents = nil
	p.HasParents = false
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
	return loadFullEditWithOptions(c, pageId, userId, nil)
}

// loadFullEditWithOptions is just like loadFullEdit, but takes the option
// parameters. Pass nil to use default options.
func loadFullEditWithOptions(c sessions.Context, pageId, userId int64, options *loadEditOptions) (*page, error) {
	if options == nil {
		options = &loadEditOptions{loadNonliveEdit: true}
	}
	pagePtr, err := loadEdit(c, pageId, userId, *options)
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
	// Don't convert loaded parents string into an array of parents
	ignoreParents bool
}

func loadEdit(c sessions.Context, pageId, userId int64, options loadEditOptions) (*page, error) {
	var p page
	whereClause := "p.isCurrentEdit"
	if options.loadNonliveEdit {
		whereClause = fmt.Sprintf(`
			p.edit=(
				SELECT MAX(edit)
				FROM pages
				WHERE pageId=%d AND deletedBy<=0 AND (creatorId=%d OR NOT (isSnapshot OR isAutosave))
			)`, pageId, userId)
	}
	// TODO: we often don't need maxEditEver
	query := fmt.Sprintf(`
		SELECT p.pageId,p.edit,p.type,p.title,p.text,p.summary,p.alias,p.creatorId,
			p.sortChildrenBy,p.hasVote,p.voteType,p.createdAt,p.karmaLock,p.privacyKey,
			p.groupName,p.parents,p.deletedBy,p.isAutosave,p.isSnapshot,p.isCurrentEdit,
			(SELECT max(isCurrentEdit) FROM pages WHERE pageId=%[1]d) AS wasPublished,
			(SELECT max(edit) FROM pages WHERE pageId=%[1]d) AS maxEditEver,
			(SELECT ifnull(max(voteType),"") FROM pages WHERE pageId=%[1]d AND NOT isAutosave AND NOT isSnapshot AND voteType!="") AS lockedVoteType
		FROM pages AS p
		WHERE p.pageId=%[1]d AND %[2]s AND
			(p.groupName="" OR p.groupName IN (SELECT groupName FROM groupMembers WHERE userId=%[3]d))`,
		pageId, whereClause, userId)
	exists, err := database.QueryRowSql(c, query, &p.PageId, &p.Edit,
		&p.Type, &p.Title, &p.Text, &p.Summary, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.Group.Name,
		&p.ParentsStr, &p.DeletedBy, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit,
		&p.WasPublished, &p.MaxEditEver, &p.LockedVoteType)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}
	if p.DeletedBy > 0 {
		return &page{PageId: p.PageId, DeletedBy: p.DeletedBy}, nil
	} else if !options.ignoreParents {
		if err := p.processParents(c, nil); err != nil {
			return nil, fmt.Errorf("Couldn't process parents: %v", err)
		}
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

// loadLinks loads the links for the given page. If checkUnlinked is true, we
// also compute whether or not each link points to a valid page.
func loadLinks(c sessions.Context, pageMap map[int64]*page, checkUnlinked bool) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsStr := pageIdsStringFromMap(pageMap)
	links := make(map[string]bool)
	for _, p := range pageMap {
		p.Links = make(map[string]bool)
	}
	query := fmt.Sprintf(`
		SELECT parentId,childAlias
		FROM links
		WHERE parentId IN (%s)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var parentId int64
		var childAlias string
		err := rows.Scan(&parentId, &childAlias)
		if err != nil {
			return fmt.Errorf("failed to scan for an alias: %v", err)
		}
		links[childAlias] = false
		pageMap[parentId].Links[childAlias] = false
		return nil
	})
	if err != nil {
		return err
	}

	// Create lists for aliases and ids.
	idsList := make([]string, 0, len(links))
	aliasesList := make([]string, 0, len(links))
	for alias, _ := range links {
		_, err := strconv.ParseInt(alias, 10, 64)
		if err != nil {
			aliasesList = append(aliasesList, fmt.Sprintf(`"%s"`, alias))
		} else {
			idsList = append(idsList, alias)
		}
	}

	// Mark which aliases correspond to published pages.
	aliasesStr := strings.Join(aliasesList, ",")
	if len(aliasesStr) > 0 {
		query = fmt.Sprintf(`
			SELECT a.fullName
			FROM (
				SELECT pageId,fullName
				FROM aliases
				WHERE fullName IN (%s)
			) AS a
			LEFT JOIN pages AS p
			ON (a.pageId=p.pageId AND p.isCurrentEdit AND p.deletedBy=0)`, aliasesStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var fullName string
			err := rows.Scan(&fullName)
			if err != nil {
				return fmt.Errorf("failed to scan for an alias: %v", err)
			}
			links[fullName] = true
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Mark which ids correspond to published pages.
	idsStr := strings.Join(idsList, ",")
	if len(idsStr) > 0 {
		query = fmt.Sprintf(`
			SELECT CAST(pageId AS CHAR)
			FROM pages
			WHERE isCurrentEdit AND deletedBy=0 AND pageId IN (%s)`, idsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId string
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan for a page id: %v", err)
			}
			links[pageId] = true
			return nil
		})
		if err != nil {
			return err
		}
	}

	for _, p := range pageMap {
		for alias, _ := range p.Links {
			p.Links[alias] = links[alias]
		}
	}
	return nil
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
		SELECT pp.parentId,pp.childId,p.type
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE parentId IN (%s)
		) AS pp JOIN (
			SELECT pageId,type
			FROM pages
			WHERE isCurrentEdit AND type!="comment"
		) AS p
		ON (p.pageId=pp.childId)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p pagePair
		var childType string
		err := rows.Scan(&p.ParentId, &p.ChildId, &childType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[p.ChildId]
		if !ok {
			newPage = &page{PageId: p.ChildId, Type: childType}
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
		query := fmt.Sprintf(`
			SELECT pp.parentId,sum(1)
			FROM (
				SELECT parentId,childId
				FROM pagePairs
				WHERE parentId IN (%s)
			) AS pp JOIN (
				SELECT pageId
				FROM pages
				WHERE isCurrentEdit AND type!="comment"
			) AS p
			ON (p.pageId=pp.childId)
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

// loadCommentIds loads the page ids for all the comments of the pages in the given pageMap.
func loadCommentIds(c sessions.Context, pageMap map[int64]*page, sourcePageMap map[int64]*page) error {
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIdsStr := pageIdsStringFromMap(sourcePageMap)
	query := fmt.Sprintf(`
		SELECT pp.parentId,pp.childId
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE parentId IN (%s)
		) AS pp JOIN (
			SELECT pageId
			FROM pages
			WHERE isCurrentEdit AND type="comment"
		) AS p
		ON (p.pageId=pp.childId)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p pagePair
		err := rows.Scan(&p.ParentId, &p.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for comments: %v", err)
		}
		newPage, ok := pageMap[p.ChildId]
		if !ok {
			newPage = &page{PageId: p.ChildId, Type: commentPageType}
			pageMap[newPage.PageId] = newPage
		}
		newPage.Parents = append(newPage.Parents, &p)
		sourcePageMap[p.ParentId].CommentIds = append(sourcePageMap[p.ParentId].CommentIds, fmt.Sprintf("%d", p.ChildId))
		return nil
	})
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
	textSelect := "\"\" AS text"
	if options.loadText {
		textSelect = "text"
	}
	summarySelect := "\"\" AS summary"
	if options.loadSummary {
		summarySelect = "summary"
	}
	publishedConstraint := "isCurrentEdit"
	if options.allowUnpublished {
		publishedConstraint = fmt.Sprintf("(isCurrentEdit || creatorId=%d)", userId)
	}
	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT pageId,edit,type,creatorId,createdAt,title,%s,karmaLock,privacyKey,
				deletedBy,hasVote,voteType,%s,alias,sortChildrenBy,groupName,parents,
				isAutosave,isSnapshot,isCurrentEdit,anchorContext,anchorText,anchorOffset
			FROM pages
			WHERE %s AND deletedBy=0 AND pageId IN (%s) AND
				(groupName="" OR groupName IN (SELECT groupName FROM groupMembers WHERE userId=%d))
			ORDER BY edit DESC
		) AS p
		GROUP BY pageId`,
		textSelect, summarySelect, publishedConstraint, pageIds, userId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p page
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.Type, &p.CreatorId, &p.CreatedAt, &p.Title,
			&p.Text, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.HasVote,
			&p.VoteType, &p.Summary, &p.Alias, &p.SortChildrenBy, &p.Group.Name,
			&p.ParentsStr, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit,
			&p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}
		if p.DeletedBy <= 0 {
			// We are reduced to this mokery of copying every variable because the page
			// in the pageMap might already have some variables populated.
			// TODO: definitely fix this somehow. Probably by refactoring how we load pages
			op := pageMap[p.PageId]
			op.Edit = p.Edit
			op.Type = p.Type
			op.CreatorId = p.CreatorId
			op.CreatedAt = p.CreatedAt
			op.Title = p.Title
			op.Text = p.Text
			op.KarmaLock = p.KarmaLock
			op.PrivacyKey = p.PrivacyKey
			op.DeletedBy = p.DeletedBy
			op.HasVote = p.HasVote
			op.VoteType = p.VoteType
			op.Summary = p.Summary
			op.Alias = p.Alias
			op.SortChildrenBy = p.SortChildrenBy
			op.Group.Name = p.Group.Name
			op.ParentsStr = p.ParentsStr
			op.IsAutosave = p.IsAutosave
			op.IsSnapshot = p.IsSnapshot
			op.IsCurrentEdit = p.IsCurrentEdit
			op.AnchorContext = p.AnchorContext
			op.AnchorText = p.AnchorText
			op.AnchorOffset = p.AnchorOffset
			if err := op.processParents(c, nil); err != nil {
				return fmt.Errorf("Couldn't process parents: %v", err)
			}
		}
		return nil
	})
	return err
}

// loadDraftExistence computes for each page whether or not the user has a
// work-in-progress draft for it.
// This only makes sense to call for pages which were loaded for isCurrentEdit=true.
func loadDraftExistence(c sessions.Context, pageMap map[int64]*page, userId int64) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := pageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT pageId,MAX(
				IF((isSnapshot OR isAutosave) AND creatorId=%d AND deletedBy=0 AND
					(groupName="" OR groupName IN (SELECT groupName FROM groupMembers WHERE userId=%d)),
				edit, -1)
			) as myMaxEdit, MAX(IF(isCurrentEdit, edit, -1)) AS currentEdit
		FROM pages
		WHERE pageId IN (%s)
		GROUP BY pageId
		HAVING myMaxEdit > currentEdit`,
		userId, userId, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var blank int
		err := rows.Scan(&pageId, &blank, &blank)
		if err != nil {
			return fmt.Errorf("failed to scan a page id: %v", err)
		}
		pageMap[pageId].HasDraft = true
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
// "comment" = can't perform action because this is a comment page the user doesn't own
// "###" = user doesn't have at least ### karma
func getEditLevel(p *page, u *user.User) string {
	if p.Type == blogPageType || p.Type == commentPageType {
		if p.CreatorId == u.Id {
			return ""
		} else {
			return p.Type
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
	if p.Type == blogPageType || p.Type == commentPageType {
		if p.CreatorId == u.Id {
			return ""
		} else if u.IsAdmin {
			return "admin"
		} else {
			return p.Type
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
