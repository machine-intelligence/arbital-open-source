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
	lensPageType     = "lens"

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
	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageId int64  `json:"pageId,string"`
	Edit   int    `json:"edit"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	// Full text of the page. Not always sent to the FE.
	Text           string `json:"text"`
	Summary        string `json:"summary"`
	Alias          string `json:"alias"`
	SortChildrenBy string `json:"sortChildrenBy"`
	HasVote        bool   `json:"hasVote"`
	VoteType       string `json:"voteType"`
	CreatorId      int64  `json:"creatorId,string"`
	CreatedAt      string `json:"createdAt"`
	KarmaLock      int    `json:"karmaLock"`
	PrivacyKey     int64  `json:"privacyKey,string"`
	GroupId        int64  `json:"groupId,string"`
	ParentsStr     string `json:"parentsStr"`
	DeletedBy      int64  `json:"deletedBy,string"`
	IsAutosave     bool   `json:"isAutosave"`
	IsSnapshot     bool   `json:"isSnapshot"`
	IsCurrentEdit  bool   `json:"isCurrentEdit"`
	AnchorContext  string `json:"anchorContext"`
	AnchorText     string `json:"anchorText"`
	AnchorOffset   int    `json:"anchorOffset"`

	// === Auxillary data. ===
	// For some pages we load additional data.
	IsSubscribed bool `json:"isSubscribed"`
	LikeCount    int  `json:"likeCount"`
	DislikeCount int  `json:"dislikeCount"`
	MyLikeValue  int  `json:"myLikeValue"`
	// Computed from LikeCount and DislikeCount
	LikeScore int `json:"likeScore"`
	// Date when this page was first published.
	OriginalCreatedAt string `json:"originalCreatedAt"`
	// Last time the user visited this page.
	LastVisit string `json:"lastVisit"`
	// True iff the user has a work-in-progress draft for this page
	HasDraft bool `json:"hasDraft"`

	// === Full data. ===
	// For pages that are displayed fully, we load more additional data.
	// True iff there is an edit that has isCurrentEdit set for this page
	WasPublished bool    `json:"wasPublished"`
	Votes        []*vote `json:"votes"`
	// We don't allow users to change the vote type once a page has been published
	// with a voteType!="" even once. If it has, this is the vote type it shall
	// always have.
	LockedVoteType string `json:"lockedVoteType"`
	// Highest edit number used for this page for all users
	MaxEditEver int `json:"maxEditEver"`
	// Map of page aliases/ids -> page title, so we can expand [alias] links
	Links map[string]string `json:"links"`
	//LinkedFrom   []string        `json:"linkedFrom"`
	RedLinkCount int `json:"redLinkCount"`
	// Set to pageId corresponding to the question/answer the user started creating for this page
	ChildDraftId int64 `json:"childDraftId,string"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".

	// Comments.
	CommentIds []string `json:"commentIds"`

	// Whether or not this page has children
	HasChildren bool `json:"hasChildren"`
	// Whether or not this page has parents
	HasParents bool        `json:"hasParents"`
	Parents    []*pagePair `json:"parents"`
	Children   []*pagePair `json:"children"`
	LensIds    []string    `json:"lensIds"`
}

// pagePair describes a parent child relationship, which are stored in pagePairs db table.
type pagePair struct {
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
			p.groupId,p.parents,p.deletedBy,p.isAutosave,p.isSnapshot,p.isCurrentEdit,
			(SELECT max(isCurrentEdit) FROM pages WHERE pageId=%[1]d) AS wasPublished,
			(SELECT max(edit) FROM pages WHERE pageId=%[1]d) AS maxEditEver,
			(SELECT ifnull(max(voteType),"") FROM pages WHERE pageId=%[1]d AND NOT isAutosave AND NOT isSnapshot AND voteType!="") AS lockedVoteType
		FROM pages AS p
		WHERE p.pageId=%[1]d AND %[2]s AND
			(p.groupId=0 OR p.groupId IN (SELECT groupId FROM groupMembers WHERE userId=%[3]d))`,
		pageId, whereClause, userId)
	exists, err := database.QueryRowSql(c, query, &p.PageId, &p.Edit,
		&p.Type, &p.Title, &p.Text, &p.Summary, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.GroupId,
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

// loadChildDraft loads a potentially existing draft for the given page. If it's
// loaded, it'll be added to the give map.
func loadChildDraft(c sessions.Context, userId int64, p *page, pageMap map[int64]*page) error {
	if p.Type != questionPageType {
		// Load potential question draft.
		query := fmt.Sprintf(`
			SELECT pageId
			FROM pages
			WHERE type="question" AND creatorId=%d AND deletedBy<=0 AND parents REGEXP "(^|,)%s($|,)"
			GROUP BY pageId
			HAVING SUM(isCurrentEdit)<=0`, userId, strconv.FormatInt(p.PageId, pageIdEncodeBase))
		_, err := database.QueryRowSql(c, query, &p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load question draft: %v", err)
		}
	} else {
		// Load potential answer draft.
		query := fmt.Sprintf(`
			SELECT pageId
			FROM pages
			WHERE type="answer" AND creatorId=%d AND deletedBy<=0 AND parents REGEXP "(^|,)%s($|,)"
			GROUP BY pageId
			HAVING SUM(isCurrentEdit)<=0`, userId, strconv.FormatInt(p.PageId, pageIdEncodeBase))
		_, err := database.QueryRowSql(c, query, &p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load answer draft id: %v", err)
		}
		if p.ChildDraftId > 0 {
			p, err := loadFullEdit(c, p.ChildDraftId, userId)
			if err != nil {
				return fmt.Errorf("Couldn't load answer draft: %v", err)
			}
			pageMap[p.PageId] = p
		}
	}
	return nil
}

// loadLinks loads the links for the given page.
func loadLinks(c sessions.Context, fullPageMap map[int64]*page) error {
	if len(fullPageMap) <= 0 {
		return nil
	}
	// Filter out pages that don't have text.
	pageMap := make(map[int64]*page)
	for id, p := range fullPageMap {
		if p.Text != "" {
			pageMap[id] = p
		}
	}
	if len(pageMap) <= 0 {
		return nil
	}
	// List of all aliases we need to get titles for
	aliasesList := make([]string, 0, 0)
	// Map of each page alias to a list of pages which have it as a link.
	linkMap := make(map[string]string)

	// Load all links.
	pageIdsStr := pageIdsStringFromMap(pageMap)
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
		aliasesList = append(aliasesList, fmt.Sprintf(`"%s"`, childAlias))
		if pageMap[parentId].Links == nil {
			pageMap[parentId].Links = make(map[string]string)
		}
		pageMap[parentId].Links[childAlias] = ""
		return nil
	})
	if err != nil {
		return err
	}

	// Get the page titles for all the links.
	aliasesStr := strings.Join(aliasesList, ",")
	if len(aliasesStr) > 0 {
		query = fmt.Sprintf(`
			SELECT alias,title
			FROM pages
			WHERE isCurrentEdit AND deletedBy=0 AND alias IN (%s)`, aliasesStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var alias, title string
			err := rows.Scan(&alias, &title)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			linkMap[alias] = title
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Set the links for all pages.
	for _, p := range pageMap {
		for alias, _ := range p.Links {
			p.Links[alias] = linkMap[alias]
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

		parent := sourcePageMap[p.ParentId]
		if newPage.Type == lensPageType {
			parent.LensIds = append(parent.LensIds, fmt.Sprintf("%d", newPage.PageId))
		} else {
			parent.Children = append(parent.Children, &p)
			parent.HasChildren = true
		}
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
				deletedBy,hasVote,voteType,%s,alias,sortChildrenBy,groupId,parents,
				isAutosave,isSnapshot,isCurrentEdit,anchorContext,anchorText,anchorOffset
			FROM pages
			WHERE %s AND deletedBy=0 AND pageId IN (%s) AND
				(groupId=0 OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=%d))
			ORDER BY edit DESC
		) AS p
		GROUP BY pageId`,
		textSelect, summarySelect, publishedConstraint, pageIds, userId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p page
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.Type, &p.CreatorId, &p.CreatedAt, &p.Title,
			&p.Text, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.HasVote,
			&p.VoteType, &p.Summary, &p.Alias, &p.SortChildrenBy, &p.GroupId,
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
			op.GroupId = p.GroupId
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
func loadDraftExistence(c sessions.Context, userId int64, pageMap map[int64]*page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := pageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT pageId,MAX(
				IF((isSnapshot OR isAutosave) AND creatorId=%d AND deletedBy=0 AND
					(groupId=0 OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=%d)),
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

// loadLastVisits loads lastVisit variable for each page.
func loadLastVisits(c sessions.Context, currentUserId int64, pageMap map[int64]*page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsStr := pageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT pageId,max(createdAt)
		FROM visits
		WHERE userId=%d AND pageId IN (%s)
		GROUP BY 1`,
		currentUserId, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var createdAt string
		err := rows.Scan(&pageId, &createdAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a comment like: %v", err)
		}
		pageMap[pageId].LastVisit = createdAt
		return nil
	})
	return err
}

// loadSubscriptions loads subscription statuses corresponding to the given
// pages, and then updates the given maps.
func loadSubscriptions(c sessions.Context, currentUserId int64, pageMap map[int64]*page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := pageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT toPageId
		FROM subscriptions
		WHERE userId=%d AND toPageId IN (%s)`,
		currentUserId, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var toPageId int64
		err := rows.Scan(&toPageId)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		pageMap[toPageId].IsSubscribed = true
		return nil
	})
	return err
}

// loadUserSubscriptions loads subscription statuses corresponding to the given
// users, and then updates the given map.
func loadUserSubscriptions(c sessions.Context, currentUserId int64, userMap map[int64]*dbUser) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIds := pageIdsStringFromUserMap(userMap)
	query := fmt.Sprintf(`
		SELECT toUserId
		FROM subscriptions
		WHERE userId=%d AND toUserId IN (%s)`,
		currentUserId, userIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var toUserId int64
		err := rows.Scan(&toUserId)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		userMap[toUserId].IsSubscribed = true
		return nil
	})
	return err
}

type loadAuxPageDataOptions struct {
	// If set, pretend that we last visited all the pages on this date.
	// Used when we refresh the page, but don't want to erase the new/updated stars just yet.
	ForcedLastVisit string
}

// loadAuxPageData loads the auxillary page data for the given pages.
func loadAuxPageData(c sessions.Context, userId int64, pageMap map[int64]*page, options *loadAuxPageDataOptions) error {
	if options == nil {
		options = &loadAuxPageDataOptions{}
	}

	// Load likes
	err := loadLikes(c, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load likes: %v", err)
	}

	// Load all the subscription statuses.
	if userId > 0 {
		err = loadSubscriptions(c, userId, pageMap)
		if err != nil {
			return fmt.Errorf("Couldn't load subscriptions: %v", err)
		}
	}

	// Load whether or not pages have drafts.
	err = loadDraftExistence(c, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load draft existence: %v", err)
	}

	// Load original creation date.
	if len(pageMap) > 0 {
		pageIdsStr := pageIdsStringFromMap(pageMap)
		query := fmt.Sprintf(`
			SELECT pageId,MIN(createdAt)
			FROM pages
			WHERE pageId IN (%s) AND NOT isAutosave AND NOT isSnapshot
			GROUP BY 1`, pageIdsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId int64
			var originalCreatedAt string
			err := rows.Scan(&pageId, &originalCreatedAt)
			if err != nil {
				return fmt.Errorf("failed to scan for original createdAt: %v", err)
			}
			pageMap[pageId].OriginalCreatedAt = originalCreatedAt
			return nil
		})
		if err != nil {
			return fmt.Errorf("Couldn't load original createdAt: %v", err)
		}
	}

	// Load last visit time.
	err = loadLastVisits(c, userId, pageMap)
	if err != nil {
		return fmt.Errorf("error while fetching a visit: %v", err)
	}
	if options.ForcedLastVisit != "" {
		// Reset the last visit date for all the pages we actually visited
		for _, p := range pageMap {
			if p.LastVisit > options.ForcedLastVisit {
				p.LastVisit = options.ForcedLastVisit
			}
		}
	}

	return nil
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

// pageIdsStringFromUserMap returns a comma separated string of all userIds in the given map.
func pageIdsStringFromUserMap(userMap map[int64]*dbUser) string {
	var buffer bytes.Buffer
	for id, _ := range userMap {
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
