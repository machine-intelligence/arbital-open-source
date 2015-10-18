// page.go contains all the page stuff
package core

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"zanaduu3/src/database"
)

const (
	// Various page types we have in our system.
	WikiPageType     = "wiki"
	CommentPageType  = "comment"
	QuestionPageType = "question"
	AnswerPageType   = "answer"
	LensPageType     = "lens"
	DeletedPageType  = "deleted"

	// Various types of page connections.
	ParentPagePairType      = "parent"
	TagPagePairType         = "tag"
	RequirementPagePairType = "requirement"

	// Various types of updates a user can get.
	TopLevelCommentUpdateType = "topLevelComment"
	ReplyUpdateType           = "reply"
	PageEditUpdateType        = "pageEdit"
	CommentEditUpdateType     = "commentEdit"
	NewPageByUserUpdateType   = "newPageByUser"
	NewChildPageUpdateType    = "newChildPage"
	AtMentionUpdateType       = "atMention"

	// Options for sorting page's children.
	ChronologicalChildSortingOption = "chronological"
	AlphabeticalChildSortingOption  = "alphabetical"
	LikesChildSortingOption         = "likes"

	// Options for vote types
	ProbabilityVoteType = "probability"
	ApprovalVoteType    = "approval"

	// When encoding a page id into a compressed string, we use this base.
	PageIdEncodeBase = 36

	// How long the page lock lasts
	PageLockDuration = 30 * 60 // in seconds

	// String that can be used inside a regexp to match an a page alias or id
	AliasRegexpStr = "[A-Za-z0-9.]+"
)

var (
	// Regexp that strictly matches an alias
	StrictAliasRegexp = regexp.MustCompile("^[0-9A-Za-z]*[A-Za-z][0-9A-Za-z]*$")
)

type Vote struct {
	Value     int    `json:"value"`
	UserId    int64  `json:"userId,string"`
	CreatedAt string `json:"createdAt"`
}

type Page struct {
	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageId    int64  `json:"pageId,string"`
	Edit      int    `json:"edit"`
	PrevEdit  int    `json:"prevEdit"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Clickbait string `json:"clickbait"`
	// Full text of the page. Not always sent to the FE.
	Text string `json:"text"`
	// Meta text of the page. Not always sent to the FE.
	MetaText       string `json:"metaText"`
	TextLength     int    `json:"textLength"`
	Summary        string `json:"summary"`
	Alias          string `json:"alias"`
	SortChildrenBy string `json:"sortChildrenBy"`
	HasVote        bool   `json:"hasVote"`
	VoteType       string `json:"voteType"`
	CreatorId      int64  `json:"creatorId,string"`
	CreatedAt      string `json:"createdAt"`
	EditKarmaLock  int    `json:"editKarmaLock"`
	SeeGroupId     int64  `json:"seeGroupId,string"`
	ParentsStr     string `json:"parentsStr"`
	IsAutosave     bool   `json:"isAutosave"`
	IsSnapshot     bool   `json:"isSnapshot"`
	IsCurrentEdit  bool   `json:"isCurrentEdit"`
	IsMinorEdit    bool   `json:"isMinorEdit"`
	TodoCount      int    `json:"todoCount"`
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
	Votes        []*Vote `json:"votes"`
	// We don't allow users to change the vote type once a page has been published
	// with a voteType!="" even once. If it has, this is the vote type it shall
	// always have.
	LockedVoteType string `json:"lockedVoteType"`
	// Highest edit number used for this page for all users
	MaxEditEver int `json:"maxEditEver"`
	// Highest edit number of an autosave this user created
	MyLastAutosaveEdit sql.NullInt64 `json:"myLastAutosaveEdit"`
	RedLinkCount       int           `json:"redLinkCount"`
	// Set to pageId corresponding to the question/answer the user started creating for this page
	ChildDraftId int64 `json:"childDraftId,string"`
	// Page is locked by this user
	LockedBy int64 `json:"lockedBy,string"`
	// User has the page lock until this time
	LockedUntil string `json:"lockedUntil"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".

	// Subpages.
	SubpageIds     []string `json:"subpageIds"`
	TaggedAsIds    []string `json:"taggedAsIds"`
	RelatedIds     []string `json:"relatedIds"`
	RequirementIds []string `json:"requirementIds"`

	// Domains.
	DomainIds []string `json:"domainIds"`

	// Edit history map. Edit number -> page edit
	EditHistoryMap map[string]*Page `json:"editHistoryMap"`

	// Whether or not this page has children
	HasChildren bool `json:"hasChildren"`
	// Whether or not this page has parents
	HasParents bool        `json:"hasParents"`
	Parents    []*PagePair `json:"parents"`
	Children   []*PagePair `json:"children"`
	LensIds    []string    `json:"lensIds"`
}

// PagePair describes a parent child relationship, which are stored in pagePairs db table.
type PagePair struct {
	Id       int64  `json:"id,string"`
	ParentId int64  `json:"parentId,string"`
	ChildId  int64  `json:"childId,string"`
	Type     string `json:"type"`
}

// Mastery is a page you should have mastered before you can understand another page.
type Mastery struct {
	PageId        int64  `json:"pageId,string"`
	Has           bool   `json:"has"`
	UpdatedAt     string `json:"updatedAt"`
	IsManuallySet bool   `json:"isManuallySet"`
}

// LoadPageOptions describes options for loading page(s) from the db
type LoadPageOptions struct {
	LoadText    bool
	LoadSummary bool
	// If set to true, load snapshots and autosaves, not only current edits
	AllowUnpublished bool
}

// LoadPages loads the given pages.
func LoadPages(db *database.DB, pageMap map[int64]*Page, userId int64, options *LoadPageOptions) error {
	if options == nil {
		options = &LoadPageOptions{}
	}
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	textSelect := `"" AS text`
	if options.LoadText {
		textSelect = "text"
	}
	summarySelect := `"" AS summary`
	if options.LoadSummary {
		summarySelect = "summary"
	}
	publishedConstraint := database.NewQuery("isCurrentEdit")
	if options.AllowUnpublished {
		publishedConstraint = database.NewQuery("(isCurrentEdit || creatorId=?)", userId)
	}
	statement := database.NewQuery(`
		SELECT * FROM (
			SELECT pageId,edit,prevEdit,type,creatorId,createdAt,title,clickbait,` + textSelect + `,
				length(text),metaText,editKarmaLock,hasVote,voteType,` + summarySelect + `,
				alias,sortChildrenBy,seeGroupId,parents,isAutosave,isSnapshot,isCurrentEdit,isMinorEdit,
				todoCount,anchorContext,anchorText,anchorOffset
			FROM pages
			WHERE`).AddPart(publishedConstraint).Add(`AND pageId IN`).AddArgsGroup(pageIds).Add(`
			AND (seeGroupId=0 OR seeGroupId IN (SELECT id FROM groups WHERE isVisible) OR seeGroupId IN (SELECT groupId FROM groupMembers WHERE userId=`).AddArg(userId).Add(`))
			ORDER BY edit DESC
		) AS p
		GROUP BY pageId`).ToStatement(db)
	rows := statement.Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p Page
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.PrevEdit, &p.Type, &p.CreatorId, &p.CreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.EditKarmaLock, &p.HasVote,
			&p.VoteType, &p.Summary, &p.Alias, &p.SortChildrenBy, &p.SeeGroupId,
			&p.ParentsStr, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
			&p.TodoCount, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}

		// We are reduced to this mokery of copying every variable because the page
		// in the pageMap might already have some variables populated.
		// TODO: definitely fix this somehow. Probably by refactoring how we load pages
		op := pageMap[p.PageId]
		op.Edit = p.Edit
		op.PrevEdit = p.PrevEdit
		op.Type = p.Type
		op.CreatorId = p.CreatorId
		op.CreatedAt = p.CreatedAt
		op.Title = p.Title
		op.Clickbait = p.Clickbait
		op.Text = p.Text
		op.MetaText = p.MetaText
		op.TextLength = p.TextLength
		op.EditKarmaLock = p.EditKarmaLock
		op.HasVote = p.HasVote
		op.VoteType = p.VoteType
		op.Summary = p.Summary
		op.Alias = p.Alias
		op.SortChildrenBy = p.SortChildrenBy
		op.SeeGroupId = p.SeeGroupId
		op.ParentsStr = p.ParentsStr
		op.IsAutosave = p.IsAutosave
		op.IsSnapshot = p.IsSnapshot
		op.IsCurrentEdit = p.IsCurrentEdit
		op.IsMinorEdit = p.IsMinorEdit
		op.TodoCount = p.TodoCount
		op.AnchorContext = p.AnchorContext
		op.AnchorText = p.AnchorText
		op.AnchorOffset = p.AnchorOffset
		if err := op.ProcessParents(db.C, nil); err != nil {
			return fmt.Errorf("Couldn't process parents: %v", err)
		}
		return nil
	})
	for _, p := range pageMap {
		if p.Type == "" {
			delete(pageMap, p.PageId)
		}
	}
	return err
}

// LoadEditHistory loads the edit history for the given page.
func LoadEditHistory(db *database.DB, page *Page, userId int64) error {
	editHistoryMap := make(map[int]*Page)
	rows := db.NewStatement(`
		SELECT pageId,edit,prevEdit,creatorId,createdAt,isAutosave,
			isSnapshot,isCurrentEdit,isMinorEdit,title,clickbait,length(text)
		FROM pages
		WHERE pageId=? AND (creatorId=? OR NOT isAutosave)
		ORDER BY edit`).Query(page.PageId, userId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p Page
		err := rows.Scan(&p.PageId, &p.Edit, &p.PrevEdit, &p.CreatorId, &p.CreatedAt,
			&p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
			&p.Title, &p.Clickbait, &p.TextLength)
		if err != nil {
			return fmt.Errorf("failed to scan a edit history page: %v", err)
		}
		editHistoryMap[p.Edit] = &p
		return nil
	})
	if err != nil {
		return err
	}

	// We have loaded some snapshots that might not be our own, but they could be
	// relevant to presenting the history data accurately. So we have to clean
	// out those edits, while preserving the ancestry.
	// First, just copy over edits we want to keep.
	page.EditHistoryMap = make(map[string]*Page)
	for editNum, edit := range editHistoryMap {
		if edit.CreatorId == userId || (!edit.IsSnapshot && !edit.IsAutosave) {
			// Only add public edits or our snapshots to the map
			page.EditHistoryMap[fmt.Sprintf("%d", editNum)] = edit
		}
	}
	// Second, fix the ancestry.
	for _, edit := range page.EditHistoryMap {
		prevEditNum := fmt.Sprintf("%d", edit.PrevEdit)
		if prevEditNum == "0" {
			continue
		}
		_, ok := page.EditHistoryMap[prevEditNum]
		for !ok {
			prevEdit64, _ := strconv.ParseInt(prevEditNum, 10, 64)
			prevEditNum = fmt.Sprintf("%d", editHistoryMap[int(prevEdit64)].PrevEdit)
			if prevEditNum == "0" {
				break
			}
			_, ok = page.EditHistoryMap[prevEditNum]
		}
		if !ok {
			prevEditNum = "0"
		}
		prevEdit64, _ := strconv.ParseInt(prevEditNum, 10, 64)
		edit.PrevEdit = int(prevEdit64)
	}

	return nil
}

type LoadEditOptions struct {
	// If true, the last edit will be loaded for the given user, even if it's an
	// autosave or a snapshot.
	LoadNonliveEdit bool
	// Don't convert loaded parents string into an array of parents
	IgnoreParents bool

	// If set, we'll load this edit of the page
	LoadSpecificEdit int
	// If set, we'll only load from edits less than this
	LoadEditWithLimit int
	// If set, we'll only load from edits with createdAt timestamp before this
	CreatedAtLimit string
}

// LoadFullEdit loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If userId is given, the last edit of the given pageId will be returned. It
// might be an autosave or a snapshot, and thus not the current live page.
// If the page couldn't be found, (nil, nil) will be returned.
func LoadFullEdit(db *database.DB, pageId, userId int64, options *LoadEditOptions) (*Page, error) {
	if options == nil {
		options = &LoadEditOptions{}
	}
	var p Page

	whereClause := database.NewQuery("p.isCurrentEdit")
	if options.LoadSpecificEdit > 0 {
		whereClause = database.NewQuery("p.edit=?", options.LoadSpecificEdit)
	} else if options.LoadNonliveEdit {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT MAX(edit)
				FROM pages
				WHERE pageId=? AND (creatorId=? OR NOT (isSnapshot OR isAutosave))
			)`, pageId, userId)
	} else if options.LoadEditWithLimit > 0 {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND edit<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.LoadEditWithLimit)
	} else if options.CreatedAtLimit != "" {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND createdAt<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.CreatedAtLimit)
	}
	statement := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,p.type,p.title,p.clickbait,p.text,p.metaText,
			p.summary,p.alias,p.creatorId,p.sortChildrenBy,p.hasVote,p.voteType,
			p.createdAt,p.editKarmaLock,p.seeGroupId,p.parents,
			p.isAutosave,p.isSnapshot,p.isCurrentEdit,p.isMinorEdit,
			p.todoCount,p.anchorContext,p.anchorText,p.anchorOffset,
			i.currentEdit>0,i.maxEdit,i.lockedBy,i.lockedUntil,
			(SELECT ifnull(max(voteType),"") FROM pages WHERE pageId=?`, pageId).Add(`
					AND NOT isAutosave AND NOT isSnapshot AND voteType!="") AS lockedVoteType
		FROM pages AS p
		JOIN (
			SELECT *
			FROM pageInfos
			WHERE pageId=?`, pageId).Add(`
		) AS i
		ON (p.pageId=i.pageId)
		WHERE`).AddPart(whereClause).Add(`AND
			(p.seeGroupId=0 OR p.seeGroupId IN (SELECT id FROM groups WHERE isVisible) OR p.seeGroupId IN (SELECT groupId FROM groupMembers WHERE userId=?))`, userId).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageId, &p.Edit, &p.PrevEdit, &p.Type, &p.Title, &p.Clickbait,
		&p.Text, &p.MetaText, &p.Summary, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.EditKarmaLock, &p.SeeGroupId,
		&p.ParentsStr, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
		&p.TodoCount, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset, &p.WasPublished,
		&p.MaxEditEver, &p.LockedBy, &p.LockedUntil, &p.LockedVoteType)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}

	p.TextLength = len(p.Text)
	if !options.IgnoreParents {
		if err := p.ProcessParents(db.C, nil); err != nil {
			return nil, fmt.Errorf("Couldn't process parents: %v", err)
		}
	}
	return &p, nil
}

// LoadPageIds from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func LoadPageIds(rows *database.Rows, pageMap map[int64]*Page) ([]string, error) {
	ids := make([]string, 0)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId int64
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}

		p, ok := pageMap[pageId]
		if !ok {
			p = &Page{PageId: pageId}
			pageMap[pageId] = p
		}
		ids = append(ids, fmt.Sprintf("%d", p.PageId))
		return nil
	})
	return ids, err
}

// LoadChildDraft loads a potentially existing draft for the given page. If it's
// loaded, it'll be added to the give map.
func LoadChildDraft(db *database.DB, userId int64, p *Page, pageMap map[int64]*Page) error {
	parentRegexp := fmt.Sprintf("(^|,)%s($|,)", strconv.FormatInt(p.PageId, PageIdEncodeBase))
	if p.Type != QuestionPageType {
		// Load potential question draft.
		row := db.NewStatement(`
			SELECT pageId
			FROM (
				SELECT pageId,creatorId
				FROM pages
				WHERE type="question" AND parents REGEXP ?
				GROUP BY pageId
				HAVING SUM(isCurrentEdit)<=0
			) AS p
			WHERE creatorId=?
			LIMIT 1`).QueryRow(parentRegexp, userId)
		_, err := row.Scan(&p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load question draft: %v", err)
		}
	} else {
		// Load potential answer draft.
		row := db.NewStatement(`
			SELECT pageId
			FROM (
				SELECT pageId,creatorId
				FROM pages
				WHERE type="answer" AND parents REGEXP ?
				GROUP BY pageId
				HAVING SUM(isCurrentEdit)<=0
			) AS p
			WHERE creatorId=?
			LIMIT 1`).QueryRow(parentRegexp, userId)
		_, err := row.Scan(&p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load answer draft id: %v", err)
		}
		if p.ChildDraftId > 0 {
			p, err := LoadFullEdit(db, p.ChildDraftId, userId, &LoadEditOptions{LoadNonliveEdit: true})
			if err != nil {
				return fmt.Errorf("Couldn't load answer draft: %v", err)
			}
			pageMap[p.PageId] = p
		}
	}
	return nil
}

// LoadLikes loads likes corresponding to the given pages and updates the pages.
func LoadLikes(db *database.DB, currentUserId int64, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT userId,pageId,value
		FROM (
			SELECT *
			FROM likes
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId int64
		var pageId int64
		var value int
		err := rows.Scan(&userId, &pageId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a like: %v", err)
		}
		page := pageMap[pageId]
		if value > 0 {
			if page.LikeCount >= page.DislikeCount {
				page.LikeScore++
			} else {
				page.LikeScore += 2
			}
			page.LikeCount++
		} else if value < 0 {
			if page.DislikeCount >= page.LikeCount {
				page.LikeScore--
			}
			page.DislikeCount++
		}
		if userId == currentUserId {
			page.MyLikeValue = value
		}
		return nil
	})
	return err
}

// LoadVotes loads probability votes corresponding to the given pages and updates the pages.
func LoadVotes(db *database.DB, currentUserId int64, pageMap map[int64]*Page, usersMap map[int64]*User) error {
	pageIds := PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT userId,pageId,value,createdAt
		FROM (
			SELECT *
			FROM votes
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var v Vote
		var pageId int64
		err := rows.Scan(&v.UserId, &pageId, &v.Value, &v.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a vote: %v", err)
		}
		if v.Value == 0 {
			return nil
		}
		page := pageMap[pageId]
		if page.Votes == nil {
			page.Votes = make([]*Vote, 0, 0)
		}
		page.Votes = append(page.Votes, &v)
		if _, ok := usersMap[v.UserId]; !ok {
			usersMap[v.UserId] = &User{Id: v.UserId}
		}
		return nil
	})
	return err
}

// LoadRedLinkCount loads the number of red links for a page.
func LoadRedLinkCount(db *database.DB, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsList := PageIdsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT l.parentId,SUM(ISNULL(p.pageId))
		FROM pages AS p
		RIGHT JOIN links AS l
		ON ((p.pageId=l.childAlias OR p.alias=l.childAlias) AND p.isCurrentEdit AND p.type!="")
		WHERE l.parentId IN`).AddArgsGroup(pageIdsList).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId int64
		var count int
		err := rows.Scan(&parentId, &count)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		pageMap[parentId].RedLinkCount = count
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error scanning for pageId links: %v", err)
	}

	return nil
}

type LoadLinksOptions struct {
	// If set, we'll only load links for the pages with these ids
	FromPageMap map[int64]*Page
}

// LoadLinks loads the links for the given pages, and adds them to the pageMap.
func LoadLinks(db *database.DB, pageMap map[int64]*Page, options *LoadLinksOptions) error {
	if options == nil {
		options = &LoadLinksOptions{}
	}

	sourceMap := options.FromPageMap
	if sourceMap == nil {
		sourceMap = pageMap
	}

	pageIds := make([]interface{}, 0, len(sourceMap))
	for id, _ := range sourceMap {
		pageIds = append(pageIds, id)
	}
	if len(pageIds) <= 0 {
		return nil
	}

	// List of all aliases we'll need to convert to pageIds
	aliasesList := make([]interface{}, 0)

	// Load all links.
	rows := db.NewStatement(`
		SELECT parentId,childAlias
		FROM links
		WHERE parentId IN ` + database.InArgsPlaceholder(len(pageIds))).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId int64
		var childAlias string
		err := rows.Scan(&parentId, &childAlias)
		if err != nil {
			return fmt.Errorf("failed to scan for a link: %v", err)
		}
		if pageId, err := strconv.ParseInt(childAlias, 10, 64); err == nil {
			if _, ok := pageMap[pageId]; !ok {
				pageMap[pageId] = &Page{PageId: pageId}
			}
		} else {
			aliasesList = append(aliasesList, childAlias)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Convert all page aliases to page ids.
	if len(aliasesList) > 0 {
		rows = database.NewQuery(`
			SELECT pageId
			FROM pages
			WHERE isCurrentEdit AND alias IN`).AddArgsGroup(aliasesList).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan for a page: %v", err)
			}
			if _, ok := pageMap[pageId]; !ok {
				pageMap[pageId] = &Page{PageId: pageId}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type LoadChildrenIdsOptions struct {
	// If set, the children will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*Page
	// Load whether or not each child has children of its own.
	LoadHasChildren bool
}

// LoadChildrenIds loads the page ids for all the children of the pages in the given pageMap.
func LoadChildrenIds(db *database.DB, pageMap map[int64]*Page, options LoadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	newPages := make(map[int64]*Page)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type,p.type
		FROM (
			SELECT parentId,childId,type
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp JOIN (
			SELECT pageId,type
			FROM pages
			WHERE isCurrentEdit AND type!=? AND type!=?`, CommentPageType, QuestionPageType).Add(`
		) AS p
		ON (p.pageId=pp.childId)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pp PagePair
		var childType string
		err := rows.Scan(&pp.ParentId, &pp.ChildId, &pp.Type, &childType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[pp.ChildId]
		if !ok {
			newPage = &Page{PageId: pp.ChildId, Type: childType}
			pageMap[newPage.PageId] = newPage
			newPages[newPage.PageId] = newPage
		}
		newPage.Parents = append(newPage.Parents, &pp)

		parent := sourcePageMap[pp.ParentId]
		if newPage.Type == LensPageType {
			parent.LensIds = append(parent.LensIds, fmt.Sprintf("%d", newPage.PageId))
		} else {
			parent.Children = append(parent.Children, &pp)
			parent.HasChildren = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if options.LoadHasChildren && len(newPages) > 0 {
		pageIds = PageIdsListFromMap(newPages)
		rows := database.NewQuery(`
			SELECT pp.parentId,sum(1)
			FROM (
				SELECT parentId,childId,type
				FROM pagePairs
				WHERE type=?`, ParentPagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIds).Add(`
			) AS pp JOIN (
				SELECT pageId
				FROM pages
				WHERE isCurrentEdit AND type!=? AND type!=?`, CommentPageType, QuestionPageType).Add(`
			) AS p
			ON (p.pageId=pp.childId)
			GROUP BY 1`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
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

// LoadSubpageIds loads the page ids for all the subpages of the pages in the given pageMap.
func LoadSubpageIds(db *database.DB, pageMap map[int64]*Page, sourcePageMap map[int64]*Page) error {
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type,p.type
		FROM (
			SELECT parentId,childId,type
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND parentId IN `).AddArgsGroup(pageIds).Add(`
		) AS pp JOIN (
			SELECT pageId,type
			FROM pages
			WHERE isCurrentEdit AND (type=? OR type=?)`, CommentPageType, QuestionPageType).Add(`
		) AS p
		ON (p.pageId=pp.childId)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pp PagePair
		var pageType string
		err := rows.Scan(&pp.ParentId, &pp.ChildId, &pp.Type, &pageType)
		if err != nil {
			return fmt.Errorf("failed to scan for subpages: %v", err)
		}
		newPage, ok := pageMap[pp.ChildId]
		if !ok {
			newPage = &Page{PageId: pp.ChildId, Type: pageType}
			pageMap[newPage.PageId] = newPage
		}
		newPage.Parents = append(newPage.Parents, &pp)

		sourcePageMap[pp.ParentId].SubpageIds = append(sourcePageMap[pp.ParentId].SubpageIds, fmt.Sprintf("%d", pp.ChildId))

		return nil
	})
	return err
}

// LoadTaggedAsIds for each page in the source map loads the ids of the pages that tag it.
func LoadTaggedAsIds(db *database.DB, pageMap map[int64]*Page, options LoadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE type=?`, TagPagePairType).Add(`AND childId IN`).AddArgsGroup(pageIds).Add(`
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pp PagePair
		err := rows.Scan(&pp.ParentId, &pp.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		child := sourcePageMap[pp.ChildId]
		child.TaggedAsIds = append(child.TaggedAsIds, fmt.Sprintf("%d", pp.ParentId))
		if _, ok := pageMap[pp.ParentId]; !ok {
			pageMap[pp.ParentId] = &Page{PageId: pp.ParentId}
		}
		return nil
	})
	return err
}

// LoadRelatedIds for each page in the source map loads the ids of the pages that are tagged by it.
func LoadRelatedIds(db *database.DB, pageMap map[int64]*Page, options LoadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE type=?`, TagPagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIds).Add(`
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pp PagePair
		err := rows.Scan(&pp.ParentId, &pp.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		parent := sourcePageMap[pp.ParentId]
		parent.RelatedIds = append(parent.RelatedIds, fmt.Sprintf("%d", pp.ChildId))
		if _, ok := pageMap[pp.ChildId]; !ok {
			pageMap[pp.ChildId] = &Page{PageId: pp.ChildId}
		}
		return nil
	})
	return err
}

// LoadRequirements for each page in the source map loads the ids of the pages that it has as a requirement.
func LoadRequirements(db *database.DB, userId int64, pageMap map[int64]*Page, masteryMap map[int64]*Mastery, options LoadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	masteryIds := make([]interface{}, 0)
	masteryIds = append(masteryIds, pageIds...)

	// Load the requirements for all pages
	rows := database.NewQuery(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE type=?`, RequirementPagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIds).Add(`
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId int64
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("failed to scan for requirement: %v", err)
		}
		parent := sourcePageMap[parentId]
		masteryIds = append(masteryIds, childId)
		parent.RequirementIds = append(parent.RequirementIds, fmt.Sprintf("%d", childId))
		if _, ok := pageMap[childId]; !ok {
			pageMap[childId] = &Page{PageId: childId}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Load what requirements the user has met
	rows = database.NewQuery(`
		SELECT masteryId,updatedAt,has,isManuallySet
		FROM userMasteryPairs
		WHERE userId=?`, userId).Add(`AND masteryId IN`).AddArgsGroup(masteryIds).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var mastery Mastery
		err := rows.Scan(&mastery.PageId, &mastery.UpdatedAt, &mastery.Has, &mastery.IsManuallySet)
		if err != nil {
			return fmt.Errorf("failed to scan for mastery: %v", err)
		}
		masteryMap[mastery.PageId] = &mastery
		return nil
	})
	if err != nil {
		return err
	}

	// Go through all the pages for which we loaded masteries, and if we haven't
	// loaded a mastery for it, just create one.
	for _, id := range masteryIds {
		masteryId := id.(int64)
		if _, ok := masteryMap[masteryId]; !ok {
			masteryMap[masteryId] = &Mastery{PageId: masteryId}
		}
	}
	return nil
}

type LoadParentsIdsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*Page
	// Load whether or not each parent has parents of its own.
	LoadHasParents bool
}

// LoadParentsIds loads the page ids for all the parents of the pages in the given pageMap.
func LoadParentsIds(db *database.DB, pageMap map[int64]*Page, options LoadParentsIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	newPages := make(map[int64]*Page)
	rows := database.NewQuery(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE type=?`, ParentPagePairType).Add(`
			AND childId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p PagePair
		err := rows.Scan(&p.ParentId, &p.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[p.ParentId]
		if !ok {
			newPage = &Page{PageId: p.ParentId}
			pageMap[newPage.PageId] = newPage
			newPages[newPage.PageId] = newPage
		}
		newPage.Children = append(newPage.Children, &p)
		sourcePageMap[p.ChildId].Parents = append(sourcePageMap[p.ChildId].Parents, &p)
		sourcePageMap[p.ChildId].HasParents = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}
	if options.LoadHasParents && len(newPages) > 0 {
		pageIds = PageIdsListFromMap(newPages)
		rows := database.NewQuery(`
			SELECT childId,sum(1)
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND childId IN`).AddArgsGroup(pageIds).Add(`
			GROUP BY 1`).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var parents int
			err := rows.Scan(&pageId, &parents)
			if err != nil {
				return fmt.Errorf("failed to scan for grandparents: %v", err)
			}
			pageMap[pageId].HasParents = parents > 0
			return nil
		})
		if err != nil {
			return fmt.Errorf("Failed to load grandparents: %v", err)
		}
	}
	return nil
}

// LoadDraftExistence computes for each page whether or not the user has an
// autosave draft for it.
// This only makes sense to call for pages which were loaded for isCurrentEdit=true.
func LoadDraftExistence(db *database.DB, userId int64, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId,MAX(
				IF(isAutosave AND creatorId=?, edit, -1)
			) as myMaxEdit, MAX(IF(isCurrentEdit, edit, -1)) AS currentEdit
		FROM pages`, userId).Add(`
		WHERE pageId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY pageId
		HAVING myMaxEdit > currentEdit`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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

// LoadLastVisits loads lastVisit variable for each page.
func LoadLastVisits(db *database.DB, currentUserId int64, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId,max(createdAt)
		FROM visits
		WHERE userId=?`, currentUserId).Add(`AND pageId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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

// LoadSubscriptions loads subscription statuses corresponding to the given
// pages, and then updates the given maps.
func LoadSubscriptions(db *database.DB, currentUserId int64, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT toPageId
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toPageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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

// LoadUserSubscriptions loads subscription statuses corresponding to the given
// users, and then updates the given map.
func LoadUserSubscriptions(db *database.DB, currentUserId int64, userMap map[int64]*User) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIds := IdsListFromUserMap(userMap)
	rows := database.NewQuery(`
		SELECT toUserId
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toUserId IN`).AddArgsGroup(userIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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

type LoadAuxPageDataOptions struct {
	// If set, pretend that we last visited all the pages on this date.
	// Used when we refresh the page, but don't want to erase the new/updated stars just yet.
	ForcedLastVisit string
}

// LoadAuxPageData loads the auxillary page data for the given pages.
func LoadAuxPageData(db *database.DB, userId int64, pageMap map[int64]*Page, options *LoadAuxPageDataOptions) error {
	if options == nil {
		options = &LoadAuxPageDataOptions{}
	}

	// Load likes
	err := LoadLikes(db, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load likes: %v", err)
	}

	// Load all the subscription statuses.
	if userId > 0 {
		err = LoadSubscriptions(db, userId, pageMap)
		if err != nil {
			return fmt.Errorf("Couldn't load subscriptions: %v", err)
		}
	}

	// Load whether or not pages have drafts.
	err = LoadDraftExistence(db, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load draft existence: %v", err)
	}

	// Load original creation date.
	if len(pageMap) > 0 {
		pageIds := PageIdsListFromMap(pageMap)
		rows := db.NewStatement(`
			SELECT pageId,MIN(createdAt)
			FROM pages
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
				AND NOT isAutosave AND NOT isSnapshot
			GROUP BY 1`).Query(pageIds...)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
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
	err = LoadLastVisits(db, userId, pageMap)
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
