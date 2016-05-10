// page.go contains all the page stuff
package core

import (
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/database"
)

const (
	// Various page types we have in our system.
	WikiPageType     = "wiki"
	CommentPageType  = "comment"
	QuestionPageType = "question"
	LensPageType     = "lens"
	GroupPageType    = "group"
	DomainPageType   = "domain"

	// Various types of page connections.
	ParentPagePairType      = "parent"
	TagPagePairType         = "tag"
	RequirementPagePairType = "requirement"
	SubjectPagePairType     = "subject"

	// Options for sorting page's children.
	RecentFirstChildSortingOption  = "recentFirst"
	OldestFirstChildSortingOption  = "oldestFirst"
	AlphabeticalChildSortingOption = "alphabetical"
	LikesChildSortingOption        = "likes"

	// Options for vote types
	ProbabilityVoteType = "probability"
	ApprovalVoteType    = "approval"

	// Various events we log when a page changes
	NewParentChangeLog          = "newParent"
	DeleteParentChangeLog       = "deleteParent"
	NewChildChangeLog           = "newChild"
	DeleteChildChangeLog        = "deleteChild"
	NewLensChangeLog            = "newLens"
	NewTagChangeLog             = "newTag"
	DeleteTagChangeLog          = "deleteTag"
	NewUsedAsTagChangeLog       = "newUsedAsTag"
	DeleteUsedAsTagChangeLog    = "deleteUsedAsTag"
	NewRequirementChangeLog     = "newRequirement"
	DeleteRequirementChangeLog  = "deleteRequirement"
	NewRequiredByChangeLog      = "newRequiredBy"
	DeleteRequiredByChangeLog   = "deleteRequiredBy"
	NewSubjectChangeLog         = "newSubject"
	DeleteSubjectChangeLog      = "deleteSubject"
	NewTeacherChangeLog         = "newTeacher"
	DeleteTeacherChangeLog      = "deleteTeacher"
	DeletePageChangeLog         = "deletePage"
	UndeletePageChangeLog       = "undeletePage"
	NewEditChangeLog            = "newEdit"
	RevertEditChangeLog         = "revertEdit"
	NewSnapshotChangeLog        = "newSnapshot"
	NewAliasChangeLog           = "newAlias"
	NewSortChildrenByChangeLog  = "newSortChildrenBy"
	TurnOnVoteChangeLog         = "turnOnVote"
	TurnOffVoteChangeLog        = "turnOffVote"
	SetVoteTypeChangeLog        = "setVoteType"
	NewEditGroupChangeLog       = "newEditGroup"
	SearchStringChangeChangeLog = "searchStringChange"
	AnswerChangeChangeLog       = "answerChange"

	// Mark types
	QueryMarkType     = "query"
	TypoMarkType      = "typo"
	ConfusionMarkType = "confusion"

	// How long the page lock lasts
	PageQuickLockDuration = 5 * 60  // in seconds
	PageLockDuration      = 30 * 60 // in seconds

	// String that can be used inside a regexp to match an a page alias or id
	AliasRegexpStr          = "[A-Za-z0-9_]+\\.?[A-Za-z0-9_]*"
	SubdomainAliasRegexpStr = "[A-Za-z0-9_]*"
	ReplaceRegexpStr        = "[^A-Za-z0-9_]" // used for replacing non-alias characters
)

var (
	// Regexp that strictly matches an alias, and not a page id
	StrictAliasRegexp = regexp.MustCompile("^[0-9A-Za-z_]*[A-Za-z_][0-9A-Za-z_]*$")
)

type Vote struct {
	Value     int    `json:"value"`
	UserId    string `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

// corePageData has data we load directly from the pages and pageInfos tables.
type corePageData struct {
	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageId                   string `json:"pageId"`
	Edit                     int    `json:"edit"`
	PrevEdit                 int    `json:"prevEdit"`
	Type                     string `json:"type"`
	Title                    string `json:"title"`
	Clickbait                string `json:"clickbait"`
	TextLength               int    `json:"textLength"` // number of characters
	Alias                    string `json:"alias"`
	SortChildrenBy           string `json:"sortChildrenBy"`
	HasVote                  bool   `json:"hasVote"`
	VoteType                 string `json:"voteType"`
	CreatorId                string `json:"creatorId"`
	CreatedAt                string `json:"createdAt"`
	OriginalCreatedAt        string `json:"originalCreatedAt"`
	OriginalCreatedBy        string `json:"originalCreatedBy"`
	SeeGroupId               string `json:"seeGroupId"`
	EditGroupId              string `json:"editGroupId"`
	IsAutosave               bool   `json:"isAutosave"`
	IsSnapshot               bool   `json:"isSnapshot"`
	IsLiveEdit               bool   `json:"isLiveEdit"`
	IsMinorEdit              bool   `json:"isMinorEdit"`
	IsRequisite              bool   `json:"isRequisite"`
	IndirectTeacher          bool   `json:"indirectTeacher"`
	TodoCount                int    `json:"todoCount"`
	LensIndex                int    `json:"lensIndex"`
	IsEditorComment          bool   `json:"isEditorComment"`
	IsEditorCommentIntention bool   `json:"isEditorCommentIntention"`
	SnapshotText             string `json:"snapshotText"`
	AnchorContext            string `json:"anchorContext"`
	AnchorText               string `json:"anchorText"`
	AnchorOffset             int    `json:"anchorOffset"`
	MergedInto               string `json:"mergedInto"`
	IsDeleted                bool   `json:"isDeleted"`

	// The following data is filled on demand.
	Text     string `json:"text"`
	MetaText string `json:"metaText"`
}

type Page struct {
	corePageData

	LoadOptions PageLoadOptions `json:"-"`

	// === Auxillary data. ===
	// For some pages we load additional data.
	IsSubscribed    bool `json:"isSubscribed"`
	SubscriberCount int  `json:"subscriberCount"`
	LikeCount       int  `json:"likeCount"`
	DislikeCount    int  `json:"dislikeCount"`
	MyLikeValue     int  `json:"myLikeValue"`
	// Computed from LikeCount and DislikeCount
	LikeScore int `json:"likeScore"`
	ViewCount int `json:"viewCount"`
	// Last time the user visited this page.
	LastVisit string `json:"lastVisit"`
	// True iff the user has a work-in-progress draft for this page
	HasDraft bool `json:"hasDraft"`

	// === Full data. ===
	// For pages that are displayed fully, we load more additional data.
	// Edit number for the currently live version
	CurrentEdit int `json:"currentEdit"`

	// True iff there has ever been an edit that had isLiveEdit set for this page
	WasPublished bool `json:"wasPublished"`

	Votes []*Vote `json:"votes"`
	// We don't allow users to change the vote type once a page has been published
	// with a voteType!="" even once. If it has, this is the vote type it shall
	// always have.
	LockedVoteType string `json:"lockedVoteType"`
	// Highest edit number used for this page for all users
	MaxEditEver  int `json:"maxEditEver"`
	RedLinkCount int `json:"redLinkCount"`
	// Page is locked by this user
	LockedBy string `json:"lockedBy"`
	// User has the page lock until this time
	LockedUntil string `json:"lockedUntil"`
	NextPageId  string `json:"nextPageId"`
	PrevPageId  string `json:"prevPageId"`
	// Whether or not the page is used as a requirement
	UsedAsMastery bool `json:"usedAsMastery"`

	// What actions the current user can perform with this page
	Permissions *Permissions `json:"permissions"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".
	Summaries map[string]string `json:"summaries"`
	// Ids of the users who edited this page. Ordered by how much they contributed.
	CreatorIds []string `json:"creatorIds"`

	// Relevant ids.
	CommentIds     []string `json:"commentIds"`
	QuestionIds    []string `json:"questionIds"`
	LensIds        []string `json:"lensIds"`
	TaggedAsIds    []string `json:"taggedAsIds"`
	RelatedIds     []string `json:"relatedIds"`
	RequirementIds []string `json:"requirementIds"`
	SubjectIds     []string `json:"subjectIds"`
	DomainIds      []string `json:"domainIds"`
	ChildIds       []string `json:"childIds"`
	ParentIds      []string `json:"parentIds"`
	MarkIds        []string `json:"markIds"`
	// For user pages, this is the domains user has access to
	DomainMembershipIds []string `json:"domainMembershipIds"`

	// Answers associated with this page, if it's a question page.
	Answers []*Answer `json:"answers"`

	// Various counts (these might be zeroes if not loaded explicitly)
	AnswerCount     int `json:"answerCount"`
	CommentCount    int `json:"commentCount"`
	LinkedMarkCount int `json:"linkedMarkCount"`

	// List of changes to this page
	ChangeLogs []*ChangeLog `json:"changeLogs"`

	// List of search strings associated with this page. Map: string id -> string text
	SearchStrings map[string]string `json:"searchStrings"`

	// Whether or not this page has children
	HasChildren bool `json:"hasChildren"`
	// Whether or not this page has parents
	HasParents bool `json:"hasParents"`

	// Populated for groups
	Members map[string]*Member `json:"members"`
}

// ChangeLog describes a row from changeLogs table.
type ChangeLog struct {
	Id               int    `json:"id"`
	UserId           string `json:"userId"`
	Edit             int    `json:"edit"`
	Type             string `json:"type"`
	CreatedAt        string `json:"createdAt"`
	AuxPageId        string `json:"auxPageId"`
	OldSettingsValue string `json:"oldSettingsValue"`
	NewSettingsValue string `json:"newSettingsValue"`
	MyLikeValue      int    `json:"myLikeValue"`
	LikeCount        int    `json:"likeCount"`
}

// Mastery is a page you should have mastered before you can understand another page.
type Mastery struct {
	PageId    string `json:"pageId"`
	Has       bool   `json:"has"`
	Wants     bool   `json:"wants"`
	UpdatedAt string `json:"updatedAt"`
}

// Mark is something attached to a page, e.g. a place where a user said they were confused.
type Mark struct {
	Id                  string `json:"id"`
	PageId              string `json:"pageId"`
	Type                string `json:"type"`
	IsCurrentUserOwned  bool   `json:"isCurrentUserOwned"`
	CreatedAt           string `json:"createdAt"`
	AnchorContext       string `json:"anchorContext"`
	AnchorText          string `json:"anchorText"`
	AnchorOffset        int    `json:"anchorOffset"`
	Text                string `json:"text"`
	RequisiteSnapshotId string `json:"requisiteSnapshotId"`
	ResolvedPageId      string `json:"resolvedPageId"`
	Answered            bool   `json:"answered"`

	// If the mark was resolved by the owner, we want to display that. But that also
	// means we can't send the ResolvedBy value to the FE, so we use IsResolveByOwner instead.
	IsResolvedByOwner bool   `json:"isResolvedByOwner"`
	ResolvedBy        string `json:"resolvedBy"`

	// Marks are anonymous, so the only info FE gets is whether this mark is owned
	// by the current user.
	CreatorId string `json:"-"`
}

// PageObject stores some information for an object embedded in a page
type PageObject struct {
	PageId string `json:"pageId"`
	Edit   int    `json:"edit"`
	Object string `json:"object"`
	Value  string `json:"value"`
}

// Answer is attached to a question page, and points to another page that
// answers the question.
type Answer struct {
	Id           int64  `json:"id,string"`
	QuestionId   string `json:"questionId"`
	AnswerPageId string `json:"answerPageId"`
	UserId       string `json:"userId"`
	CreatedAt    string `json:"createdAt"`
}

// SearchString is attached to a question page to help with directing users
// towards it via search or marks.
type SearchString struct {
	Id     int64
	PageId string
	Text   string
}

// CommonHandlerData is what handlers fill out and return
type CommonHandlerData struct {
	// If set, then this packet will erase all data on the FE
	ResetEverything bool `json:"resetEverything"`
	// Optional user object with the current user's data
	User *CurrentUser `json:"user"`
	// Map of page id -> currently live version of the page
	PageMap map[string]*Page `json:"pages"`
	// Map of page id -> some edit of the page
	EditMap    map[string]*Page    `json:"edits"`
	UserMap    map[string]*User    `json:"users"`
	MasteryMap map[string]*Mastery `json:"masteries"`
	MarkMap    map[string]*Mark    `json:"marks"`
	// Page id -> {object alias -> object}
	PageObjectMap map[string]map[string]*PageObject `json:"pageObjects"`
	// ResultMap contains various data the specific handler returns
	ResultMap map[string]interface{} `json:"result"`
}

// NewHandlerData creates and initializes a new commonHandlerData object.
func NewHandlerData(u *CurrentUser) *CommonHandlerData {
	var data CommonHandlerData
	data.User = u
	data.PageMap = make(map[string]*Page)
	data.EditMap = make(map[string]*Page)
	data.UserMap = make(map[string]*User)
	data.MasteryMap = make(map[string]*Mastery)
	data.MarkMap = make(map[string]*Mark)
	data.PageObjectMap = make(map[string]map[string]*PageObject)
	data.ResultMap = make(map[string]interface{})
	return &data
}

func (data CommonHandlerData) AddMark(idStr string) {
	if _, ok := data.MarkMap[idStr]; !ok {
		data.MarkMap[idStr] = &Mark{Id: idStr}
	}
}

func (data CommonHandlerData) SetResetEverything() *CommonHandlerData {
	data.ResetEverything = true
	return &data
}

// LoadDataOption is used to set some simple loading options for loading functions
type LoadDataOptions struct {
	// If set, we'll only load links for the pages with these ids
	ForPages map[string]*Page
}

// ExecuteLoadPipeline runs the pages in the pageMap through the pipeline to load
// all the data they need.
func ExecuteLoadPipeline(db *database.DB, data *CommonHandlerData) error {
	u := data.User
	pageMap := data.PageMap
	userMap := data.UserMap
	masteryMap := data.MasteryMap
	pageObjectMap := data.PageObjectMap
	markMap := data.MarkMap

	// Load comments
	filteredPageMap := filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Comments })
	err := LoadCommentIds(db, u, pageMap, &LoadDataOptions{ForPages: filteredPageMap})
	if err != nil {
		return fmt.Errorf("LoadCommentIds for failed: %v", err)
	}

	// Load questions
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Questions })
	err = LoadChildIds(db, pageMap, u, &LoadChildIdsOptions{
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
	err = LoadChildIds(db, pageMap, u, &LoadChildIdsOptions{
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
	err = LoadParentIds(db, pageMap, u, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: ParentPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for parents failed: %v", err)
	}

	// Load tags
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Tags })
	err = LoadParentIds(db, pageMap, u, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: TagPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for tags failed: %v", err)
	}

	// Load related
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Related })
	err = LoadChildIds(db, pageMap, u, &LoadChildIdsOptions{
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
	err = LoadChildIds(db, pageMap, u, &LoadChildIdsOptions{
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
	err = LoadParentIds(db, pageMap, u, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: RequirementPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
		MasteryMap:   masteryMap,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for requirements failed: %v", err)
	}

	// Load subjects
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Subjects })
	err = LoadParentIds(db, pageMap, u, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: SubjectPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
		MasteryMap:   masteryMap,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for subjects failed: %v", err)
	}

	// Load answers
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Answers })
	err = LoadAnswers(db, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadAnswers failed: %v", err)
	}

	// Load user's marks
	if u.Id != "" {
		filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.UserMarks })
		err = LoadMarkIds(db, u, pageMap, markMap, &LoadMarkIdsOptions{
			ForPages:              filteredPageMap,
			CurrentUserConstraint: true,
		})
		if err != nil {
			return fmt.Errorf("LoadMarkIds for user's marks failed: %v", err)
		}

		// Load unresolved marks
		filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.UnresolvedMarks })
		err = LoadMarkIds(db, u, pageMap, markMap, &LoadMarkIdsOptions{
			ForPages:         filteredPageMap,
			EditorConstraint: true,
		})
		if err != nil {
			return fmt.Errorf("LoadMarkIds for unresolved marks failed: %v", err)
		}

		// Load all marks if forced to
		filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.AllMarks })
		err = LoadMarkIds(db, u, pageMap, markMap, &LoadMarkIdsOptions{
			ForPages:        filteredPageMap,
			LoadResolvedToo: true,
		})
		if err != nil {
			return fmt.Errorf("LoadMarkIds for all marks failed: %v", err)
		}
	}

	// Load data for all marks in the map
	err = LoadMarkData(db, pageMap, userMap, markMap, u)
	if err != nil {
		return fmt.Errorf("LoadMarkData failed: %v", err)
	}

	// Load what domains these users belong to
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.DomainMembership })
	err = LoadDomainMemberships(db, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadDomainMemberships failed: %v", err)
	}

	// Load links
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Links })
	err = LoadLinks(db, u, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLinks failed: %v", err)
	}

	// Load domains
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.DomainsAndPermissions })
	err = LoadDomainIds(db, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadSummaries failed: %v", err)
	}

	// Load change logs
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ChangeLogs })
	err = LoadChangeLogs(db, u.Id, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadChangeLogs failed: %v", err)
	}

	// Load search strings
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SearchStrings })
	err = LoadSearchStrings(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSearchStrings failed: %v", err)
	}

	// Load whether or not the pages have an unpublished draft
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.HasDraft })
	err = LoadDraftExistence(db, u.Id, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadDraftExistence failed: %v", err)
	}

	// Load (dis)likes
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Likes })
	err = LoadLikes(db, u, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadLikes failed: %v", err)
	}

	// Load views
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ViewCount })
	err = LoadViewCounts(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadViewCounts failed: %v", err)
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

	// Load subscriber count
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SubscriberCount })
	err = LoadSubscriberCount(db, u.Id, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSubscriberCount failed: %v", err)
	}

	// Load subpage counts
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SubpageCounts })
	err = LoadSubpageCounts(db, u, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSubpageCounts failed: %v", err)
	}

	// Load answer counts
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.AnswerCounts })
	err = LoadAnswerCounts(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadAnswerCounts failed: %v", err)
	}

	// Load number of red links.
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.RedLinkCount })
	err = LoadRedLinkCount(db, u, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadRedLinkCount failed: %v", err)
	}

	// Load whether the page is used as a requirement
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.UsedAsMastery })
	err = LoadUsedAsMastery(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadUsedAsMastery failed: %v", err)
	}

	// Load pages' creator's ids
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Creators })
	err = LoadCreatorIds(db, u, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadCreatorIds failed: %v", err)
	}

	// Add other pages we'll need
	AddUserGroupIdsToPageMap(u, pageMap)

	// Load page data
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return !p.LoadOptions.Edit && !p.LoadOptions.IncludeDeleted })
	err = LoadPages(db, u, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadPages failed: %v", err)
	}

	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return !p.LoadOptions.Edit && p.LoadOptions.IncludeDeleted })
	err = LoadPagesWithOptions(db, u, filteredPageMap, PageInfosTableWithOptions(u, &PageInfosOptions{Deleted: true}))
	if err != nil {
		return fmt.Errorf("LoadPages (deleted) failed: %v", err)
	}

	// Compute edit permissions
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.DomainsAndPermissions })
	ComputePermissionsForMap(db.C, filteredPageMap, u)

	// Load summaries
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Summaries })
	err = LoadSummaries(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSummaries failed: %v", err)
	}

	// Load incoming mark count
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.LinkedMarkCount && p.Type == QuestionPageType })
	err = LoadLinkedMarkCounts(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadLinkedMarkCounts failed: %v", err)
	}

	// Load prev/next ids
	/*filteredPageMap = filterPageMap(pageMap, func(p *Page) bool {
		return p.Type == WikiPageType && p.LoadOptions.NextPrevIds
	})
	err = LoadNextPrevPageIds(db, u.Id, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadNextPrevPageIds failed: %v", err)
	}*/

	// Load page objects
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.PageObjects })
	err = LoadPageObjects(db, u, filteredPageMap, pageObjectMap)
	if err != nil {
		return fmt.Errorf("LoadPageObject failed: %v", err)
	}

	// Add pages that need a corresponding mastery to the masteryMap
	for _, p := range pageMap {
		if p.LoadOptions.Mastery {
			masteryMap[p.PageId] = &Mastery{PageId: p.PageId}
		}
	}

	// Load what requirements the user has met
	err = LoadMasteries(db, u, masteryMap)
	if err != nil {
		return fmt.Errorf("LoadMasteries failed: %v", err)
	}

	// Load all the users
	userMap[u.Id] = &User{Id: u.Id}
	for _, p := range pageMap {
		if p.LoadOptions.Text || p.LoadOptions.Summaries {
			userMap[p.CreatorId] = &User{Id: p.CreatorId}
		}
		if IsIdValid(p.LockedBy) {
			userMap[p.LockedBy] = &User{Id: p.LockedBy}
		}
	}
	err = LoadUsers(db, userMap, u.Id)
	if err != nil {
		return fmt.Errorf("LoadUsers failed: %v", err)
	}

	// Computed which pages count as visited.
	visitedValues := make([]interface{}, 0)
	visitorId := u.GetSomeId()
	if visitorId != "" {
		for pageId, p := range pageMap {
			if p.Text != "" {
				visitedValues = append(visitedValues, visitorId, u.SessionId, db.C.R.RemoteAddr, pageId, database.Now())
			}
		}
	}

	// Add a visit to pages for which we loaded text.
	if len(visitedValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO visits (userId, sessionId, ipAddress, pageId, createdAt)
			VALUES ` + database.ArgsPlaceholder(len(visitedValues), 5))
		if _, err = statement.Exec(visitedValues...); err != nil {
			return fmt.Errorf("Couldn't update visits", err)
		}
	}

	return nil
}

// LoadMasteries loads the masteries.
func LoadMasteries(db *database.DB, u *CurrentUser, masteryMap map[string]*Mastery) error {
	userId := u.GetSomeId()
	if userId == "" {
		return nil
	}

	rows := database.NewQuery(`
		SELECT masteryId,updatedAt,has,wants
		FROM userMasteryPairs
		WHERE userId=?`, userId).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var mastery Mastery
		err := rows.Scan(&mastery.PageId, &mastery.UpdatedAt, &mastery.Has, &mastery.Wants)
		if err != nil {
			return fmt.Errorf("failed to scan for mastery: %v", err)
		}
		masteryMap[mastery.PageId] = &mastery
		return nil
	})
	return err
}

// LoadPageObjects loads all the page objects necessary for the given pages.
func LoadPageObjects(db *database.DB, u *CurrentUser, pageMap map[string]*Page, pageObjectMap map[string]map[string]*PageObject) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)

	userId := u.GetSomeId()
	if userId == "" {
		return nil
	}

	rows := database.NewQuery(`
		SELECT pageId,edit,object,value
		FROM userPageObjectPairs
		WHERE userId=?`, userId).Add(`AND pageId IN `).AddArgsGroup(pageIds).Add(`
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var obj PageObject
		err := rows.Scan(&obj.PageId, &obj.Edit, &obj.Object, &obj.Value)
		if err != nil {
			return fmt.Errorf("Failed to scan for user: %v", err)
		}
		if _, ok := pageObjectMap[obj.PageId]; !ok {
			pageObjectMap[obj.PageId] = make(map[string]*PageObject)
		}
		pageObjectMap[obj.PageId][obj.Object] = &obj
		return nil
	})
	return err
}

// LoadPages loads the given pages.
func LoadPages(db *database.DB, u *CurrentUser, pageMap map[string]*Page) error {
	return LoadPagesWithOptions(db, u, pageMap, PageInfosTable(u))
}

func LoadPagesWithOptions(db *database.DB, u *CurrentUser, pageMap map[string]*Page, pageInfosTable *database.QueryPart) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)

	// Compute pages for which to load text / summary
	textIds := make([]interface{}, 0)
	for _, p := range pageMap {
		if p.LoadOptions.Text {
			textIds = append(textIds, p.PageId)
		}
	}
	textSelect := database.NewQuery(`IF(p.pageId IN`).AddIdsGroup(textIds).Add(`,p.text,"") AS text`)

	// Load the page data
	rows := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,p.creatorId,p.createdAt,p.title,p.clickbait,`).AddPart(textSelect).Add(`,
			length(p.text),p.metaText,pi.type,pi.hasVote,pi.voteType,
			pi.alias,pi.createdAt,pi.createdBy,pi.sortChildrenBy,pi.seeGroupId,pi.editGroupId,
			pi.lensIndex,pi.isEditorComment,pi.isEditorCommentIntention,pi.isRequisite,pi.indirectTeacher,
			p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,pi.isDeleted,pi.mergedInto,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset
		FROM pages AS p
		JOIN`).AddPart(pageInfosTable).Add(`AS pi
		ON (p.pageId = pi.pageId AND p.edit = pi.currentEdit)
		WHERE p.pageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p corePageData
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.PrevEdit, &p.CreatorId, &p.CreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.Type, &p.HasVote,
			&p.VoteType, &p.Alias, &p.OriginalCreatedAt, &p.OriginalCreatedBy, &p.SortChildrenBy,
			&p.SeeGroupId, &p.EditGroupId, &p.LensIndex, &p.IsEditorComment, &p.IsEditorCommentIntention,
			&p.IsRequisite, &p.IndirectTeacher,
			&p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit, &p.IsDeleted, &p.MergedInto,
			&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("Failed to scan a page: %v", err)
		}
		pageMap[p.PageId].corePageData = p
		pageMap[p.PageId].WasPublished = true
		return nil
	})
	for _, p := range pageMap {
		if p.Type == "" {
			delete(pageMap, p.PageId)
		}
	}
	return err
}

// LoadSummaries loads summaries for the given pages.
func LoadSummaries(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT pageId,name,text
		FROM pageSummaries
		WHERE pageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		var name, text string
		err := rows.Scan(&pageId, &name, &text)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[pageId].Summaries[name] = text
		return nil
	})
	return err
}

// LoadLinkedMarkCounts loads the number of marks that link to these questions.
func LoadLinkedMarkCounts(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT resolvedPageId,SUM(1)
		FROM marks
		WHERE resolvedPageId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var resolvedPageId string
		var count int
		err := rows.Scan(&resolvedPageId, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[resolvedPageId].LinkedMarkCount = count
		return nil
	})
	return err
}

// LoadChangeLogs loads the edit history for the given page.
func LoadChangeLogs(db *database.DB, userId string, pageMap map[string]*Page, userMap map[string]*User, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	for _, p := range sourcePageMap {
		p.ChangeLogs = make([]*ChangeLog, 0)
		rows := database.NewQuery(`
		SELECT id,userId,edit,type,createdAt,auxPageId,oldSettingsValue,newSettingsValue
		FROM changeLogs
		WHERE pageId=?`, p.PageId).Add(`
			AND (userId=? OR type!=?)`, userId, NewSnapshotChangeLog).Add(`
		ORDER BY createdAt DESC`).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var l ChangeLog
			err := rows.Scan(&l.Id, &l.UserId, &l.Edit, &l.Type, &l.CreatedAt, &l.AuxPageId, &l.OldSettingsValue, &l.NewSettingsValue)
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
			AddPageToMap(log.AuxPageId, pageMap, TitlePlusIncludeDeletedLoadOptions)
		}

		err = LoadLikesForChangeLogs(db, userId, p.ChangeLogs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Load LikeCount and MyLikeValue for a set of ChangeLogs
func LoadLikesForChangeLogs(db *database.DB, currentUserId string, changeLogs []*ChangeLog) error {
	if len(changeLogs) == 0 {
		return nil
	}

	changeLogMap := make(map[int]*ChangeLog)
	changeLogIds := make([]interface{}, 0)
	for _, changeLog := range changeLogs {
		changeLogMap[changeLog.Id] = changeLog
		changeLogIds = append(changeLogIds, changeLog.Id)
		changeLog.MyLikeValue = 0
		changeLog.LikeCount = 0
	}

	rows := database.NewQuery(`
		SELECT cl.id, l.userId,l.value
		FROM likes as l
		JOIN changeLogs as cl
		ON cl.likeableId=l.likeableId
		WHERE cl.id IN`).AddArgsGroup(changeLogIds).ToStatement(db).Query()

	var changeLogId int
	var value int
	var likeUserId string
	var changeLog *ChangeLog
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		err := rows.Scan(&changeLogId, &likeUserId, &value)
		if err != nil {
			return fmt.Errorf("failed to load likes for a changelog: %v", err)
		}
		changeLog = changeLogMap[changeLogId]
		if likeUserId == currentUserId {
			changeLog.MyLikeValue = value
		} else {
			changeLog.LikeCount += value
		}
		return nil
	})
	return err
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
func LoadFullEdit(db *database.DB, pageId string, u *CurrentUser, options *LoadEditOptions) (*Page, error) {
	if options == nil {
		options = &LoadEditOptions{}
	}
	p := NewPage(pageId)

	whereClause := database.NewQuery("p.isLiveEdit")
	if options.LoadSpecificEdit > 0 {
		whereClause = database.NewQuery("p.edit=?", options.LoadSpecificEdit)
	} else if options.CreatedAtLimit != "" {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND createdAt<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.CreatedAtLimit)
	} else if options.LoadEditWithLimit > 0 {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND edit<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.LoadEditWithLimit)
	} else if options.LoadNonliveEdit {
		// Load the most recent edit we have for the current user.
		// If there is an autosave, load that.
		// If there is are snapshots, only consider those that are based off of the currently live edit.
		// Otherwise, load the currently live edit.
		// Note: "z" is just a hack to make sure autosave is sorted to the top.
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT p.edit
				FROM pages AS p
				JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
				ON (p.pageId=pi.pageId)
				WHERE p.pageId=?`, pageId).Add(`
					AND (p.prevEdit=pi.currentEdit OR p.edit=pi.currentEdit OR p.isAutosave) AND
					(p.creatorId=? OR NOT (p.isSnapshot OR p.isAutosave))`, u.Id).Add(`
				ORDER BY IF(p.isAutosave,"z",p.createdAt) DESC
				LIMIT 1
			)`)
	}
	statement := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,pi.type,p.title,p.clickbait,p.text,p.metaText,
			pi.alias,p.creatorId,pi.sortChildrenBy,pi.hasVote,pi.voteType,
			p.createdAt,pi.seeGroupId,pi.editGroupId,pi.createdAt,
			pi.createdBy,pi.lensIndex,pi.isEditorComment,pi.isEditorCommentIntention,
			p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset,
			pi.currentEdit>0,pi.isDeleted,pi.mergedInto,pi.currentEdit,pi.maxEdit,pi.lockedBy,pi.lockedUntil,
			pi.voteType,pi.isRequisite,pi.indirectTeacher
		FROM pages AS p
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (p.pageId=pi.pageId AND p.pageId=?)`, pageId).Add(`
		WHERE`).AddPart(whereClause).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageId, &p.Edit, &p.PrevEdit, &p.Type, &p.Title, &p.Clickbait,
		&p.Text, &p.MetaText, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.SeeGroupId,
		&p.EditGroupId, &p.OriginalCreatedAt, &p.OriginalCreatedBy, &p.LensIndex,
		&p.IsEditorComment, &p.IsEditorCommentIntention, &p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit,
		&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset, &p.WasPublished,
		&p.IsDeleted, &p.MergedInto, &p.CurrentEdit, &p.MaxEditEver, &p.LockedBy, &p.LockedUntil, &p.LockedVoteType,
		&p.IsRequisite, &p.IndirectTeacher)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}

	if exists {
		err = LoadDomainIdsForPage(db, p)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load domain ids for page: %v", err)
		}
	}
	p.ComputePermissions(db.C, u)

	p.TextLength = len(p.Text)
	return p, nil
}

// LoadPageIds from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func LoadPageIds(rows *database.Rows, pageMap map[string]*Page, loadOptions *PageLoadOptions) ([]string, error) {
	ids := make([]string, 0)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}

		AddPageToMap(pageId, pageMap, loadOptions)
		ids = append(ids, pageId)
		return nil
	})
	return ids, err
}

// LoadLikes loads likes corresponding to the given pages and updates the pages.
func LoadLikes(db *database.DB, u *CurrentUser, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT l.userId,pi.pageId,l.value
		FROM likes as l
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON l.likeableId=pi.likeableId
		WHERE pi.pageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId string
		var pageId string
		var value int
		err := rows.Scan(&userId, &pageId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a like: %v", err)
		}
		page := pageMap[pageId]
		// We count the current user's like value towards the sum here in the FE.
		if userId == u.Id {
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

// LoadSearchStrings loads all the search strings for the given pages
func LoadSearchStrings(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT id,pageId,text
		FROM searchStrings
		WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds))).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var id int64
		var pageId, text string
		err := rows.Scan(&id, &pageId, &text)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		pageMap[pageId].SearchStrings[fmt.Sprintf("%d", id)] = text
		return nil
	})
	return err
}

// LoadSearchString loads just one search string given its id
func LoadSearchString(db *database.DB, id string) (*SearchString, error) {
	var searchString SearchString
	rows := db.NewStatement(`
		SELECT id,pageId,text
		FROM searchStrings
		WHERE id=?`).Query(id)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		err := rows.Scan(&searchString.Id, &searchString.PageId, &searchString.Text)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		return nil
	})
	return &searchString, err
}

// LoadViewCounts loads view counts for the pages
func LoadViewCounts(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT pageId,count(distinct userId)
		FROM visits
		WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
		GROUP BY 1`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		var count int
		err := rows.Scan(&pageId, &count)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		pageMap[pageId].ViewCount = count
		return nil
	})
	return err
}

// LoadVotes loads probability votes corresponding to the given pages and updates the pages.
func LoadVotes(db *database.DB, currentUserId string, pageMap map[string]*Page, usersMap map[string]*User) error {
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
		var pageId string
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
func LoadRedLinkCount(db *database.DB, u *CurrentUser, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsList := PageIdsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT l.parentId,SUM(ISNULL(pi.pageId))
		FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
		RIGHT JOIN links AS l
		ON (pi.pageId=l.childAlias OR pi.alias=l.childAlias)
		WHERE l.parentId IN`).AddArgsGroup(pageIdsList).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId string
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

// LoadUsedAsMastery loads if the page is ever used as mastery/requirement.
func LoadUsedAsMastery(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsList := PageIdsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT parentId,count(*)
		FROM pagePairs
		WHERE type=?`, RequirementPagePairType).Add(`
			AND parentId IN`).AddArgsGroup(pageIdsList).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId string
		var count int
		err := rows.Scan(&parentId, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[parentId].UsedAsMastery = count > 0
		return nil
	})
	return err
}

// LoadCreatorIds loads creator ids for the pages
func LoadCreatorIds(db *database.DB, u *CurrentUser, pageMap map[string]*Page, userMap map[string]*User, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}
	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}
	if len(sourceMap) <= 0 {
		return nil
	}
	pageIdsList := PageIdsListFromMap(sourceMap)

	rows := database.NewQuery(`
		SELECT pageId,creatorId,COUNT(*)
		FROM pages
		WHERE pageId IN`).AddArgsGroup(pageIdsList).Add(`
			AND NOT isAutosave AND NOT isSnapshot AND NOT isMinorEdit
		GROUP BY 1,2
		ORDER BY 3 DESC`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, creatorId string
		var count int
		err := rows.Scan(&pageId, &creatorId, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[pageId].CreatorIds = append(pageMap[pageId].CreatorIds, creatorId)
		userMap[creatorId] = &User{Id: creatorId}
		return nil
	})
	if err != nil {
		return err
	}

	// For pages that have editGroupId set, make sure we load those pages too.
	rows = database.NewQuery(`
		SELECT pi.pageId,pi.editGroupId,ISNULL(u.id)
		FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
		LEFT JOIN users AS u
		ON (pi.pageId = u.id)
		WHERE pi.pageId IN`).AddArgsGroup(pageIdsList).Add(`
			AND pi.editGroupId!=""`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, editGroupId string
		var isPage bool
		err := rows.Scan(&pageId, &editGroupId, &isPage)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		if isPage {
			AddPageToMap(editGroupId, pageMap, TitlePlusLoadOptions)
		} else {
			userMap[editGroupId] = &User{Id: editGroupId}
		}
		return nil
	})
	return err
}

// LoadLinks loads the links for the given pages, and adds them to the pageMap.
func LoadLinks(db *database.DB, u *CurrentUser, pageMap map[string]*Page, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}

	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}

	pageIds := PageIdsListFromMap(sourceMap)
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
		var parentId string
		var childAlias string
		err := rows.Scan(&parentId, &childAlias)
		if err != nil {
			return fmt.Errorf("failed to scan for a link: %v", err)
		}
		if IsIdValid(childAlias) {
			AddPageIdToMap(childAlias, pageMap)
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
			FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
			WHERE alias IN`).AddArgsGroup(aliasesList).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId string
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

// LoadAnswers loads the answers for the given pages, and adds the corresponding pages to the pageMap.
func LoadAnswers(db *database.DB, pageMap map[string]*Page, userMap map[string]*User, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}

	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}

	pageIds := PageIdsListFromMap(sourceMap)
	if len(pageIds) <= 0 {
		return nil
	}

	rows := db.NewStatement(`
		SELECT id,questionId,answerPageId,userId,createdAt
		FROM answers
		WHERE questionId IN ` + database.InArgsPlaceholder(len(pageIds))).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var answer Answer
		err := rows.Scan(&answer.Id, &answer.QuestionId, &answer.AnswerPageId, &answer.UserId, &answer.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		AddPageToMap(answer.AnswerPageId, pageMap, AnswerLoadOptions)
		userMap[answer.UserId] = &User{Id: answer.UserId}
		pageMap[answer.QuestionId].Answers = append(pageMap[answer.QuestionId].Answers, &answer)
		return nil
	})
	return err
}

// LoadAnswer just loads the data for one specific answer.
func LoadAnswer(db *database.DB, answerId string) (*Answer, error) {
	var answer Answer
	_, err := db.NewStatement(`
		SELECT id,questionId,answerPageId,userId,createdAt
		FROM answers
		WHERE id=?`).QueryRow(answerId).Scan(&answer.Id, &answer.QuestionId,
		&answer.AnswerPageId, &answer.UserId, &answer.CreatedAt)
	return &answer, err
}

type LoadMarkIdsOptions struct {
	// If set, we'll only load links for the pages with these ids
	ForPages map[string]*Page

	// If set, only load marks owned by the current user
	CurrentUserConstraint bool
	// If true, load unresolved marks only iff you are an editor
	EditorConstraint bool
	// If true, load resolved marks too
	LoadResolvedToo bool
}

// LoadMarkIds loads all the marks owned by the given user
func LoadMarkIds(db *database.DB, u *CurrentUser, pageMap map[string]*Page, markMap map[string]*Mark, options *LoadMarkIdsOptions) error {
	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}

	pageIds := PageIdsListFromMap(sourceMap)
	if len(pageIds) <= 0 {
		return nil
	}

	userConstraint := database.NewQuery(``)
	if options.CurrentUserConstraint {
		userConstraint = database.NewQuery(`AND m.creatorId=?`, u.Id)
	}
	pageIdsPart := database.NewQuery(``).AddArgsGroup(pageIds)
	if options.EditorConstraint {
		pageIdsPart = database.NewQuery(`(
			SELECT p.pageId
			FROM pages AS p
			WHERE p.pageId IN`).AddArgsGroup(pageIds).Add(`
				AND NOT p.isSnapshot AND NOT p.isAutosave
				AND p.creatorId=?`, u.Id).Add(`
		)`)
	}
	resolvedConstraint := database.NewQuery(`AND m.resolvedBy=""`)
	if options.LoadResolvedToo {
		resolvedConstraint = database.NewQuery(``)
	}

	rows := database.NewQuery(`
		SELECT m.id
		FROM marks AS m
		WHERE m.pageId IN`).AddPart(pageIdsPart).Add(`
			`).AddPart(userConstraint).AddPart(resolvedConstraint).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var markId string
		err := rows.Scan(&markId)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		if _, ok := markMap[markId]; !ok {
			markMap[markId] = &Mark{Id: markId}
		}
		return nil
	})
	return err
}

// LoadMarkData loads all the relevant data for each mark
func LoadMarkData(db *database.DB, pageMap map[string]*Page, userMap map[string]*User, markMap map[string]*Mark, u *CurrentUser) error {
	if len(markMap) <= 0 {
		return nil
	}
	markIds := make([]interface{}, 0)
	for markId, _ := range markMap {
		markIds = append(markIds, markId)
	}

	rows := db.NewStatement(`
		SELECT id,type,pageId,creatorId,createdAt,anchorContext,anchorText,anchorOffset,
			text,requisiteSnapshotId,resolvedPageId,resolvedBy,answered
		FROM marks
		WHERE id IN` + database.InArgsPlaceholder(len(markIds))).Query(markIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var creatorId string
		mark := &Mark{}
		err := rows.Scan(&mark.Id, &mark.Type, &mark.PageId, &mark.CreatorId, &mark.CreatedAt,
			&mark.AnchorContext, &mark.AnchorText, &mark.AnchorOffset, &mark.Text,
			&mark.RequisiteSnapshotId, &mark.ResolvedPageId, &mark.ResolvedBy, &mark.Answered)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		mark.IsCurrentUserOwned = mark.CreatorId == u.Id
		if mark.CreatorId == mark.ResolvedBy {
			mark.ResolvedBy = "0"
			mark.IsResolvedByOwner = true
		}
		*markMap[mark.Id] = *mark

		if page, ok := pageMap[mark.PageId]; ok {
			page.MarkIds = append(page.MarkIds, mark.Id)
		}
		AddPageToMap(mark.PageId, pageMap, TitlePlusLoadOptions)
		if mark.ResolvedPageId != "" {
			AddPageToMap(mark.ResolvedPageId, pageMap, TitlePlusLoadOptions)
			userMap[mark.ResolvedBy] = &User{Id: mark.ResolvedBy}
		}
		userMap[creatorId] = &User{Id: creatorId}
		return nil
	})
	return err
}

type LoadChildIdsOptions struct {
	// If set, the children will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[string]*Page
	// Type of children to load
	Type string
	// Type of the child relationship to follow
	PagePairType string
	// Load options to set for the new pages
	LoadOptions *PageLoadOptions
}

// LoadChildIds loads the page ids for all the children of the pages in the given pageMap.
func LoadChildIds(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadChildIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pairTypeFilter := ""
	if options.PagePairType != "" {
		pairTypeFilter = "type = '" + options.PagePairType + "' AND"
	}

	pageTypeFilter := ""
	if options.Type != "" {
		pageTypeFilter = "AND pi.type = '" + options.Type + "'"
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type,pi.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE`).Add(pairTypeFilter).Add(`parentId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON pi.pageId=pp.childId`).Add(pageTypeFilter).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		var ppType string
		var piType string
		err := rows.Scan(&parentId, &childId, &ppType, &piType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(childId, pageMap, options.LoadOptions)

		parent := sourcePageMap[parentId]
		if piType == LensPageType {
			parent.LensIds = append(parent.LensIds, newPage.PageId)
			newPage.ParentIds = append(newPage.ParentIds, parent.PageId)
		} else if piType == CommentPageType {
			parent.CommentIds = append(parent.CommentIds, newPage.PageId)
		} else if piType == QuestionPageType {
			parent.QuestionIds = append(parent.QuestionIds, newPage.PageId)
		} else if piType == WikiPageType && ppType == ParentPagePairType {
			parent.ChildIds = append(parent.ChildIds, childId)
			parent.HasChildren = true
			if parent.LoadOptions.HasGrandChildren {
				newPage.LoadOptions.SubpageCounts = true
			}
			if parent.LoadOptions.RedLinkCountForChildren {
				newPage.LoadOptions.RedLinkCount = true
			}
		} else if piType == WikiPageType && ppType == TagPagePairType {
			parent.RelatedIds = append(parent.RelatedIds, childId)
		}
		return nil
	})
	return err
}

// LoadSubpageCounts loads the number of various types of children the pages have
func LoadSubpageCounts(db *database.DB, u *CurrentUser, pageMap map[string]*Page) error {
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
		) AS pp
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON (pi.pageId=pp.childId)
		GROUP BY 1,2`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		var childType string
		var count int
		err := rows.Scan(&pageId, &childType, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		if childType == CommentPageType {
			pageMap[pageId].CommentCount = count
		}
		return nil
	})
	return err
}

// LoadAnswerCounts loads the number of answers the pages have
func LoadAnswerCounts(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT questionId,sum(1)
		FROM answers
		WHERE questionId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var questionId string
		var count int
		err := rows.Scan(&questionId, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[questionId].AnswerCount = count

		return nil
	})
	return err
}

// LoadTaggedAsIds for each page in the source map loads the ids of the pages that tag it.
func LoadTaggedAsIds(db *database.DB, pageMap map[string]*Page, options *LoadChildIdsOptions) error {
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
		var parentId, childId string
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		child := sourcePageMap[childId]
		child.TaggedAsIds = append(child.TaggedAsIds, parentId)
		AddPageIdToMap(parentId, pageMap)
		return nil
	})
	return err
}

type LoadParentIdsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[string]*Page
	// Load whether or not each parent has parents of its own.
	LoadHasParents bool
	// Type of the parent relationship to follow
	PagePairType string
	// Load options to set for the new pages
	LoadOptions *PageLoadOptions
	// Mastery map to populate with masteries necessary for a requirement
	MasteryMap map[string]*Mastery
}

// LoadParentIds loads the page ids for all the parents of the pages in the given pageMap.
func LoadParentIds(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadParentIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pairTypeFilter := ""
	if options.PagePairType != "" {
		pairTypeFilter = "type = '" + options.PagePairType + "' AND"
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	newPages := make(map[string]*Page)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE `).Add(pairTypeFilter).Add(`childId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (pi.pageId=pp.parentId)
		WHERE (pi.currentEdit>0 AND NOT pi.isDeleted) OR pp.parentId=pp.childId
		`).ToStatement(db).Query()

	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		var ppType string
		err := rows.Scan(&parentId, &childId, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(parentId, pageMap, options.LoadOptions)
		childPage := sourcePageMap[childId]

		if ppType == ParentPagePairType {
			childPage.ParentIds = append(childPage.ParentIds, parentId)
			childPage.HasParents = true
			newPages[newPage.PageId] = newPage
		} else if ppType == RequirementPagePairType {
			childPage.RequirementIds = append(childPage.RequirementIds, parentId)
			options.MasteryMap[parentId] = &Mastery{PageId: parentId}
		} else if ppType == TagPagePairType {
			childPage.TaggedAsIds = append(childPage.TaggedAsIds, parentId)
		} else if ppType == SubjectPagePairType {
			childPage.SubjectIds = append(childPage.SubjectIds, parentId)
			options.MasteryMap[parentId] = &Mastery{PageId: parentId}
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
			var pageId string
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
func LoadCommentIds(db *database.DB, u *CurrentUser, pageMap map[string]*Page, options *LoadDataOptions) error {
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
			JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
			ON (pi.pageId=pp.childId)
			WHERE pi.type=?`, CommentPageType).Add(`
				AND pp.type=?`, ParentPagePairType).Add(`
				AND pp.parentId IN`).AddArgsGroup(pageIds).Add(`
		)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		parentPage := AddPageToMap(parentId, pageMap, SubpageLoadOptions)
		childPage := AddPageToMap(childId, pageMap, SubpageLoadOptions)
		parentPage.CommentIds = append(parentPage.CommentIds, childPage.PageId)
		return nil
	})

	// Now we have pages that have in comment ids both top level comments and
	// replies, so we need to remove the replies.
	for _, p := range sourcePageMap {
		replies := make(map[string]bool)
		for _, c := range p.CommentIds {
			for _, r := range pageMap[c].CommentIds {
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
func loadOrderedChildrenIds(db *database.DB, u *CurrentUser, parentId string, sortType string) ([]string, error) {
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
	childrenIds := make([]string, 0)
	rows := database.NewQuery(`
		SELECT pp.childId
		FROM pagePairs AS pp
		JOIN pages AS p
		ON (pp.childId=p.pageId)
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON (pi.pageId=p.pageId)
		WHERE p.isLiveEdit
			AND pi.type!=? AND pi.type!=? AND pi.type!=?`, CommentPageType, QuestionPageType, LensPageType).Add(`
			AND pp.type=?`, ParentPagePairType).Add(`AND pp.parentId=?`, parentId).Add(`
		ORDER BY ` + orderClause).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var childId string
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
func loadSiblingId(db *database.DB, u *CurrentUser, pageId string, useNextSibling bool) (string, error) {
	// Load the parent and the sorting order
	var parentId string
	var sortType string
	var parentCount int
	row := database.NewQuery(`
		SELECT
			ifnull(max(pp.parentId), 0),
			ifnull(max(pi.sortChildrenBy), ""),
			count(*)
		FROM pagePairs AS pp
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON (pp.parentId=pi.pageId)
		WHERE pp.type=?`, ParentPagePairType).Add(`AND pp.childId=?`, pageId).ToStatement(db).QueryRow()
	found, err := row.Scan(&parentId, &sortType, &parentCount)
	if err != nil {
		return "", err
	} else if !found || parentCount != 1 {
		return "", nil
	}

	// Load the sibling pages in order
	orderedSiblingIds, err := loadOrderedChildrenIds(db, u, parentId, sortType)
	if err != nil {
		return "", fmt.Errorf("Failed to load children: %v", err)
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
			return loadSiblingId(db, u, parentId, useNextSibling)
		}
	} else {
		if pageSiblingIndex > 0 {
			return orderedSiblingIds[pageSiblingIndex-1], nil
		} else if pageSiblingIndex == 0 {
			// It's the first child, so just return the parent
			return parentId, nil
		}
	}
	return "", nil
}

// LoadNextPrevPageIds loads the pages that come before / after the given page
// in the learning list.
func LoadNextPrevPageIds(db *database.DB, u *CurrentUser, options *LoadDataOptions) error {
	if len(options.ForPages) > 1 {
		db.C.Warningf("LoadNextPrevPageIds called with more than one page")
	}
	for _, p := range options.ForPages {
		var err error
		p.PrevPageId, err = loadSiblingId(db, u, p.PageId, false)
		if err != nil {
			return fmt.Errorf("Error while loading prev page id: %v", err)
		}

		// NextPageId will be the first child if there are children
		orderedChildrenIds, err := loadOrderedChildrenIds(db, u, p.PageId, p.SortChildrenBy)
		if err != nil {
			return fmt.Errorf("Error getting first child: %v", err)
		}
		if len(orderedChildrenIds) > 0 {
			p.NextPageId = orderedChildrenIds[0]
		}

		// If there are no children, then get the next sibling
		if !IsIdValid(p.NextPageId) {
			p.NextPageId, err = loadSiblingId(db, u, p.PageId, true)
			if err != nil {
				return fmt.Errorf("Error while loading next page id: %v", err)
			}
		}
	}
	return nil
}

// LoadDraftExistence computes for each page whether or not the user has an
// autosave draft for it.
// This only makes sense to call for pages which were loaded for isLiveEdit=true.
func LoadDraftExistence(db *database.DB, userId string, options *LoadDataOptions) error {
	pageMap := options.ForPages
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId
		FROM pages
		WHERE pageId IN`).AddArgsGroup(pageIds).Add(`
			AND isAutosave AND creatorId=?`, userId).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("Failed to scan a draft existence: %v", err)
		}
		pageMap[pageId].HasDraft = true
		return nil
	})
	return err
}

// LoadLastVisits loads lastVisit variable for each page.
func LoadLastVisits(db *database.DB, currentUserId string, pageMap map[string]*Page) error {
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
		var pageId string
		var createdAt string
		err := rows.Scan(&pageId, &createdAt)
		if err != nil {
			return fmt.Errorf("Failed to scan for a comment like: %v", err)
		}
		pageMap[pageId].LastVisit = createdAt
		return nil
	})
	return err
}

// LoadSubscriptions loads subscription statuses corresponding to the given
// pages, and then updates the given maps.
func LoadSubscriptions(db *database.DB, currentUserId string, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT toId
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageId string
		err := rows.Scan(&toPageId)
		if err != nil {
			return fmt.Errorf("Failed to scan for a subscription: %v", err)
		}
		pageMap[toPageId].IsSubscribed = true
		return nil
	})
	return err
}

// LoadSubscriberCount loads number of subscribers the page has.
func LoadSubscriberCount(db *database.DB, currentUserId string, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT toId,count(*)
		FROM subscriptions
		WHERE userId!=?`, currentUserId).Add(`AND toId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageId string
		var count int
		err := rows.Scan(&toPageId, &count)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		pageMap[toPageId].SubscriberCount = count
		return nil
	})
	return err
}

// LoadDomainIds loads the domain ids for the given page and adds them to the map
func LoadDomainIds(db *database.DB, pageMap map[string]*Page, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pageId,domainId
		FROM pageDomainPairs
		WHERE pageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, domainId string
		err := rows.Scan(&pageId, &domainId)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		sourcePageMap[pageId].DomainIds = append(sourcePageMap[pageId].DomainIds, domainId)
		if pageMap != nil {
			AddPageToMap(domainId, pageMap, TitlePlusLoadOptions)
		}
		return nil
	})

	// Pages that are not part of any domain are part of the general ("") domain
	for _, p := range sourcePageMap {
		if len(p.DomainIds) <= 0 {
			p.DomainIds = append(p.DomainIds, "")
		}
	}
	return err
}
func LoadDomainIdsForPage(db *database.DB, page *Page) error {
	pageMap := map[string]*Page{page.PageId: page}
	return LoadDomainIds(db, pageMap, &LoadDataOptions{
		ForPages: pageMap,
	})
}

// LoadDomainMemberships loads which domains the users belong to
func LoadDomainMemberships(db *database.DB, pageMap map[string]*Page, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT toUserId,domainId
		FROM invites
		WHERE toUserId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId, domainId string
		err := rows.Scan(&userId, &domainId)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		sourcePageMap[userId].DomainMembershipIds = append(sourcePageMap[userId].DomainMembershipIds, domainId)
		if pageMap != nil {
			AddPageToMap(domainId, pageMap, TitlePlusLoadOptions)
		}
		return nil
	})
	return err
}

// LoadAliasToPageIdMap loads the mapping from aliases to page ids.
func LoadAliasToPageIdMap(db *database.DB, u *CurrentUser, aliases []string) (map[string]string, error) {
	aliasToIdMap := make(map[string]string)

	strictAliases := make([]string, 0)
	strictPageIds := make([]string, 0)
	for _, alias := range aliases {
		if IsIdValid(alias) {
			strictPageIds = append(strictPageIds, strings.ToLower(alias))
		} else {
			strictAliases = append(strictAliases, strings.ToLower(alias))
		}
	}

	var query *database.Stmt

	if len(strictPageIds) > 0 || len(strictAliases) > 0 {
		if len(strictPageIds) <= 0 {
			query = database.NewQuery(`
			SELECT pageId,alias
			FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
			WHERE alias IN`).AddArgsGroupStr(strictAliases).ToStatement(db)
		} else if len(strictAliases) <= 0 {
			query = database.NewQuery(`
			SELECT pageId,alias
			FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
			WHERE pageId IN`).AddArgsGroupStr(strictPageIds).ToStatement(db)
		} else {
			query = database.NewQuery(`
			SELECT pageId,alias
			FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
			WHERE pageId IN`).AddArgsGroupStr(strictPageIds).Add(`
				OR alias IN`).AddArgsGroupStr(strictAliases).ToStatement(db)
		}

		rows := query.Query()

		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId string
			var alias string
			err := rows.Scan(&pageId, &alias)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			aliasToIdMap[strings.ToLower(alias)] = pageId
			aliasToIdMap[strings.ToLower(pageId)] = pageId
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("Couldn't convert pageId=>alias", err)
		}

		// The query only gets results for when the page is published
		// We also want to return the pageIds even if they aren't for valid pages
		for _, pageId := range strictPageIds {
			aliasToIdMap[strings.ToLower(pageId)] = strings.ToLower(pageId)
		}
	}
	return aliasToIdMap, nil
}

// LoadOldAliasToPageId converts the given old (base 10) page alias to page id.
func LoadOldAliasToPageId(db *database.DB, u *CurrentUser, alias string) (string, bool, error) {
	aliasToUse := alias

	rows := database.NewQuery(`
		SELECT base10id,base36id
		FROM base10tobase36
		WHERE base10id=?`, alias).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var base10Id string
		var base36Id string
		err := rows.Scan(&base10Id, &base36Id)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		aliasToUse = base36Id
		return nil
	})
	if err != nil {
		return "", false, fmt.Errorf("Couldn't convert base10Id=>base36Id", err)
	}

	pageId, ok, err := LoadAliasToPageId(db, u, aliasToUse)
	if err != nil {
		return "", false, fmt.Errorf("Couldn't convert alias", err)
	}

	return pageId, ok, nil
}

// LoadAliasToPageId converts the given page alias to page id.
func LoadAliasToPageId(db *database.DB, u *CurrentUser, alias string) (string, bool, error) {
	aliasToIdMap, err := LoadAliasToPageIdMap(db, u, []string{alias})
	if err != nil {
		return "", false, err
	}
	pageId, ok := aliasToIdMap[strings.ToLower(alias)]
	return pageId, ok, nil
}

// LoadAliasAndPageId returns both the alias and the pageId for a given alias or pageId.
func LoadAliasAndPageId(db *database.DB, u *CurrentUser, alias string) (string, string, bool, error) {
	aliasToIdMap, err := LoadAliasToPageIdMap(db, u, []string{alias})
	if err != nil {
		return "", "", false, err
	}

	// return the matching alias->pageId entry, but not the matching pageId->pageId entry
	for nextAlias, nextPageId := range aliasToIdMap {
		if (nextAlias == strings.ToLower(alias) || nextPageId == strings.ToLower(alias)) && nextAlias != nextPageId {
			return nextAlias, nextPageId, true, err
		}
	}

	return "", "", false, err
}
