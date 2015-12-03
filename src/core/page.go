// page.go contains all the page stuff
package core

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

const (
	// Various page types we have in our system.
	WikiPageType     = "wiki"
	CommentPageType  = "comment"
	QuestionPageType = "question"
	AnswerPageType   = "answer"
	LensPageType     = "lens"
	GroupPageType    = "group"
	DomainPageType   = "domain"

	// Various types of page connections.
	ParentPagePairType      = "parent"
	TagPagePairType         = "tag"
	RequirementPagePairType = "requirement"

	// Options for sorting page's children.
	RecentFirstChildSortingOption  = "recentFirst"
	OldestFirstChildSortingOption  = "oldestFirst"
	AlphabeticalChildSortingOption = "alphabetical"
	LikesChildSortingOption        = "likes"

	// Options for vote types
	ProbabilityVoteType = "probability"
	ApprovalVoteType    = "approval"

	// Various events we log when a page changes
	NewParentChangeLog         = "newParent"
	DeleteParentChangeLog      = "deleteParent"
	NewChildChangeLog          = "newChild"
	DeleteChildChangeLog       = "deleteChild"
	NewTagChangeLog            = "newTag"
	DeleteTagChangeLog         = "deleteTag"
	NewTagTargetChangeLog      = "newTagTarget"
	DeleteTagTargetChangeLog   = "deleteTagTarget"
	NewRequirementChangeLog    = "newRequirement"
	DeleteRequirementChangeLog = "deleteRequirement"
	NewRequiredForChangeLog    = "newRequiredFor"
	DeleteRequiredForChangeLog = "deleteRequiredFor"
	DeletePageChangeLog        = "deletePage"
	NewEditChangeLog           = "newEdit"
	RevertEditChangeLog        = "revertEdit"
	NewSnapshotChangeLog       = "newSnapshot"
	NewAliasChangeLog          = "newAlias"
	NewSortChildrenByChangeLog = "newSortChildrenBy"
	TurnOnVoteChangeLog        = "turnOnVote"
	TurnOffVoteChangeLog       = "turnOffVote"
	SetVoteTypeChangeLog       = "setVoteType"
	NewEditKarmaLockChangeLog  = "newEditKarmaLock"
	NewEditGroupChangeLog      = "newEditGroup"

	// How long the page lock lasts
	PageQuickLockDuration = 5 * 60  // in seconds
	PageLockDuration      = 30 * 60 // in seconds

	// String that can be used inside a regexp to match an a page alias or id
	AliasRegexpStr          = "[A-Za-z0-9]+\\.?[A-Za-z0-9]*"
	SubdomainAliasRegexpStr = "[A-Za-z0-9]*"
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

// corePageData has data we load directly from pages table.
type corePageData struct {
	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageId            int64  `json:"pageId,string"`
	Edit              int    `json:"edit"`
	Type              string `json:"type"`
	Title             string `json:"title"`
	Clickbait         string `json:"clickbait"`
	TextLength        int    `json:"textLength"`
	Alias             string `json:"alias"`
	SortChildrenBy    string `json:"sortChildrenBy"`
	HasVote           bool   `json:"hasVote"`
	VoteType          string `json:"voteType"`
	CreatorId         int64  `json:"creatorId,string"`
	CreatedAt         string `json:"createdAt"`
	OriginalCreatedAt string `json:"originalCreatedAt"`
	EditKarmaLock     int    `json:"editKarmaLock"`
	SeeGroupId        int64  `json:"seeGroupId,string"`
	EditGroupId       int64  `json:"editGroupId,string"`
	IsAutosave        bool   `json:"isAutosave"`
	IsSnapshot        bool   `json:"isSnapshot"`
	IsCurrentEdit     bool   `json:"isCurrentEdit"`
	IsMinorEdit       bool   `json:"isMinorEdit"`
	TodoCount         int    `json:"todoCount"`
	AnchorContext     string `json:"anchorContext"`
	AnchorText        string `json:"anchorText"`
	AnchorOffset      int    `json:"anchorOffset"`

	// The following data is filled on demand.
	Text     string `json:"text"`
	MetaText string `json:"metaText"`
	Summary  string `json:"summary"`
}

type Page struct {
	corePageData

	LoadOptions PageLoadOptions `json:"-"`

	// === Auxillary data. ===
	// For some pages we load additional data.
	IsSubscribed bool `json:"isSubscribed"`
	LikeCount    int  `json:"likeCount"`
	DislikeCount int  `json:"dislikeCount"`
	MyLikeValue  int  `json:"myLikeValue"`
	// Computed from LikeCount and DislikeCount
	LikeScore int `json:"likeScore"`
	// Last time the user visited this page.
	LastVisit string `json:"lastVisit"`
	// True iff the user has a work-in-progress draft for this page
	HasDraft bool `json:"hasDraft"`

	// === Full data. ===
	// For pages that are displayed fully, we load more additional data.
	// Edit number for the currently live version
	CurrentEditNum int `json:"currentEditNum"`
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
	NextPageId  int64  `json:"nextPageId,string"`
	PrevPageId  int64  `json:"prevPageId,string"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".

	// Subpages.
	AnswerIds      []string `json:"answerIds"`
	CommentIds     []string `json:"commentIds"`
	QuestionIds    []string `json:"questionIds"`
	LensIds        []string `json:"lensIds"`
	TaggedAsIds    []string `json:"taggedAsIds"`
	RelatedIds     []string `json:"relatedIds"`
	RequirementIds []string `json:"requirementIds"`

	// Subpage counts (these might be zeroes if not loaded explicitly)
	AnswerCount  int `json:"answerCount"`
	CommentCount int `json:"commentCount"`

	// Domains.
	DomainIds []string `json:"domainIds"`

	// List of changes to this page
	ChangeLogs []*ChangeLog `json:"changeLogs"`

	// Whether or not this page has children
	HasChildren bool `json:"hasChildren"`
	// Whether or not this page has parents
	HasParents bool     `json:"hasParents"`
	ChildIds   []string `json:"childIds"`
	ParentIds  []string `json:"parentIds"`

	// Populated for groups
	Members map[string]*Member `json:"members"`
}

// ChangeLog describes a row from changeLogs table.
type ChangeLog struct {
	UserId    int64  `json:"userId,string"`
	Edit      int    `json:"edit"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	AuxPageId int64  `json:"auxPageId,string"`
}

// Mastery is a page you should have mastered before you can understand another page.
type Mastery struct {
	PageId        int64  `json:"pageId,string"`
	Has           bool   `json:"has"`
	UpdatedAt     string `json:"updatedAt"`
	IsManuallySet bool   `json:"isManuallySet"`
}

// LoadDataOption is used to set some simple loading options for loading functions
type LoadDataOptions struct {
	// If set, we'll only load links for the pages with these ids
	ForPages map[int64]*Page
}

// ExecuteLoadPipeline runs the pages in the pageMap through the pipeline to load
// all the data they need.
func ExecuteLoadPipeline(db *database.DB, u *user.User, pageMap map[int64]*Page, userMap map[int64]*User, masteryMap map[int64]*Mastery) error {

	// Load answers
	filteredPageMap := filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Answers })
	err := LoadChildIds(db, pageMap, &LoadChildIdsOptions{
		ForPages:     filteredPageMap,
		Type:         AnswerPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  SubpageLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadChildIds for answers failed: %v", err)
	}

	// Load comments
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Comments })
	err = LoadCommentIds(db, pageMap, &LoadDataOptions{ForPages: filteredPageMap})
	if err != nil {
		return fmt.Errorf("LoadCommentIds for failed: %v", err)
	}

	// Load questions
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Questions })
	err = LoadChildIds(db, pageMap, &LoadChildIdsOptions{
		ForPages:     filteredPageMap,
		Type:         QuestionPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  IntrasitePopoverLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadChildIds for questions failed: %v", err)
	}

	// Load children
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Children })
	err = LoadChildIds(db, pageMap, &LoadChildIdsOptions{
		ForPages:     filteredPageMap,
		Type:         WikiPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadChildIds for children failed: %v", err)
	}

	// Load parents
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Parents })
	err = LoadParentIds(db, pageMap, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: ParentPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for parents failed: %v", err)
	}

	// Load tags
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Tags })
	err = LoadParentIds(db, pageMap, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: TagPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for tags failed: %v", err)
	}

	// Load related
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Related })
	err = LoadChildIds(db, pageMap, &LoadChildIdsOptions{
		ForPages:     filteredPageMap,
		Type:         WikiPageType,
		PagePairType: TagPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadChildIds for related failed: %v", err)
	}

	// Load available lenses
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Lenses })
	err = LoadChildIds(db, pageMap, &LoadChildIdsOptions{
		ForPages:     filteredPageMap,
		Type:         LensPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  LensInfoLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadChildIds for lenses failed: %v", err)
	}

	// Load requirements
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Requirements })
	err = LoadParentIds(db, pageMap, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: RequirementPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
		MasteryMap:   masteryMap,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for requirements failed: %v", err)
	}

	// Load links
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Links })
	err = LoadLinks(db, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLinks failed: %v", err)
	}

	// TODO: Load domains

	// Load change logs
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ChangeLogs })
	err = LoadChangeLogs(db, u.Id, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadChangeLogs failed: %v", err)
	}

	// Load whether or not the pages have child drafts
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ChildDraftId })
	err = LoadChildDrafts(db, u.Id, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadChildDrafts failed: %v", err)
	}

	// Load whether or not the pages have an unpublished draft
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.HasDraft })
	err = LoadDraftExistence(db, u.Id, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadChildDrafts failed: %v", err)
	}

	// Load (dis)likes
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Likes })
	err = LoadLikes(db, u.Id, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadLikes failed: %v", err)
	}

	// Load votes
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Votes })
	err = LoadVotes(db, u.Id, filteredPageMap, userMap)
	if err != nil {
		return fmt.Errorf("LoadLikes failed: %v", err)
	}

	// Load last visit dates
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.LastVisit })
	err = LoadLastVisits(db, u.Id, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadLastVisits failed: %v", err)
	}

	// Load subscriptions
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.IsSubscribed })
	err = LoadSubscriptions(db, u.Id, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSubscriptions failed: %v", err)
	}

	// Load subpage counts
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SubpageCounts })
	err = LoadSubpageCounts(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSubpageCounts failed: %v", err)
	}

	// Load number of red links.
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.RedLinkCount })
	err = LoadRedLinkCount(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadRedLinkCount failed: %v", err)
	}

	// Add other pages we'll need
	AddUserGroupIdsToPageMap(u, pageMap)

	// Load page data
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return !p.LoadOptions.Edit })
	err = LoadPages(db, u, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadPages failed: %v", err)
	}

	// Load prev/next ids
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool {
		return p.Type == WikiPageType && p.LoadOptions.NextPrevIds
	})
	err = LoadNextPrevPageIds(db, u.Id, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadNextPrevPageIds failed: %v", err)
	}

	// Add pages that need a corresponding mastery to the masteryMap
	for _, p := range pageMap {
		if p.LoadOptions.Mastery {
			masteryMap[p.PageId] = &Mastery{PageId: p.PageId}
		}
	}

	// Load what requirements the user has met
	err = LoadMasteries(db, u.Id, masteryMap)
	if err != nil {
		return fmt.Errorf("LoadMasteries failed: %v", err)
	}

	// Load all the users
	userMap[u.Id] = &User{Id: u.Id}
	for _, p := range pageMap {
		if p.LoadOptions.Text || p.LoadOptions.Summary {
			userMap[p.CreatorId] = &User{Id: p.CreatorId}
		}
		if p.LockedBy != 0 {
			userMap[p.LockedBy] = &User{Id: p.LockedBy}
		}
	}
	err = LoadUsers(db, userMap, u.Id)
	if err != nil {
		return fmt.Errorf("LoadUsers failed: %v", err)
	}

	return nil
}

// LoadMasteries loads the mastery.
func LoadMasteries(db *database.DB, userId int64, masteryMap map[int64]*Mastery) error {
	if len(masteryMap) <= 0 {
		return nil
	}
	masteryIds := make([]interface{}, 0)
	for id, _ := range masteryMap {
		masteryIds = append(masteryIds, id)
	}
	rows := database.NewQuery(`
		SELECT masteryId,updatedAt,has,isManuallySet
		FROM userMasteryPairs
		WHERE userId=?`, userId).Add(`AND masteryId IN`).AddArgsGroup(masteryIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var mastery Mastery
		err := rows.Scan(&mastery.PageId, &mastery.UpdatedAt, &mastery.Has, &mastery.IsManuallySet)
		if err != nil {
			return fmt.Errorf("failed to scan for mastery: %v", err)
		}
		masteryMap[mastery.PageId] = &mastery
		return nil
	})
	return err
}

// LoadPages loads the given pages.
func LoadPages(db *database.DB, user *user.User, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)

	// Compute pages for which to load text / summary
	textIds := make([]interface{}, 0)
	summaryIds := make([]interface{}, 0)
	for _, p := range pageMap {
		if p.LoadOptions.Text {
			textIds = append(textIds, p.PageId)
		}
		if p.LoadOptions.Summary {
			summaryIds = append(summaryIds, p.PageId)
		}
	}
	textSelect := database.NewQuery(`IF(p.pageId IN`).AddIdsGroup(textIds).Add(`,p.text,"") AS text`)
	summarySelect := database.NewQuery(`IF(p.pageId IN`).AddIdsGroup(summaryIds).Add(`,p.summary,"") AS summary`)

	// Load the page data
	rows := database.NewQuery(`
		SELECT p.pageId,p.edit,p.creatorId,p.createdAt,p.title,p.clickbait,`).AddPart(textSelect).Add(`,
			length(p.text),p.metaText,pi.type,pi.editKarmaLock,pi.hasVote,pi.voteType,`).AddPart(summarySelect).Add(`,
			pi.alias,pi.createdAt,pi.sortChildrenBy,pi.seeGroupId,pi.editGroupId,
			p.isAutosave,p.isSnapshot,p.isCurrentEdit,p.isMinorEdit,
			p.todoCount,p.anchorContext,p.anchorText,p.anchorOffset
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId = pi.pageId AND p.isCurrentEdit)
		WHERE p.pageId IN`).AddArgsGroup(pageIds).Add(`
			AND (pi.seeGroupId=0 OR pi.seeGroupId IN`).AddIdsGroupStr(user.GroupIds).Add(`)
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p corePageData
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.CreatorId, &p.CreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.Type, &p.EditKarmaLock, &p.HasVote,
			&p.VoteType, &p.Summary, &p.Alias, &p.OriginalCreatedAt, &p.SortChildrenBy,
			&p.SeeGroupId, &p.EditGroupId,
			&p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
			&p.TodoCount, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}
		pageMap[p.PageId].corePageData = p
		return nil
	})
	for _, p := range pageMap {
		if p.Type == "" {
			delete(pageMap, p.PageId)
		}
	}
	return err
}

// LoadChangeLogs loads the edit history for the given page.
func LoadChangeLogs(db *database.DB, userId int64, pageMap map[int64]*Page, userMap map[int64]*User, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	for _, p := range sourcePageMap {
		p.ChangeLogs = make([]*ChangeLog, 0)
		rows := database.NewQuery(`
		SELECT userId,edit,type,createdAt,auxPageId
		FROM changeLogs
		WHERE pageId=?`, p.PageId).Add(`
			AND (userId=? OR type!=?)`, userId, NewSnapshotChangeLog).Add(`
		ORDER BY createdAt DESC`).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var l ChangeLog
			err := rows.Scan(&l.UserId, &l.Edit, &l.Type, &l.CreatedAt, &l.AuxPageId)
			if err != nil {
				return fmt.Errorf("failed to scan a page log: %v", err)
			}
			p.ChangeLogs = append(p.ChangeLogs, &l)
			return nil
		})
		if err != nil {
			return fmt.Errorf("Couldn't load changeLogs: %v", err)
		}

		// Process change logs
		for _, log := range p.ChangeLogs {
			userMap[log.UserId] = &User{Id: log.UserId}
			AddPageIdToMap(log.AuxPageId, pageMap)
		}
	}
	return nil
}

type LoadEditOptions struct {
	// If true, the last edit will be loaded for the given user, even if it's an
	// autosave or a snapshot.
	LoadNonliveEdit bool

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
	p := NewPage(pageId)

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
		SELECT p.pageId,p.edit,pi.type,p.title,p.clickbait,p.text,p.metaText,
			p.summary,pi.alias,p.creatorId,pi.sortChildrenBy,pi.hasVote,pi.voteType,
			p.createdAt,pi.editKarmaLock,pi.seeGroupId,pi.editGroupId,pi.createdAt,
			p.isAutosave,p.isSnapshot,p.isCurrentEdit,p.isMinorEdit,
			p.todoCount,p.anchorContext,p.anchorText,p.anchorOffset,
			pi.currentEdit>0,pi.currentEdit,pi.maxEdit,pi.lockedBy,pi.lockedUntil,
			pi.voteType
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId AND p.pageId=?)`, pageId).Add(`
		WHERE`).AddPart(whereClause).Add(`AND
			(pi.seeGroupId=0 OR pi.seeGroupId IN (SELECT groupId FROM groupMembers WHERE userId=?))`, userId).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageId, &p.Edit, &p.Type, &p.Title, &p.Clickbait,
		&p.Text, &p.MetaText, &p.Summary, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.EditKarmaLock, &p.SeeGroupId,
		&p.EditGroupId, &p.OriginalCreatedAt, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
		&p.TodoCount, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset, &p.WasPublished,
		&p.CurrentEditNum, &p.MaxEditEver, &p.LockedBy, &p.LockedUntil, &p.LockedVoteType)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}

	p.TextLength = len(p.Text)
	return p, nil
}

// LoadPageIds from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func LoadPageIds(rows *database.Rows, pageMap map[int64]*Page, loadOptions *PageLoadOptions) ([]string, error) {
	ids := make([]string, 0)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId int64
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}

		AddPageToMap(pageId, pageMap, loadOptions)
		ids = append(ids, fmt.Sprintf("%d", pageId))
		return nil
	})
	return ids, err
}

// LoadChildDrafts loads a potentially existing draft for the given page. If it's
// loaded, it'll be added to the given map.
func LoadChildDrafts(db *database.DB, userId int64, options *LoadDataOptions) error {
	if len(options.ForPages) > 1 {
		db.C.Warningf("LoadChildDrafts called with more than one page")
	}
	for _, p := range options.ForPages {
		row := database.NewQuery(`
				SELECT a.pageId
				FROM (
					SELECT p.pageId,p.creatorId
					FROM pages AS p
					JOIN pagePairs AS pp
					ON (p.pageId=pp.childId)
					JOIN pageInfos AS pi
					ON (p.pageId=pi.pageId)
					WHERE pp.parentId=? AND`, p.PageId).Add(`
						(pi.type=? OR pi.type=?)`, QuestionPageType, AnswerPageType).Add(`
					GROUP BY p.pageId
					HAVING SUM(p.isCurrentEdit)<=0
				) AS a
				WHERE a.creatorId=?`, userId).Add(`
				LIMIT 1`).ToStatement(db).QueryRow()
		_, err := row.Scan(&p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load answer draft id: %v", err)
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
		// We count the current user's like value towards the sum here in the FE.
		if userId == currentUserId {
			page.MyLikeValue = value
		} else if value > 0 {
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
		return nil
	})
	return err
}

// LoadVotes loads probability votes corresponding to the given pages and updates the pages.
func LoadVotes(db *database.DB, currentUserId int64, pageMap map[int64]*Page, usersMap map[int64]*User) error {
	if len(pageMap) <= 0 {
		return nil
	}

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
		SELECT l.parentId,SUM(ISNULL(pi.pageId))
		FROM pageInfos AS pi
		RIGHT JOIN links AS l
		ON ((pi.pageId=l.childAlias OR pi.alias=l.childAlias) AND pi.currentEdit>0 AND pi.type!="")
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

// LoadLinks loads the links for the given pages, and adds them to the pageMap.
func LoadLinks(db *database.DB, pageMap map[int64]*Page, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}

	sourceMap := options.ForPages
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
			AddPageIdToMap(pageId, pageMap)
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
			FROM pageInfos
			WHERE currentEdit>0 AND alias IN`).AddArgsGroup(aliasesList).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan for a page: %v", err)
			}
			AddPageIdToMap(pageId, pageMap)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type LoadChildIdsOptions struct {
	// If set, the children will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*Page
	// Type of children to load
	Type string
	// Type of the child relationship to follow
	PagePairType string
	// Load options to set for the new pages
	LoadOptions *PageLoadOptions
}

// LoadChildIds loads the page ids for all the children of the pages in the given pageMap.
func LoadChildIds(db *database.DB, pageMap map[int64]*Page, options *LoadChildIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE type=?`, options.PagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN pageInfos AS pi
		ON (pi.pageId=pp.childId AND pi.currentEdit>0 AND pi.type=?)`, options.Type).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId int64
		var ppType string
		err := rows.Scan(&parentId, &childId, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(childId, pageMap, options.LoadOptions)

		parent := sourcePageMap[parentId]
		if options.Type == LensPageType {
			parent.LensIds = append(parent.LensIds, fmt.Sprintf("%d", newPage.PageId))
		} else if options.Type == AnswerPageType {
			parent.AnswerIds = append(parent.AnswerIds, fmt.Sprintf("%d", newPage.PageId))
		} else if options.Type == CommentPageType {
			parent.CommentIds = append(parent.CommentIds, fmt.Sprintf("%d", newPage.PageId))
		} else if options.Type == QuestionPageType {
			parent.QuestionIds = append(parent.QuestionIds, fmt.Sprintf("%d", newPage.PageId))
		} else if options.Type == WikiPageType && options.PagePairType == ParentPagePairType {
			parent.ChildIds = append(parent.ChildIds, fmt.Sprintf("%d", childId))
			parent.HasChildren = true
			if parent.LoadOptions.HasGrandChildren {
				newPage.LoadOptions.SubpageCounts = true
			}
			if parent.LoadOptions.RedLinkCountForChildren {
				newPage.LoadOptions.RedLinkCount = true
			}
		}
		return nil
	})
	return err
}

// LoadSubpageCounts loads the number of various types of children the pages have
func LoadSubpageCounts(db *database.DB, pageMap map[int64]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pi.type,sum(1)
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp JOIN (
			SELECT pageId,type
			FROM pageInfos
			WHERE currentEdit>0`).Add(`
		) AS pi
		ON (pi.pageId=pp.childId)
		GROUP BY 1,2`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId int64
		var childType string
		var count int
		err := rows.Scan(&pageId, &childType, &count)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		if childType == CommentPageType {
			pageMap[pageId].CommentCount = count
		} else if childType == AnswerPageType {
			pageMap[pageId].AnswerCount = count
		}
		return nil
	})
	return err
}

// LoadTaggedAsIds for each page in the source map loads the ids of the pages that tag it.
func LoadTaggedAsIds(db *database.DB, pageMap map[int64]*Page, options *LoadChildIdsOptions) error {
	if options == nil {
		options = &LoadChildIdsOptions{}
	}
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
		var parentId, childId int64
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		child := sourcePageMap[childId]
		child.TaggedAsIds = append(child.TaggedAsIds, fmt.Sprintf("%d", parentId))
		AddPageIdToMap(parentId, pageMap)
		return nil
	})
	return err
}

type LoadParentIdsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*Page
	// Load whether or not each parent has parents of its own.
	LoadHasParents bool
	// Type of the parent relationship to follow
	PagePairType string
	// Load options to set for the new pages
	LoadOptions *PageLoadOptions
	// Mastery map to populate with masteries necessary for a requirement
	MasteryMap map[int64]*Mastery
}

// LoadParentIds loads the page ids for all the parents of the pages in the given pageMap.
func LoadParentIds(db *database.DB, pageMap map[int64]*Page, options *LoadParentIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	newPages := make(map[int64]*Page)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE type=?`, options.PagePairType).Add(`AND childId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN pageInfos AS pi
		ON (pi.pageId=pp.parentId AND pi.currentEdit>0)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId int64
		var ppType string
		err := rows.Scan(&parentId, &childId, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(parentId, pageMap, options.LoadOptions)
		childPage := pageMap[childId]

		if options.PagePairType == ParentPagePairType {
			childPage.ParentIds = append(childPage.ParentIds, fmt.Sprintf("%d", parentId))
			childPage.HasParents = true
			newPages[newPage.PageId] = newPage
		} else if options.PagePairType == RequirementPagePairType {
			childPage.RequirementIds = append(childPage.RequirementIds, fmt.Sprintf("%d", parentId))
			// If it's a requirement, add to the mastery map
			if _, ok := options.MasteryMap[parentId]; !ok {
				options.MasteryMap[parentId] = &Mastery{PageId: parentId}
			}
		} else if options.PagePairType == TagPagePairType {
			childPage.TaggedAsIds = append(childPage.TaggedAsIds, fmt.Sprintf("%d", parentId))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}

	// Load if parents have parents
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

// LoadCommentIds loads ids of all the comments for the pages in the given pageMap.
func LoadCommentIds(db *database.DB, pageMap map[int64]*Page, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE type=?`, ParentPagePairType).Add(`AND childId IN (
			SELECT pp.childId
			FROM pagePairs AS pp
			JOIN pageInfos AS pi
			ON (pi.pageId=pp.childId AND pi.currentEdit>0 AND pi.type=?`, CommentPageType).Add(`
				AND pp.type=?`, ParentPagePairType).Add(`
				AND pp.parentId IN`).AddArgsGroup(pageIds).Add(`)
		)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId int64
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		parentPage := AddPageToMap(parentId, pageMap, SubpageLoadOptions)
		childPage := AddPageToMap(childId, pageMap, SubpageLoadOptions)
		parentPage.CommentIds = append(parentPage.CommentIds, fmt.Sprintf("%d", childPage.PageId))
		return nil
	})

	// Now we have pages that have in comment ids both top level comments and
	// replies, so we need to remove the replies.
	for _, p := range sourcePageMap {
		replies := make(map[string]bool)
		for _, c := range p.CommentIds {
			commentId, _ := strconv.ParseInt(c, 10, 64)
			for _, r := range pageMap[commentId].CommentIds {
				replies[r] = true
			}
		}
		onlyTopCommentIds := make([]string, 0)
		for _, c := range p.CommentIds {
			if !replies[c] {
				onlyTopCommentIds = append(onlyTopCommentIds, c)
			}
		}
		p.CommentIds = onlyTopCommentIds
	}
	return err
}

// loadOrderedChildrenIds loads and returns ordered list of children for the
// given parent page
func loadOrderedChildrenIds(db *database.DB, parentId int64, sortType string) ([]int64, error) {
	orderClause := ""
	if sortType == RecentFirstChildSortingOption {
		orderClause = "pi.createdAt DESC"
	} else if sortType == OldestFirstChildSortingOption {
		orderClause = "pi.createdAt"
	} else if sortType == AlphabeticalChildSortingOption {
		orderClause = "p.title"
	} else {
		return nil, nil
	}
	childrenIds := make([]int64, 0)
	rows := database.NewQuery(`
		SELECT pp.childId
		FROM pagePairs AS pp
		JOIN pages AS p
		ON (pp.childId=p.pageId) 
		JOIN pageInfos AS pi
		ON (pi.pageId=p.pageId)
		WHERE p.isCurrentEdit
			AND pi.type!=? AND pi.type!=? AND pi.type!=?`, CommentPageType, QuestionPageType, LensPageType).Add(`
			AND pp.type=?`, ParentPagePairType).Add(`AND pp.parentId=?`, parentId).Add(`
		ORDER BY ` + orderClause).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var childId int64
		err := rows.Scan(&childId)
		if err != nil {
			return fmt.Errorf("failed to scan for childId: %v", err)
		}
		childrenIds = append(childrenIds, childId)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to load children: %v", err)
	}
	return childrenIds, nil
}

// loadSiblingId loads the next/prev sibling page id, based on the "parent"
// relationships. We search recursively up the hierarchy if necessary.
func loadSiblingId(db *database.DB, pageId int64, useNextSibling bool) (int64, error) {
	// Load the parent and the sorting order
	var parentId int64
	var sortType string
	var parentCount int
	row := database.NewQuery(`
		SELECt ifnull(max(pp.parentId),0),ifnull(max(pi.sortChildrenBy),""),count(*)
		FROM pagePairs AS pp
		JOIN pageInfos AS pi
		ON (pp.parentId=pi.pageId AND pi.currentEdit>0)
		WHERE pp.type=?`, ParentPagePairType).Add(`AND pp.childId=?`, pageId).ToStatement(db).QueryRow()
	found, err := row.Scan(&parentId, &sortType, &parentCount)
	if err != nil {
		return 0, err
	} else if !found || parentCount != 1 {
		return 0, nil
	}

	// Load the sibling pages in order
	orderedSiblingIds, err := loadOrderedChildrenIds(db, parentId, sortType)
	if err != nil {
		return 0, fmt.Errorf("Failed to load children: %v", err)
	}

	// Find where the current page sits in the ordered sibling list
	pageSiblingIndex := -1
	for i, childId := range orderedSiblingIds {
		if childId == pageId {
			pageSiblingIndex = i
			break
		}
	}
	// Then get the next / prev sibling accordingly
	if useNextSibling {
		if pageSiblingIndex < len(orderedSiblingIds)-1 {
			return orderedSiblingIds[pageSiblingIndex+1], nil
		} else if pageSiblingIndex == len(orderedSiblingIds)-1 {
			// It's the last child, so we need to recurse
			return loadSiblingId(db, parentId, useNextSibling)
		}
	} else {
		if pageSiblingIndex > 0 {
			return orderedSiblingIds[pageSiblingIndex-1], nil
		} else if pageSiblingIndex == 0 {
			// It's the first child, so just return the parent
			return parentId, nil
		}
	}
	return 0, nil
}

// LoadNextPrevPageIds loads the pages that come before / after the given page
// in the reading sequence.
func LoadNextPrevPageIds(db *database.DB, userId int64, options *LoadDataOptions) error {
	if len(options.ForPages) > 1 {
		db.C.Warningf("LoadNextPrevPageIds called with more than one page")
	}
	for _, p := range options.ForPages {
		var err error
		p.PrevPageId, err = loadSiblingId(db, p.PageId, false)
		if err != nil {
			return fmt.Errorf("Error while loading prev page id: %v", err)
		}

		// NextPageId will be the first child if there are children
		orderedChildrenIds, err := loadOrderedChildrenIds(db, p.PageId, p.SortChildrenBy)
		if err != nil {
			return fmt.Errorf("Error getting first child: %v", err)
		}
		if len(orderedChildrenIds) > 0 {
			p.NextPageId = orderedChildrenIds[0]
		}

		// If there are no children, then get the next sibling
		if p.NextPageId <= 0 {
			p.NextPageId, err = loadSiblingId(db, p.PageId, true)
			if err != nil {
				return fmt.Errorf("Error while loading next page id: %v", err)
			}
		}
	}
	return nil
}

// LoadDraftExistence computes for each page whether or not the user has an
// autosave draft for it.
// This only makes sense to call for pages which were loaded for isCurrentEdit=true.
func LoadDraftExistence(db *database.DB, userId int64, options *LoadDataOptions) error {
	pageMap := options.ForPages
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId,MAX(
				IF(isAutosave AND creatorId=?, edit, -1)
			) AS myMaxEdit, MAX(IF(isCurrentEdit, edit, -1)) AS currentEdit
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
		SELECT toId
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
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

// LoadAliasToPageIdMap loads the mapping from aliases to page ids.
func LoadAliasToPageIdMap(db *database.DB, aliases []string) (map[string]int64, error) {
	aliasToIdMap := make(map[string]int64)

	strictAliases := make([]string, 0)
	for _, alias := range aliases {
		pageId, err := strconv.ParseInt(alias, 10, 64)
		if err == nil {
			aliasToIdMap[alias] = pageId
		} else {
			strictAliases = append(strictAliases, alias)
		}
	}

	if len(strictAliases) > 0 {
		rows := database.NewQuery(`
			SELECT pageId,alias
			FROM pageInfos
			WHERE currentEdit>0 AND alias IN`).AddArgsGroupStr(strictAliases).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var alias string
			err := rows.Scan(&pageId, &alias)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			aliasToIdMap[strings.ToLower(alias)] = pageId
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("Couldn't convert pageId=>alias", err)
		}
	}
	return aliasToIdMap, nil
}

// LoadAliasToPageId converts the given page alias to page id.
func LoadAliasToPageId(db *database.DB, alias string) (int64, bool, error) {
	aliasToIdMap, err := LoadAliasToPageIdMap(db, []string{alias})
	if err != nil {
		return 0, false, err
	}
	pageId, ok := aliasToIdMap[strings.ToLower(alias)]
	return pageId, ok, nil
}
