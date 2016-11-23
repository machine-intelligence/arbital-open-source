// page.go contains all the page stuff
package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"zanaduu3/src/database"
)

const (
	// Various page types we have in our system.
	WikiPageType     = "wiki"
	CommentPageType  = "comment"
	QuestionPageType = "question"
	GroupPageType    = "group"

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
	NewEditProposalChangeLog    = "newEditProposal"
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
	LensOrderChangedChangeLog   = "lensOrderChanged"
	PathOrderChangedChangeLog   = "pathOrderChanged"

	// Mark types
	QueryMarkType     = "query"
	TypoMarkType      = "typo"
	ConfusionMarkType = "confusion"

	// How long the page lock lasts
	PageQuickLockDuration = 5 * 60  // in seconds
	PageLockDuration      = 30 * 60 // in seconds
)

const (
	NoMasteryLevel        = iota
	LooseMasteryLevel     = iota
	BasicMasteryLevel     = iota
	TechnicalMasteryLevel = iota
	ResearchMasteryLevel  = iota

	// Total count of different mastery levels
	MasteryLevelCount = iota
)

var (
	// Regexp that strictly matches an alias, and not a page id
	StrictAliasRegexp = regexp.MustCompile("^[A-Za-z_][0-9A-Za-z_]*$")
)

// corePageData has data we load directly from the pages and pageInfos tables.
type corePageData struct {
	Likeable

	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageID                   string `json:"pageId"`
	Edit                     int    `json:"edit"`
	EditSummary              string `json:"editSummary"`
	PrevEdit                 int    `json:"prevEdit"`
	CurrentEdit              int    `json:"currentEdit"`
	WasPublished             bool   `json:"wasPublished"`
	Type                     string `json:"type"`
	Title                    string `json:"title"`
	Clickbait                string `json:"clickbait"`
	TextLength               int    `json:"textLength"` // number of characters
	Alias                    string `json:"alias"`
	SortChildrenBy           string `json:"sortChildrenBy"`
	HasVote                  bool   `json:"hasVote"`
	VoteType                 string `json:"voteType"`
	EditCreatorID            string `json:"editCreatorId"`
	EditCreatedAt            string `json:"editCreatedAt"`
	PageCreatorID            string `json:"pageCreatorId"`
	PageCreatedAt            string `json:"pageCreatedAt"`
	SeeDomainID              string `json:"seeDomainId"`
	EditDomainID             string `json:"editDomainId"`
	IsAutosave               bool   `json:"isAutosave"`
	IsSnapshot               bool   `json:"isSnapshot"`
	IsLiveEdit               bool   `json:"isLiveEdit"`
	IsMinorEdit              bool   `json:"isMinorEdit"`
	IsRequisite              bool   `json:"isRequisite"`
	IndirectTeacher          bool   `json:"indirectTeacher"`
	TodoCount                int    `json:"todoCount"`
	IsEditorComment          bool   `json:"isEditorComment"`
	IsEditorCommentIntention bool   `json:"isEditorCommentIntention"`
	IsResolved               bool   `json:"isResolved"`
	SnapshotText             string `json:"snapshotText"`
	AnchorContext            string `json:"anchorContext"`
	AnchorText               string `json:"anchorText"`
	AnchorOffset             int    `json:"anchorOffset"`
	MergedInto               string `json:"mergedInto"`
	IsDeleted                bool   `json:"isDeleted"`
	ViewCount                int    `json:"viewCount"`

	// The following data is filled on demand.
	Text     string `json:"text"`
	MetaText string `json:"metaText"`
}

// NewCorePageData returns a pointer to a new corePageData object created with the given page id
func NewCorePageData(pageID string) *corePageData {
	data := &corePageData{PageID: pageID}
	data.Likeable = *NewLikeable(PageLikeableType)
	return data
}

type Page struct {
	corePageData

	LoadOptions PageLoadOptions `json:"-"`

	// === Auxillary data. ===
	// For some pages we load additional data.
	IsSubscribed             bool `json:"isSubscribed"`
	IsSubscribedAsMaintainer bool `json:"isSubscribedAsMaintainer"`
	SubscriberCount          int  `json:"subscriberCount"`
	MaintainerCount          int  `json:"maintainerCount"`
	// Last time the user visited this page.
	LastVisit string `json:"lastVisit"`
	// True iff the user has a work-in-progress draft for this page
	HasDraft bool `json:"hasDraft"`

	// === Full data. ===
	// For pages that are displayed fully, we load more additional data.
	// Edit number for the currently live version
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
	NextPageID  string `json:"nextPageId"`
	PrevPageID  string `json:"prevPageId"`
	// Whether or not the page is used as a requirement
	UsedAsMastery bool `json:"usedAsMastery"`
	// If not 0, this is a pending edit which has been proposed
	ProposalEditNum int `json:"proposalEditNum"`

	// What actions the current user can perform with this page
	Permissions *Permissions `json:"permissions"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".
	Summaries map[string]string `json:"summaries"`
	// Ids of the users who edited this page. Ordered by how much they contributed.
	CreatorIDs []string `json:"creatorIds"`

	// Relevant ids.
	ChildIDs    []string `json:"childIds"`
	ParentIDs   []string `json:"parentIds"`
	CommentIDs  []string `json:"commentIds"`
	QuestionIDs []string `json:"questionIds"`
	TagIDs      []string `json:"tagIds"`
	RelatedIDs  []string `json:"relatedIds"`
	MarkIDs     []string `json:"markIds"`

	// Page pairs for Concept pages
	Explanations []*PagePair `json:"explanations"`
	LearnMore    []*PagePair `json:"learnMore"`

	// Requisite stuff
	Requirements []*PagePair `json:"requirements"`
	Subjects     []*PagePair `json:"subjects"`

	// Lens stuff
	Lenses       LensList `json:"lenses"`
	LensParentID string   `json:"lensParentId"`

	// Path stuff
	PathPages Path `json:"pathPages"`

	// Data for showing what pages user can learn after reading this page.
	// Map: page id of the subject that was taught/covered -> list of pageIds to read
	LearnMoreTaughtMap   map[string][]string `json:"learnMoreTaughtMap"`
	LearnMoreCoveredMap  map[string][]string `json:"learnMoreCoveredMap"`
	LearnMoreRequiredMap map[string][]string `json:"learnMoreRequiredMap"`

	// Edit history for when we need to know which edits are based on which edits (key is 'edit' number)
	EditHistory map[string]*EditInfo `json:"editHistory"`

	// All domains this page has been submitted to (map key: domainId)
	DomainSubmissions map[string]*PageToDomainSubmission `json:"domainSubmissions"`

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

	// === FE data ===
	// This data isn't loaded on the BE, but populated on the FE
	// Map: red alias -> pretty text
	RedAliases map[string]string `json:"redAliases"`
	// List of page's meta tags that indicate the page should be improved
	ImprovementTagIDs []string `json:"improvementTagIds"`
	// List of page's tags that are not meta tags
	NonMetaTagIDs []string `json:"nonMetaTagIds"`
	// TODOs extracted from the page's text
	Todos []string `json:"todos"`
	// PagePairs for "go slower/faster" suggestions; subjectId -> list of pagePairs
	SlowDownMap map[string][]*PagePair `json:"slowDownMap"`
	SpeedUpMap  map[string][]*PagePair `json:"speedUpMap"`
	ArcPageIDs  []string               `json:"arcPageIds"`

	// Map from request type to a ContentRequest object for that type.
	ContentRequests map[string]*ContentRequest `json:"contentRequests"`
}

// NewPage returns a pointer to a new page object created with the given page id
func NewPage(pageID string) *Page {
	p := &Page{corePageData: *NewCorePageData(pageID)}
	p.Votes = make([]*Vote, 0)
	p.Summaries = make(map[string]string)
	p.CreatorIDs = make([]string, 0)
	p.CommentIDs = make([]string, 0)
	p.QuestionIDs = make([]string, 0)
	p.TagIDs = make([]string, 0)
	p.RelatedIDs = make([]string, 0)
	p.Requirements = make([]*PagePair, 0)
	p.Subjects = make([]*PagePair, 0)
	p.ChangeLogs = make([]*ChangeLog, 0)
	p.ChildIDs = make([]string, 0)
	p.ParentIDs = make([]string, 0)
	p.MarkIDs = make([]string, 0)
	p.Explanations = make([]*PagePair, 0)
	p.LearnMore = make([]*PagePair, 0)
	p.Lenses = make(LensList, 0)
	p.PathPages = make(Path, 0)
	p.LearnMoreTaughtMap = make(map[string][]string)
	p.LearnMoreCoveredMap = make(map[string][]string)
	p.LearnMoreRequiredMap = make(map[string][]string)
	p.DomainSubmissions = make(map[string]*PageToDomainSubmission)
	p.Answers = make([]*Answer, 0)
	p.SearchStrings = make(map[string]string)
	p.EditHistory = make(map[string]*EditInfo)
	p.RedAliases = make(map[string]string)
	p.ImprovementTagIDs = make([]string, 0)
	p.NonMetaTagIDs = make([]string, 0)
	p.Todos = make([]string, 0)
	p.ContentRequests = make(map[string]*ContentRequest)

	// Some fields are explicitly nil until they are loaded, so we can differentiate
	// between "not loaded" and "loaded, but empty"
	p.SlowDownMap = nil
	p.SpeedUpMap = nil

	// NOTE: we want permissions to be explicitly null so that if someone refers to them
	// they get an error. The permissions are only set when they are also fully computed.
	p.Permissions = nil
	return p
}

// Lens connection
type Lens struct {
	ID           int64  `json:"id,string"`
	PageID       string `json:"pageId"`
	LensID       string `json:"lensId"`
	LensIndex    int    `json:"lensIndex"`
	LensName     string `json:"lensName"`
	LensSubtitle string `json:"lensSubtitle"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
	UpdatedBy    string `json:"updatedBy"`
	UpdatedAt    string `json:"updatedAt"`
}

type LensList []*Lens

func (a LensList) Len() int           { return len(a) }
func (a LensList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a LensList) Less(i, j int) bool { return a[i].LensIndex < a[j].LensIndex }

// PathPage connection
// Corresponds to one row from pathPages table
type PathPage struct {
	ID         int64  `json:"id,string"`
	GuideID    string `json:"guideId"`
	PathPageID string `json:"pathPageId"`
	PathIndex  int    `json:"pathIndex"`
	CreatedBy  string `json:"createdBy"`
	CreatedAt  string `json:"createdAt"`
	UpdatedBy  string `json:"updatedBy"`
	UpdatedAt  string `json:"updatedAt"`
}

type Path []*PathPage

func (a Path) Len() int           { return len(a) }
func (a Path) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Path) Less(i, j int) bool { return a[i].PathIndex < a[j].PathIndex }

// When a user starts a path, they get their own path instance
// Corresponds to one row from pathInstances table
type PathInstance struct {
	ID         int64               `json:"id,string"`
	GuideID    string              `json:"guideId"`
	Pages      []*PathInstancePage `json:"pages"`
	Progress   int                 `json:"progress"`
	CreatedAt  string              `json:"createdAt"`
	UpdatedAt  string              `json:"updatedAt"`
	IsFinished bool                `json:"isFinished"`

	// True if the path was created by current user
	IsByCurrentUser bool `json:"isByCurrentUser"`

	// FE data
	// Insert these page ids when continuing the path
	// question alias -> list of pages
	PagesToInsert map[string][]*PathInstancePage `json:"-"`
}

func NewPathInstance() *PathInstance {
	var p PathInstance
	p.Pages = make([]*PathInstancePage, 0)
	return &p
}

// A structure for each page on a path
type PathInstancePage struct {
	PageID   string `json:"pageId"`
	SourceID string `json:"sourceId"`
}

// User's probability vote
type Vote struct {
	Value     int    `json:"value"`
	UserID    string `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

// ChangeLog describes a row from changeLogs table.
type ChangeLog struct {
	Likeable

	ID               string `json:"id"`
	PageID           string `json:"pageId"`
	UserID           string `json:"userId"`
	Edit             int    `json:"edit"`
	Type             string `json:"type"`
	CreatedAt        string `json:"createdAt"`
	AuxPageID        string `json:"auxPageId"`
	OldSettingsValue string `json:"oldSettingsValue"`
	NewSettingsValue string `json:"newSettingsValue"`
}

func NewChangeLog() *ChangeLog {
	var cl ChangeLog
	cl.Likeable = *NewLikeable(ChangeLogLikeableType)
	return &cl
}

// Concise information about a particular edit
type EditInfo struct {
	Edit     int `json:"edit"`
	PrevEdit int `json:"prevEdit"`
}

// Mastery is a page you should have mastered before you can understand another page.
type Mastery struct {
	PageID    string `json:"pageId"`
	Has       bool   `json:"has"`
	Wants     bool   `json:"wants"`
	Level     int    `json:"level"`
	UpdatedAt string `json:"updatedAt"`
}

// Mark is something attached to a page, e.g. a place where a user said they were confused.
type Mark struct {
	ID                  string `json:"id"`
	PageID              string `json:"pageId"`
	Type                string `json:"type"`
	IsCurrentUserOwned  bool   `json:"isCurrentUserOwned"`
	CreatedAt           string `json:"createdAt"`
	AnchorContext       string `json:"anchorContext"`
	AnchorText          string `json:"anchorText"`
	AnchorOffset        int    `json:"anchorOffset"`
	Text                string `json:"text"`
	RequisiteSnapshotID string `json:"requisiteSnapshotId"`
	ResolvedPageID      string `json:"resolvedPageId"`
	IsSubmitted         bool   `json:"isSubmitted"`
	Answered            bool   `json:"answered"`

	// If the mark was resolved by the owner, we want to display that. But that also
	// means we can't send the ResolvedBy value to the FE, so we use IsResolveByOwner instead.
	IsResolvedByOwner bool   `json:"isResolvedByOwner"`
	ResolvedBy        string `json:"resolvedBy"`

	// Marks are anonymous, so the only info FE gets is whether this mark is owned
	// by the current user.
	CreatorID string `json:"-"`
}

// PageObject stores some information for an object embedded in a page
type PageObject struct {
	PageID string `json:"pageId"`
	Edit   int    `json:"edit"`
	Object string `json:"object"`
	Value  string `json:"value"`
}

// Information about a page which was submitted to a domain
type PageToDomainSubmission struct {
	PageID      string `json:"pageId"`
	DomainID    string `json:"domainId"`
	CreatedAt   string `json:"createdAt"`
	SubmitterID string `json:"submitterId"`
	ApprovedAt  string `json:"approvedAt"`
	ApproverID  string `json:"approverId"`
}

// Answer is attached to a question page, and points to another page that
// answers the question.
type Answer struct {
	ID           int64  `json:"id,string"`
	QuestionID   string `json:"questionId"`
	AnswerPageID string `json:"answerPageId"`
	UserID       string `json:"userId"`
	CreatedAt    string `json:"createdAt"`
}

// SearchString is attached to a question page to help with directing users
// towards it via search or marks.
type SearchString struct {
	ID     int64
	PageID string
	Text   string
}

// ContentRequest is unique to a page,requestType pair. Each time a user makes a content
// request (e.g. slowDown, moreWords) on a page, their vote is counted as a like on the
// corresponding ContentRequest's likeable.
type ContentRequest struct {
	Likeable

	ID          int64  `json:"id,string"`
	PageID      string `json:"pageId"`
	RequestType string `json:"requestType"`
	CreatedAt   string `json:"createdAt"`
}

func NewContentRequest() *ContentRequest {
	var cr ContentRequest
	cr.Likeable = *NewLikeable(ContentRequestLikeableType)
	return &cr
}

// GlobalHandlerData includes misc data that we send to the FE with each request
// that resets FE data.
type GlobalHandlerData struct {
	// Private domain the current user is in
	PrivateDomain *Domain `json:"privateDomain"`
	// List of all domains
	//DomainIDs []string `json:"domainIds"`
	// List of tags ids that mean a page should be improved
	ImprovementTagIDs []string `json:"improvementTagIds"`
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
	DomainMap  map[string]*Domain  `json:"domains"`
	MasteryMap map[string]*Mastery `json:"masteries"`
	MarkMap    map[string]*Mark    `json:"marks"`
	// Page id -> {object alias -> object}
	PageObjectMap map[string]map[string]*PageObject `json:"pageObjects"`
	// ResultMap contains various data the specific handler returns
	ResultMap map[string]interface{} `json:"result"`

	GlobalData *GlobalHandlerData `json:"globalData"`
}

// NewHandlerData creates and initializes a new commonHandlerData object.
func NewHandlerData(u *CurrentUser) *CommonHandlerData {
	var data CommonHandlerData
	data.User = u
	data.PageMap = make(map[string]*Page)
	data.EditMap = make(map[string]*Page)
	data.UserMap = make(map[string]*User)
	data.DomainMap = make(map[string]*Domain)
	data.MasteryMap = make(map[string]*Mastery)
	data.MarkMap = make(map[string]*Mark)
	data.PageObjectMap = make(map[string]map[string]*PageObject)
	data.ResultMap = make(map[string]interface{})

	if u.ID != "" {
		data.UserMap[u.ID] = &u.User
	}
	return &data
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

	// For fresh data, make sure that various things are definitely loaded
	/*if data.ResetEverything {
		_, err := LoadAllDomainIDs(db, pageMap)
		if err != nil {
			return fmt.Errorf("LoadAllDomainIds for failed: %v", err)
		}
	}*/

	// Load comments
	filteredPageMap := filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Comments })
	err := LoadCommentIDs(db, u, pageMap, &LoadDataOptions{ForPages: filteredPageMap})
	if err != nil {
		return fmt.Errorf("LoadCommentIds for failed: %v", err)
	}

	// Load questions
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Questions })
	err = LoadChildIDs(db, pageMap, u, &LoadChildIdsOptions{
		ForPages:     filteredPageMap,
		Type:         QuestionPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  IntrasitePopoverLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadChildIds for questions failed: %v", err)
	}

	// Load explanations
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Explanations })
	err = LoadExplanations(db, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadExplanationIds failed: %v", err)
	}

	// Load children
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Children })
	err = LoadChildIDs(db, pageMap, u, &LoadChildIdsOptions{
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
	err = LoadParentIDs(db, pageMap, u, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: ParentPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for parents failed: %v", err)
	}

	// Load tags
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Tags })
	err = LoadParentIDs(db, pageMap, u, &LoadParentIdsOptions{
		ForPages:     filteredPageMap,
		PagePairType: TagPagePairType,
		LoadOptions:  TitlePlusLoadOptions,
	})
	if err != nil {
		return fmt.Errorf("LoadParentIds for tags failed: %v", err)
	}

	// Load related
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Related })
	err = LoadChildIDs(db, pageMap, u, &LoadChildIdsOptions{
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
	err = LoadLensesForPages(db, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLensesForPages failed: %v", err)
	}

	// Load path pages
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Path })
	err = LoadPathForPages(db, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadPathForPages failed: %v", err)
	}

	// Load requisites
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Requisites })
	err = LoadRequisites(db, pageMap, u, &LoadReqsOptions{
		ForPages:   filteredPageMap,
		MasteryMap: masteryMap,
	})
	if err != nil {
		return fmt.Errorf("LoadRequisites failed: %v", err)
	}

	// Load domains the pages have been submitted to
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SubmittedTo })
	err = LoadPageToDomainSubmissionsForPages(db, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadPageToDomainSubmissionsForPages failed: %v", err)
	}

	// Load answers
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Answers })
	err = LoadAnswers(db, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadAnswers failed: %v", err)
	}

	if u.ID != "" {
		// Load user's marks
		filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.UserMarks })
		err = LoadMarkIDs(db, u, pageMap, markMap, &LoadMarkIDsOptions{
			ForPages:              filteredPageMap,
			CurrentUserConstraint: true,
		})
		if err != nil {
			return fmt.Errorf("LoadMarkIds for user's marks failed: %v", err)
		}

		// Load unresolved marks
		filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.UnresolvedMarks })
		err = LoadMarkIDs(db, u, pageMap, markMap, &LoadMarkIDsOptions{
			ForPages:         filteredPageMap,
			EditorConstraint: true,
		})
		if err != nil {
			return fmt.Errorf("LoadMarkIds for unresolved marks failed: %v", err)
		}

		// Load all marks if forced to
		filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.AllMarks })
		err = LoadMarkIDs(db, u, pageMap, markMap, &LoadMarkIDsOptions{
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

	// Load links
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Links })
	err = LoadLinks(db, u, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLinks failed: %v", err)
	}

	// Load domains
	/*filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.DomainsAndPermissions })
	err = LoadDomainsForPages(db, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadDomainsForPages failed: %v", err)
	}*/

	// Load learn more pages
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.LearnMore })
	err = LoadLearnMore(db, u, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLearnMore failed: %v", err)
	}

	// Load change logs
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ChangeLogs })
	err = LoadChangeLogsForPages(db, u, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadChangeLogsForPages failed: %v", err)
	}

	// Load content requests
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ContentRequests })
	err = LoadContentRequestsForPages(db, u, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadContentRequestsForPages failed: %v", err)
	}

	// Load search strings
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SearchStrings })
	err = LoadSearchStrings(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSearchStrings failed: %v", err)
	}

	// Load whether the pages are lenses for other pages
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.LensParentID })
	err = LoadLensParentIDs(db, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLensParentIds failed: %v", err)
	}

	// Load whether or not the pages have an unpublished draft
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.HasDraft })
	err = LoadDraftExistence(db, u.ID, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadDraftExistence failed: %v", err)
	}

	// Load votes
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Votes })
	err = LoadVotes(db, u.ID, filteredPageMap, userMap)
	if err != nil {
		return fmt.Errorf("LoadVotes failed: %v", err)
	}

	// Load last visit dates
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.LastVisit })
	err = LoadLastVisits(db, u.ID, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadLastVisits failed: %v", err)
	}

	// Load subscriptions
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.IsSubscribed })
	err = LoadSubscriptions(db, u.ID, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSubscriptions failed: %v", err)
	}

	// Load subscriber count
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SubscriberCount })
	err = LoadSubscriberCount(db, u.ID, filteredPageMap)
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
	err = LoadCreatorIDs(db, u, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadCreatorIds failed: %v", err)
	}

	// Load pages' edit history
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.EditHistory })
	err = LoadEditHistory(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadEditHistory failed: %v", err)
	}

	// Add other pages we'll need
	/*for _, u := range userMap {
		for _, dm := range u.DomainMembershipMap {
			AddPageIDToMap(dm.PageID, pageMap)
		}
	}*/

	// Load domain roles for user pages that need it
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.DomainRoles })
	for pageID, p := range filteredPageMap {
		if p.Type == GroupPageType {
			err = LoadUserDomainMembership(db, userMap[pageID], data.DomainMap)
			if err != nil {
				return fmt.Errorf("LoadUserDomainMembership failed: %v", err)
			}
		}
	}

	// Load all relevant domains
	err = LoadDomainsForPages(db, u, pageMap, userMap, data.DomainMap)
	if err != nil {
		return fmt.Errorf("LoadAllDomains failed: %v", err)
	}

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

	// Load proposal edit number (if any)
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.ProposalEditNum })
	err = LoadProposalEditNum(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadProposalEditNum failed: %v", err)
	}

	// Load (dis)likes
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Likes })
	individualLikesPageMap := filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.IndividualLikes })
	err = LoadLikesForPages(db, u, filteredPageMap, individualLikesPageMap, userMap)
	if err != nil {
		return fmt.Errorf("LoadLikesForPages failed: %v", err)
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
			masteryMap[p.PageID] = &Mastery{PageID: p.PageID}
		}
	}

	// Load what requirements the user has met
	err = LoadMasteries(db, u, masteryMap)
	if err != nil {
		return fmt.Errorf("LoadMasteries failed: %v", err)
	}

	// Load all the users
	AddUserIDToMap(u.ID, userMap)
	for _, p := range pageMap {
		AddUserIDToMap(p.PageCreatorID, userMap)
		AddUserIDToMap(p.EditCreatorID, userMap)
		if IsIDValid(p.LockedBy) {
			AddUserIDToMap(p.LockedBy, userMap)
		}
	}
	err = LoadUsers(db, userMap, u.ID)
	if err != nil {
		return fmt.Errorf("LoadUsers failed: %v", err)
	}

	// Computed which pages count as visited.
	// TODO: refactor this code to use multiple-maps insert wrapper
	visitedValues := make([]interface{}, 0)
	visitorID := u.GetSomeID()
	if visitorID != "" {
		for pageID, p := range pageMap {
			if p.Text != "" {
				visitedValues = append(visitedValues, visitorID, u.SessionID, u.AnalyticsID, db.C.R.RemoteAddr, pageID, database.Now())
			}
		}
	}

	// Add a visit to pages for which we loaded text.
	if len(visitedValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO visits (userId, sessionId, analyticsId, ipAddress, pageId, createdAt)
			VALUES ` + database.ArgsPlaceholder(len(visitedValues), 6))
		if _, err = statement.Exec(visitedValues...); err != nil {
			return fmt.Errorf("Couldn't update visits", err)
		}
	}

	return nil
}

// LoadMasteries loads the masteries.
func LoadMasteries(db *database.DB, u *CurrentUser, masteryMap map[string]*Mastery) error {
	userID := u.GetSomeID()
	if userID == "" {
		return nil
	}

	rows := database.NewQuery(`
		SELECT masteryId,updatedAt,has,wants,level
		FROM userMasteryPairs
		WHERE userId=?`, userID).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var mastery Mastery
		err := rows.Scan(&mastery.PageID, &mastery.UpdatedAt, &mastery.Has, &mastery.Wants, &mastery.Level)
		if err != nil {
			return fmt.Errorf("failed to scan for mastery: %v", err)
		}
		masteryMap[mastery.PageID] = &mastery
		return nil
	})
	return err
}

// LoadPageObjects loads all the page objects necessary for the given pages.
func LoadPageObjects(db *database.DB, u *CurrentUser, pageMap map[string]*Page, pageObjectMap map[string]map[string]*PageObject) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)

	userID := u.GetSomeID()
	if userID == "" {
		return nil
	}

	rows := database.NewQuery(`
		SELECT pageId,edit,object,value
		FROM userPageObjectPairs
		WHERE userId=?`, userID).Add(`AND pageId IN `).AddArgsGroup(pageIDs).Add(`
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var obj PageObject
		err := rows.Scan(&obj.PageID, &obj.Edit, &obj.Object, &obj.Value)
		if err != nil {
			return fmt.Errorf("Failed to scan for user: %v", err)
		}
		if _, ok := pageObjectMap[obj.PageID]; !ok {
			pageObjectMap[obj.PageID] = make(map[string]*PageObject)
		}
		pageObjectMap[obj.PageID][obj.Object] = &obj
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
	pageIDs := PageIDsListFromMap(pageMap)

	// Compute pages for which to load text / summary
	textIDs := make([]interface{}, 0)
	for _, p := range pageMap {
		if p.LoadOptions.Text {
			textIDs = append(textIDs, p.PageID)
		}
	}
	textSelect := database.NewQuery(`IF(p.pageId IN`).AddIdsGroup(textIDs).Add(`,p.text,"") AS text`)

	// Load the page data
	rows := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,p.creatorId,p.createdAt,p.title,p.clickbait,`).AddPart(textSelect).Add(`,
			length(p.text),p.metaText,pi.type,pi.hasVote,pi.voteType,
			pi.alias,pi.createdAt,pi.createdBy,pi.sortChildrenBy,pi.seeDomainId,pi.editDomainId,
			pi.isEditorComment,pi.isEditorCommentIntention,pi.isResolved,
			pi.isRequisite,pi.indirectTeacher,pi.currentEdit,pi.likeableId,pi.viewCount,
			p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,p.editSummary,pi.isDeleted,pi.mergedInto,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset
		FROM pages AS p
		JOIN`).AddPart(pageInfosTable).Add(`AS pi
		ON (p.pageId = pi.pageId AND p.edit = pi.currentEdit)
		WHERE p.pageId IN`).AddArgsGroup(pageIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		p := NewCorePageData("")
		err := rows.Scan(
			&p.PageID, &p.Edit, &p.PrevEdit, &p.EditCreatorID, &p.EditCreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.Type, &p.HasVote,
			&p.VoteType, &p.Alias, &p.PageCreatedAt, &p.PageCreatorID, &p.SortChildrenBy,
			&p.SeeDomainID, &p.EditDomainID, &p.IsEditorComment, &p.IsEditorCommentIntention,
			&p.IsResolved, &p.IsRequisite, &p.IndirectTeacher, &p.CurrentEdit, &p.LikeableID, &p.ViewCount,
			&p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit, &p.EditSummary, &p.IsDeleted, &p.MergedInto,
			&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("Failed to scan a page: %v", err)
		}
		pageMap[p.PageID].corePageData = *p
		pageMap[p.PageID].WasPublished = true
		return nil
	})
	for _, p := range pageMap {
		if p.Type == "" {
			delete(pageMap, p.PageID)
		}
	}
	return err
}

// LoadSummaries loads summaries for the given pages.
func LoadSummaries(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT pageId,name,text
		FROM pageSummaries
		WHERE pageId IN`).AddArgsGroup(pageIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var name, text string
		err := rows.Scan(&pageID, &name, &text)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[pageID].Summaries[name] = text
		return nil
	})
	return err
}

// LoadEditHistory loads edit histories for given pages
func LoadEditHistory(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT pageId,edit,prevEdit
		FROM pages
		WHERE pageId IN`).AddArgsGroup(pageIDs).Add(`
			AND NOT isSnapshot AND NOT isAutosave`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var editInfo EditInfo
		err := rows.Scan(&pageID, &editInfo.Edit, &editInfo.PrevEdit)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[pageID].EditHistory[fmt.Sprintf("%d", editInfo.Edit)] = &editInfo
		return nil
	})
	return err
}

// LoadLinkedMarkCounts loads the number of marks that link to these questions.
func LoadLinkedMarkCounts(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT resolvedPageId,SUM(1)
		FROM marks
		WHERE resolvedPageId IN`).AddArgsGroup(pageIDs).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var resolvedPageID string
		var count int
		err := rows.Scan(&resolvedPageID, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[resolvedPageID].LinkedMarkCount = count
		return nil
	})
	return err
}

type ProcessChangeLogCallback func(db *database.DB, changeLog *ChangeLog) error

// LoadChangeLogs loads the change logs matching the given condition.
func LoadChangeLogs(db *database.DB, queryPart *database.QueryPart, resultData *CommonHandlerData, callback ProcessChangeLogCallback) error {
	rows := database.NewQuery(`
		SELECT cl.id,cl.pageId,cl.userId,cl.edit,cl.type,cl.createdAt,cl.auxPageId,cl.likeableId,
			cl.oldSettingsValue,cl.newSettingsValue
		FROM changeLogs as cl`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		l := NewChangeLog()
		err := rows.Scan(&l.ID, &l.PageID, &l.UserID, &l.Edit, &l.Type, &l.CreatedAt,
			&l.AuxPageID, &l.LikeableID, &l.OldSettingsValue, &l.NewSettingsValue)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		if resultData != nil {
			AddUserToMap(l.UserID, resultData.UserMap)
			AddPageToMap(l.AuxPageID, resultData.PageMap, TitlePlusIncludeDeletedLoadOptions)
		}
		return callback(db, l)
	})
	if err != nil {
		return fmt.Errorf("Couldn't load changeLogs: %v", err)
	}
	return nil
}

// LoadChangeLogsForPages loads the edit history for the given pages.
func LoadChangeLogsForPages(db *database.DB, u *CurrentUser, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	for _, p := range sourcePageMap {
		p.ChangeLogs = make([]*ChangeLog, 0)
		queryPart := database.NewQuery(`
			WHERE pageId=?`, p.PageID).Add(`
				AND (userId=? OR type!=?)`, u.ID, NewSnapshotChangeLog).Add(`
			ORDER BY createdAt DESC`)
		err := LoadChangeLogs(db, queryPart, resultData, func(db *database.DB, changeLog *ChangeLog) error {
			p.ChangeLogs = append(p.ChangeLogs, changeLog)
			return nil
		})
		if err != nil {
			return fmt.Errorf("Couldn't load changeLogs: %v", err)
		}

		err = LoadLikesForChangeLogs(db, u, p.ChangeLogs)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadChangeLogsByIds loads the changelogs with given ids
func LoadChangeLogsByIDs(db *database.DB, ids []string, typeConstraint string) (map[string]*ChangeLog, error) {
	changeLogs := make(map[string]*ChangeLog)
	queryPart := database.NewQuery(`
			WHERE id IN`).AddArgsGroupStr(ids).Add(`
				AND type=?`, typeConstraint)
	err := LoadChangeLogs(db, queryPart, nil, func(db *database.DB, changeLog *ChangeLog) error {
		changeLogs[changeLog.ID] = changeLog
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load changeLogs: %v", err)
	}
	return changeLogs, nil
}

// Load LikeCount and MyLikeValue for a set of ChangeLogs
func LoadLikesForChangeLogs(db *database.DB, u *CurrentUser, changeLogs []*ChangeLog) error {
	likeablesMap := make(map[int64]*Likeable)
	for _, changeLog := range changeLogs {
		if changeLog.LikeableID != 0 {
			likeablesMap[changeLog.LikeableID] = &changeLog.Likeable
		}
	}
	return LoadLikes(db, u, likeablesMap, nil, nil)
}

// LoadContentRequestsForPages loads content requests for the given pages.
func LoadContentRequestsForPages(db *database.DB, u *CurrentUser, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	pageIDs := PageIDsListFromMap(sourcePageMap)
	if len(pageIDs) <= 0 {
		return nil
	}

	rows := database.NewQuery(`
		SELECT id, pageId, type, likeableId, createdAt
		FROM contentRequests
		WHERE pageId IN`).AddArgsGroup(pageIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		cr := NewContentRequest()
		err := rows.Scan(&cr.ID, &cr.PageID, &cr.RequestType, &cr.LikeableID, &cr.CreatedAt)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		p := resultData.PageMap[cr.PageID]
		p.ContentRequests[cr.RequestType] = cr
		return nil
	})
	if err != nil {
		return err
	}

	likeablesMap := make(map[int64]*Likeable)
	for id := range sourcePageMap {
		p := resultData.PageMap[id]
		for _, cr := range p.ContentRequests {
			l := &cr.Likeable
			if l.LikeableID != 0 {
				likeablesMap[l.LikeableID] = l
			}
		}
	}
	return LoadLikes(db, u, likeablesMap, nil, nil)
}

// LoadProposalEditNum loads the proposal edit number
func LoadProposalEditNum(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT p.pageId,p.prevEdit,cl.edit
		FROM changeLogs AS cl
		JOIN pages AS p
		ON (p.pageId=cl.pageId AND p.edit=cl.edit)
		WHERE cl.pageId IN`).AddArgsGroup(pageIDs).Add(`
			AND cl.type=?`, NewEditProposalChangeLog).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var prevEdit, edit int
		err := rows.Scan(&pageID, &prevEdit, &edit)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		p := pageMap[pageID]
		if p.CurrentEdit == prevEdit {
			p.ProposalEditNum = edit
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load proposalEditNum: %v", err)
	}
	return nil
}

type LoadEditOptions struct {
	// If true, the last edit will be loaded for the given user, even if it's an
	// autosave or a snapshot.
	LoadNonliveEdit bool
	// If true, when we load a non-live edit, we'll try to load the live edit first
	PreferLiveEdit bool

	// If set, we'll load this edit of the page
	LoadSpecificEdit int
	// If set, we'll only load from edits less than this
	LoadEditWithLimit int
	// If set, we'll only load from edits with createdAt timestamp before this
	CreatedAtLimit string
}

// LoadFullEdit loads and returns an edit for the given page id from the DB. It
// also computes the permissions for the edit.
// If the page couldn't be found, (nil, nil) will be returned.
func LoadFullEdit(db *database.DB, pageID string, u *CurrentUser, options *LoadEditOptions) (*Page, error) {
	if options == nil {
		options = &LoadEditOptions{}
	}
	p := NewPage(pageID)

	whereClause := database.NewQuery("p.isLiveEdit")
	if options.LoadSpecificEdit > 0 {
		whereClause = database.NewQuery("p.edit=?", options.LoadSpecificEdit)
	} else if options.CreatedAtLimit != "" {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND createdAt<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageID, options.CreatedAtLimit)
	} else if options.LoadEditWithLimit > 0 {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND edit<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageID, options.LoadEditWithLimit)
	} else if options.LoadNonliveEdit {
		orderClause := database.NewQuery(`
			/* From most to least preferred edits: autosave, (applicable) snapshot, currentEdit, anything else */
			ORDER BY p.isAutosave DESC,
				p.isSnapshot DESC,
				p.edit=pi.currentEdit DESC,
				p.createdAt DESC
		`)
		if options.PreferLiveEdit {
			orderClause = database.NewQuery(`
			/* From most to least preferred edits: autosave, (applicable) snapshot, currentEdit, anything else */
			ORDER BY p.edit=pi.currentEdit DESC,
				p.isAutosave DESC,
				p.isSnapshot DESC,
				p.createdAt DESC
		`)
		}
		// Load the most recent edit we have for the current user.
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT p.edit
				FROM pages AS p
				JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
				ON (p.pageId=pi.pageId)
				WHERE p.pageId=?`, pageID).Add(`
					AND (p.creatorId=? OR (NOT p.isSnapshot AND NOT p.isAutosave))`, u.ID).Add(`
					/* To consider a snapshot, it has to be based on the current edit */
					AND (NOT p.isSnapshot OR pi.currentEdit=0 OR p.prevEdit=pi.currentEdit)`).AddPart(orderClause).Add(`
				LIMIT 1
			)`)
	}
	statement := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,pi.type,p.title,p.clickbait,p.text,p.metaText,
			pi.alias,p.creatorId,pi.sortChildrenBy,pi.hasVote,pi.voteType,
			p.createdAt,pi.seeDomainId,pi.editDomainId,pi.createdAt,
			pi.createdBy,pi.isEditorComment,pi.isEditorCommentIntention,
			pi.isResolved,pi.likeableId,p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,p.editSummary,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset,
			pi.currentEdit>0,pi.isDeleted,pi.mergedInto,pi.currentEdit,pi.maxEdit,pi.lockedBy,pi.lockedUntil,
			pi.viewCount,pi.voteType,pi.isRequisite,pi.indirectTeacher
		FROM pages AS p
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (p.pageId=pi.pageId AND p.pageId=?)`, pageID).Add(`
		WHERE`).AddPart(whereClause).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageID, &p.Edit, &p.PrevEdit, &p.Type, &p.Title, &p.Clickbait,
		&p.Text, &p.MetaText, &p.Alias, &p.EditCreatorID, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.EditCreatedAt, &p.SeeDomainID,
		&p.EditDomainID, &p.PageCreatedAt, &p.PageCreatorID,
		&p.IsEditorComment, &p.IsEditorCommentIntention, &p.IsResolved, &p.LikeableID,
		&p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit, &p.EditSummary,
		&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset, &p.WasPublished,
		&p.IsDeleted, &p.MergedInto, &p.CurrentEdit, &p.MaxEditEver, &p.LockedBy, &p.LockedUntil,
		&p.ViewCount, &p.LockedVoteType, &p.IsRequisite, &p.IndirectTeacher)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}

	/*if exists {
		err = LoadDomainIDsForPage(db, p)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load domain ids for page: %v", err)
		}
	}*/
	p.ComputePermissions(db.C, u)

	p.TextLength = len(p.Text)
	return p, nil
}

// LoadPageIDs from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func LoadPageIDs(rows *database.Rows, pageMap map[string]*Page, loadOptions *PageLoadOptions) ([]string, error) {
	ids := []string{}
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		err := rows.Scan(&pageID)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}

		AddPageToMap(pageID, pageMap, loadOptions)
		ids = append(ids, pageID)
		return nil
	})
	return ids, err
}

// LoadSearchStrings loads all the search strings for the given pages
func LoadSearchStrings(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT id,pageId,text
		FROM searchStrings
		WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIDs))).Query(pageIDs...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var id int64
		var pageID, text string
		err := rows.Scan(&id, &pageID, &text)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		pageMap[pageID].SearchStrings[fmt.Sprintf("%d", id)] = text
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
		err := rows.Scan(&searchString.ID, &searchString.PageID, &searchString.Text)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		return nil
	})
	return &searchString, err
}

// LoadVotes loads probability votes corresponding to the given pages and updates the pages.
func LoadVotes(db *database.DB, currentUserID string, pageMap map[string]*Page, userMap map[string]*User) error {
	if len(pageMap) <= 0 {
		return nil
	}

	pageIDs := PageIDsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT userId,pageId,value,createdAt
		FROM (
			SELECT *
			FROM votes
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIDs)) + `
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`).Query(pageIDs...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var v Vote
		var pageID string
		err := rows.Scan(&v.UserID, &pageID, &v.Value, &v.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a vote: %v", err)
		}
		if v.Value == 0 {
			return nil
		}
		page := pageMap[pageID]
		if page.Votes == nil {
			page.Votes = make([]*Vote, 0, 0)
		}
		page.Votes = append(page.Votes, &v)
		if _, ok := userMap[v.UserID]; !ok {
			AddUserIDToMap(v.UserID, userMap)
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
	pageIdsList := PageIDsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT l.parentId,SUM(ISNULL(pi.pageId))
		FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
		RIGHT JOIN links AS l
		ON (pi.pageId=l.childAlias OR pi.alias=l.childAlias)
		WHERE l.parentId IN`).AddArgsGroup(pageIdsList).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID string
		var count int
		err := rows.Scan(&parentID, &count)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		pageMap[parentID].RedLinkCount = count
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
	pageIdsList := PageIDsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT parentId,count(*)
		FROM pagePairs
		WHERE type=?`, RequirementPagePairType).Add(`
			AND parentId IN`).AddArgsGroup(pageIdsList).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID string
		var count int
		err := rows.Scan(&parentID, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[parentID].UsedAsMastery = count > 0
		return nil
	})
	return err
}

// LoadCreatorIds loads creator ids for the pages
func LoadCreatorIDs(db *database.DB, u *CurrentUser, pageMap map[string]*Page, userMap map[string]*User, options *LoadDataOptions) error {
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
	pageIdsList := PageIDsListFromMap(sourceMap)

	rows := database.NewQuery(`
		SELECT pageId,creatorId,COUNT(*)
		FROM pages
		WHERE pageId IN`).AddArgsGroup(pageIdsList).Add(`
			AND NOT isAutosave AND NOT isSnapshot AND NOT isMinorEdit
		GROUP BY 1,2
		ORDER BY 3 DESC`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID, creatorID string
		var count int
		err := rows.Scan(&pageID, &creatorID, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[pageID].CreatorIDs = append(pageMap[pageID].CreatorIDs, creatorID)
		AddUserIDToMap(creatorID, userMap)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load page contributors: %v", err)
	}
	return nil
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

	pageIDs := PageIDsListFromMap(sourceMap)
	if len(pageIDs) <= 0 {
		return nil
	}

	// List of all aliases we'll need to convert to pageIds
	aliasesList := make([]interface{}, 0)

	// Load all links.
	rows := db.NewStatement(`
		SELECT parentId,childAlias
		FROM links
		WHERE parentId IN ` + database.InArgsPlaceholder(len(pageIDs))).Query(pageIDs...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID string
		var childAlias string
		err := rows.Scan(&parentID, &childAlias)
		if err != nil {
			return fmt.Errorf("failed to scan for a link: %v", err)
		}
		if IsIDValid(childAlias) {
			AddPageIDToMap(childAlias, pageMap)
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
			var pageID string
			err := rows.Scan(&pageID)
			if err != nil {
				return fmt.Errorf("failed to scan for a page: %v", err)
			}
			AddPageIDToMap(pageID, pageMap)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadLearnMore loads the "learn more" data for the given pages
func LoadLearnMore(db *database.DB, u *CurrentUser, pageMap map[string]*Page, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}

	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}

	// Compute all subject ids we need to consider
	var subjectIDs []string
	for _, page := range sourceMap {
		for _, subject := range page.Subjects {
			subjectIDs = append(subjectIDs, subject.ParentID)
		}
	}
	if len(subjectIDs) <= 0 {
		return nil
	}

	queryPart := database.NewQuery(`
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON (pi.pageId=pp.childId)
		WHERE pp.parentId IN`).AddArgsGroupStr(subjectIDs).Add(`
			AND (pp.type=? || pp.type=?)`, RequirementPagePairType, SubjectPagePairType)
	err := LoadPagePairs(db, queryPart, func(db *database.DB, pp *PagePair) error {
		for _, page := range sourceMap {
			for _, subject := range page.Subjects {
				if pp.ParentID != subject.ParentID || pp.Level != subject.Level || pp.ChildID == page.PageID {
					continue
				}
				if pp.Type == SubjectPagePairType && !pp.IsStrong {
					if subject.IsStrong {
						page.LearnMoreTaughtMap[subject.ParentID] = append(page.LearnMoreTaughtMap[subject.ParentID], pp.ChildID)
					} else {
						page.LearnMoreCoveredMap[subject.ParentID] = append(page.LearnMoreCoveredMap[subject.ParentID], pp.ChildID)
					}
				} else if pp.Type == RequirementPagePairType && pp.IsStrong && subject.IsStrong {
					page.LearnMoreRequiredMap[subject.ParentID] = append(page.LearnMoreRequiredMap[subject.ParentID], pp.ChildID)
				}
			}
		}
		AddPageIDToMap(pp.ChildID, pageMap)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type ProcessLoadPageToDomainSubmissionCallback func(db *database.DB, submission *PageToDomainSubmission) error

// LoadPageToDomainSubmissions loads information the pages that have been submitted to a domain
func LoadPageToDomainSubmissions(db *database.DB, queryPart *database.QueryPart, callback ProcessLoadPageToDomainSubmissionCallback) error {
	rows := database.NewQuery(`
		SELECT pds.pageId,pds.domainId,pds.createdAt,pds.submitterId,pds.approvedAt,pds.approverId
		FROM pageToDomainSubmissions AS pds`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var submission PageToDomainSubmission
		err := rows.Scan(&submission.PageID, &submission.DomainID, &submission.CreatedAt,
			&submission.SubmitterID, &submission.ApprovedAt, &submission.ApproverID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		return callback(db, &submission)
	})
	return err
}

// LoadPageToDomainSubmissionsForPages loads information about domains the pages have been submitted to
func LoadPageToDomainSubmissionsForPages(db *database.DB, pageMap map[string]*Page, userMap map[string]*User, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}

	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}

	pageIDs := PageIDsListFromMap(sourceMap)
	if len(pageIDs) <= 0 {
		return nil
	}

	queryPart := database.NewQuery(`WHERE pds.pageId IN`).AddArgsGroup(pageIDs)
	err := LoadPageToDomainSubmissions(db, queryPart, func(db *database.DB, submission *PageToDomainSubmission) error {
		AddPageIDToMap(submission.DomainID, pageMap)
		AddUserToMap(submission.SubmitterID, userMap)
		AddUserToMap(submission.ApproverID, userMap)
		pageMap[submission.PageID].DomainSubmissions[submission.DomainID] = submission
		return nil
	})
	return err
}

// LoadPageToDomainSubmission loads information about a specific page that was submitted to a specific domain
func LoadPageToDomainSubmission(db *database.DB, pageID, domainID string) (*PageToDomainSubmission, error) {
	queryPart := database.NewQuery(`
		WHERE pds.pageId=?`, pageID).Add(`
			AND pds.domainId=?`, domainID).Add(`
		LIMIT 1`)
	var resultSubmission *PageToDomainSubmission
	err := LoadPageToDomainSubmissions(db, queryPart, func(db *database.DB, submission *PageToDomainSubmission) error {
		resultSubmission = submission
		return nil
	})
	return resultSubmission, err
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

	pageIDs := PageIDsListFromMap(sourceMap)
	if len(pageIDs) <= 0 {
		return nil
	}

	rows := db.NewStatement(`
	SELECT id,questionId,answerPageId,userId,createdAt
	FROM answers
	WHERE questionId IN ` + database.InArgsPlaceholder(len(pageIDs))).Query(pageIDs...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var answer Answer
		err := rows.Scan(&answer.ID, &answer.QuestionID, &answer.AnswerPageID, &answer.UserID, &answer.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		AddPageToMap(answer.AnswerPageID, pageMap, AnswerLoadOptions)
		AddUserIDToMap(answer.UserID, userMap)
		pageMap[answer.QuestionID].Answers = append(pageMap[answer.QuestionID].Answers, &answer)
		return nil
	})
	return err
}

// LoadAnswer just loads the data for one specific answer.
func LoadAnswer(db *database.DB, answerID string) (*Answer, error) {
	var answer Answer
	_, err := db.NewStatement(`
		SELECT id,questionId,answerPageId,userId,createdAt
		FROM answers
		WHERE id=?`).QueryRow(answerID).Scan(&answer.ID, &answer.QuestionID,
		&answer.AnswerPageID, &answer.UserID, &answer.CreatedAt)
	return &answer, err
}

type LoadMarkIDsOptions struct {
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
func LoadMarkIDs(db *database.DB, u *CurrentUser, pageMap map[string]*Page, markMap map[string]*Mark, options *LoadMarkIDsOptions) error {
	sourceMap := options.ForPages
	if sourceMap == nil {
		sourceMap = pageMap
	}

	pageIDs := PageIDsListFromMap(sourceMap)
	if len(pageIDs) <= 0 {
		return nil
	}

	// Only load the marks the current user created
	constraint := database.NewQuery(``)
	if options.CurrentUserConstraint {
		constraint = database.NewQuery(`AND m.creatorId=?`, u.ID)
	}

	// Only load for pages in which current user is an author
	pageIdsPart := database.NewQuery(``).AddArgsGroup(pageIDs)
	if options.EditorConstraint {
		pageIdsPart = database.NewQuery(`(
			SELECT s.toId
			FROM subscriptions AS s
			WHERE s.toId IN`).AddArgsGroup(pageIDs).Add(`
				AND s.userId=?`, u.ID).Add(`
				AND s.asMaintainer
		)`)
		constraint.Add(`AND m.isSubmitted`)
	}

	// Whether or not to load marks that have been resolved
	if !options.LoadResolvedToo {
		constraint.Add(`AND m.resolvedBy=""`)
	}

	rows := database.NewQuery(`
		SELECT m.id
		FROM marks AS m
		WHERE m.pageId IN`).AddPart(pageIdsPart).Add(`
			`).AddPart(constraint).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var markID string
		err := rows.Scan(&markID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		if _, ok := markMap[markID]; !ok {
			markMap[markID] = &Mark{ID: markID}
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
	markIDs := make([]interface{}, 0)
	for markID := range markMap {
		markIDs = append(markIDs, markID)
	}

	rows := db.NewStatement(`
		SELECT id,type,pageId,creatorId,createdAt,anchorContext,anchorText,anchorOffset,
			text,requisiteSnapshotId,resolvedPageId,resolvedBy,answered,isSubmitted
		FROM marks
		WHERE id IN` + database.InArgsPlaceholder(len(markIDs))).Query(markIDs...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var creatorID string
		mark := &Mark{}
		err := rows.Scan(&mark.ID, &mark.Type, &mark.PageID, &mark.CreatorID, &mark.CreatedAt,
			&mark.AnchorContext, &mark.AnchorText, &mark.AnchorOffset, &mark.Text,
			&mark.RequisiteSnapshotID, &mark.ResolvedPageID, &mark.ResolvedBy, &mark.Answered,
			&mark.IsSubmitted)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		mark.IsCurrentUserOwned = mark.CreatorID == u.ID
		if mark.CreatorID == mark.ResolvedBy {
			mark.ResolvedBy = "0"
			mark.IsResolvedByOwner = true
		}
		*markMap[mark.ID] = *mark

		if page, ok := pageMap[mark.PageID]; ok {
			page.MarkIDs = append(page.MarkIDs, mark.ID)
		}
		AddPageToMap(mark.PageID, pageMap, TitlePlusLoadOptions)
		if mark.ResolvedPageID != "" {
			AddPageToMap(mark.ResolvedPageID, pageMap, TitlePlusLoadOptions)
			AddUserIDToMap(mark.ResolvedBy, userMap)
		}
		AddUserIDToMap(creatorID, userMap)
		return nil
	})
	return err
}

// Given a parent page that collects meta tags, load all its children.
func LoadMetaTags(db *database.DB, parentID string) ([]string, error) {
	pageMap := make(map[string]*Page)
	page := AddPageIDToMap(parentID, pageMap)
	options := &LoadChildIdsOptions{
		ForPages:     pageMap,
		Type:         WikiPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  EmptyLoadOptions,
	}
	err := LoadChildIDs(db, pageMap, nil, options)
	if err != nil {
		return nil, err
	}
	return page.ChildIDs, nil
}

// LoadSubpageCounts loads the number of various types of children the pages have
func LoadSubpageCounts(db *database.DB, u *CurrentUser, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pi.type,sum(1)
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND parentId IN`).AddArgsGroup(pageIDs).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON (pi.pageId=pp.childId)
		GROUP BY 1,2`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var childType string
		var count int
		err := rows.Scan(&pageID, &childType, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		if childType == CommentPageType {
			pageMap[pageID].CommentCount = count
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
	pageIDs := PageIDsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT questionId,sum(1)
		FROM answers
		WHERE questionId IN`).AddArgsGroup(pageIDs).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var questionID string
		var count int
		err := rows.Scan(&questionID, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[questionID].AnswerCount = count

		return nil
	})
	return err
}

// LoadCommentIds loads ids of all the comments for the pages in the given pageMap.
func LoadCommentIDs(db *database.DB, u *CurrentUser, pageMap map[string]*Page, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIDs := PageIDsListFromMap(sourcePageMap)
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
				AND pp.parentId IN`).AddArgsGroup(pageIDs).Add(`
		)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID, childID string
		err := rows.Scan(&parentID, &childID)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		parentPage := AddPageToMap(parentID, pageMap, SubpageLoadOptions)
		childPage := AddPageToMap(childID, pageMap, SubpageLoadOptions)
		parentPage.CommentIDs = append(parentPage.CommentIDs, childPage.PageID)
		return nil
	})

	// Now we have pages that have in comment ids both top level comments and
	// replies, so we need to remove the replies.
	for _, p := range sourcePageMap {
		replies := make(map[string]bool)
		for _, c := range p.CommentIDs {
			for _, r := range pageMap[c].CommentIDs {
				replies[r] = true
			}
		}
		onlyTopCommentIDs := make([]string, 0)
		for _, c := range p.CommentIDs {
			if !replies[c] {
				onlyTopCommentIDs = append(onlyTopCommentIDs, c)
			}
		}
		p.CommentIDs = onlyTopCommentIDs
	}
	return err
}

// LoadDraftExistence computes for each page whether or not the user has an
// autosave draft for it.
// This only makes sense to call for pages which were loaded for isLiveEdit=true.
func LoadDraftExistence(db *database.DB, userID string, options *LoadDataOptions) error {
	pageMap := options.ForPages
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId
		FROM pages
		WHERE pageId IN`).AddArgsGroup(pageIDs).Add(`
			AND isAutosave AND creatorId=?`, userID).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		err := rows.Scan(&pageID)
		if err != nil {
			return fmt.Errorf("Failed to scan a draft existence: %v", err)
		}
		pageMap[pageID].HasDraft = true
		return nil
	})
	return err
}

// LoadLastVisits loads lastVisit variable for each page.
func LoadLastVisits(db *database.DB, currentUserID string, pageMap map[string]*Page) error {
	// NOTE: Loading last visits is expensive; let's try to avoid it
	return nil
	/*if len(pageMap) <= 0 {
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
	return err*/
}

// LoadSubscriptions loads subscription statuses corresponding to the given
// pages, and then updates the given maps.
func LoadSubscriptions(db *database.DB, currentUserID string, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT toId,asMaintainer
		FROM subscriptions
		WHERE userId=?`, currentUserID).Add(`AND toId IN`).AddArgsGroup(pageIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageID string
		var asMaintainer bool
		err := rows.Scan(&toPageID, &asMaintainer)
		if err != nil {
			return fmt.Errorf("Failed to scan for a subscription: %v", err)
		}
		pageMap[toPageID].IsSubscribed = true
		pageMap[toPageID].IsSubscribedAsMaintainer = asMaintainer
		return nil
	})
	return err
}

// LoadSubscriberCount loads number of subscribers the page has.
func LoadSubscriberCount(db *database.DB, currentUserID string, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT toId,COUNT(*),SUM(asMaintainer)
		FROM subscriptions
		WHERE userId!=?`, currentUserID).Add(`
			AND toId IN`).AddArgsGroup(pageIDs).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageID string
		var subscriberCount, maintainerCount int
		err := rows.Scan(&toPageID, &subscriberCount, &maintainerCount)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		pageMap[toPageID].SubscriberCount = subscriberCount
		pageMap[toPageID].MaintainerCount = maintainerCount
		return nil
	})
	return err
}

// LoadDomainsForPages loads the domain info for the given page and adds them to the map
/*func LoadDomainsForPages(db *database.DB, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(sourcePageMap)
	queryPart := database.NewQuery(`
		JOIN pageInfos AS pi
		ON (pi.seeDomainId = d.domainId OR pi.editDomainId = d.domainId)
		WHERE pi.pageId IN`).AddArgsGroup(pageIDs)
	err := LoadDomains(db, queryPart, func(db *database.DB, domain *Domain) error {
		resultData.DomainMap[domain.ID] = domain
		AddPageToMap(domain.PageID, resultData.PageMap, TitlePlusLoadOptions)
		return nil
	})
	return err
}*/

/*func LoadDomainIDsForPage(db *database.DB, page *Page) error {
	pageMap := map[string]*Page{page.PageID: page}
	return LoadDomains(db, pageMap, &LoadDataOptions{
		ForPages: pageMap,
	})
}*/

// LoadAliasToPageIdMap loads the mapping from aliases to page ids.
func LoadAliasToPageIDMap(db *database.DB, u *CurrentUser, aliases []string) (map[string]string, error) {
	aliasToIDMap := make(map[string]string)
	if len(aliases) <= 0 {
		return aliasToIDMap, nil
	}

	strictAliases := make([]string, 0)
	strictPageIDs := make([]string, 0)
	for _, alias := range aliases {
		if IsIDValid(alias) {
			strictPageIDs = append(strictPageIDs, strings.ToLower(alias))
		} else {
			strictAliases = append(strictAliases, strings.ToLower(alias))
		}
	}

	var query *database.Stmt
	// TODO: refactor these queries into one query + additional parts
	if len(strictPageIDs) <= 0 {
		query = database.NewQuery(`
				SELECT pageId,alias
				FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
				WHERE alias IN`).AddArgsGroupStr(strictAliases).ToStatement(db)
	} else if len(strictAliases) <= 0 {
		query = database.NewQuery(`
				SELECT pageId,alias
				FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
				WHERE pageId IN`).AddArgsGroupStr(strictPageIDs).ToStatement(db)
	} else {
		query = database.NewQuery(`
				SELECT pageId,alias
				FROM`).AddPart(PageInfosTable(u)).Add(`AS pi
				WHERE pageId IN`).AddArgsGroupStr(strictPageIDs).Add(`
					OR alias IN`).AddArgsGroupStr(strictAliases).ToStatement(db)
	}

	rows := query.Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var alias string
		err := rows.Scan(&pageID, &alias)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		aliasToIDMap[strings.ToLower(alias)] = pageID
		aliasToIDMap[strings.ToLower(pageID)] = pageID
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't convert pageId=>alias", err)
	}

	// The query only gets results for when the page is published
	// We also want to return the pageIds even if they aren't for valid pages
	for _, pageID := range strictPageIDs {
		aliasToIDMap[strings.ToLower(pageID)] = strings.ToLower(pageID)
	}
	return aliasToIDMap, nil
}

// LoadAliasToPageId converts the given page alias to page id.
func LoadAliasToPageID(db *database.DB, u *CurrentUser, alias string) (string, bool, error) {
	aliasToIDMap, err := LoadAliasToPageIDMap(db, u, []string{alias})
	if err != nil {
		return "", false, err
	}
	pageID, ok := aliasToIDMap[strings.ToLower(alias)]
	return pageID, ok, nil
}

// Load all explanations for the given pages
func LoadExplanations(db *database.DB, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	explanationLoadOptions := (&PageLoadOptions{
		Tags: true,
	}).Add(IntrasitePopoverLoadOptions)

	pageIDs := PageIDsListFromMap(sourcePageMap)
	queryPart := database.NewQuery(`
		JOIN`).AddPart(PageInfosTable(resultData.User)).Add(`AS pi
		ON (pp.childId=pi.pageId)`).Add(`
		WHERE pp.parentId IN`).AddArgsGroup(pageIDs).Add(`
			AND pp.type=?`, SubjectPagePairType)
	err := LoadPagePairs(db, queryPart, func(db *database.DB, pp *PagePair) error {
		if pp.IsStrong {
			sourcePageMap[pp.ParentID].Explanations = append(sourcePageMap[pp.ParentID].Explanations, pp)
		} else {
			sourcePageMap[pp.ParentID].LearnMore = append(sourcePageMap[pp.ParentID].LearnMore, pp)
		}
		AddPageToMap(pp.ChildID, resultData.PageMap, explanationLoadOptions)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load explanations: %v", err)
	}
	return nil
}

type ProcessLensCallback func(db *database.DB, lens *Lens) error

// Load all lenses for the given pages
func LoadLenses(db *database.DB, queryPart *database.QueryPart, resultData *CommonHandlerData, callback ProcessLensCallback) error {
	rows := database.NewQuery(`
		SELECT l.id,l.pageId,l.lensId,l.lensIndex,l.lensName,l.lensSubtitle,
			l.createdBy,l.createdAt,l.updatedBy,l.updatedAt
		FROM lenses AS l`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var lens Lens
		err := rows.Scan(&lens.ID, &lens.PageID, &lens.LensID, &lens.LensIndex, &lens.LensName,
			&lens.LensSubtitle, &lens.CreatedBy, &lens.CreatedAt, &lens.UpdatedBy, &lens.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		return callback(db, &lens)
	})
	if err != nil {
		return fmt.Errorf("Couldn't load lenses: %v", err)
	}
	return nil
}

// Load all lenses for the given pages
func LoadLensesForPages(db *database.DB, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIDs := PageIDsListFromMap(sourcePageMap)
	queryPart := database.NewQuery(`
		JOIN`).AddPart(PageInfosTable(resultData.User)).Add(`AS pi
		ON (l.lensId=pi.pageId)`).Add(`
		WHERE l.pageId IN`).AddArgsGroup(pageIDs)
	err := LoadLenses(db, queryPart, resultData, func(db *database.DB, lens *Lens) error {
		sourcePageMap[lens.PageID].Lenses = append(sourcePageMap[lens.PageID].Lenses, lens)
		AddPageToMap(lens.LensID, resultData.PageMap, LensInfoLoadOptions)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load lenses: %v", err)
	}

	for _, p := range sourcePageMap {
		sort.Sort(p.Lenses)
	}
	return nil
}

// Load the given lens
func LoadLens(db *database.DB, id string) (*Lens, error) {
	var lens *Lens
	queryPart := database.NewQuery(`WHERE l.id=?`, id)
	err := LoadLenses(db, queryPart, nil, func(db *database.DB, l *Lens) error {
		lens = l
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load the lens: %v", err)
	} else if lens == nil {
		return nil, fmt.Errorf("Couldn't find the lens")
	}
	return lens, nil
}

// Load parent pages for which the given pages are lenses
func LoadLensParentIDs(db *database.DB, pageMap map[string]*Page, options *LoadDataOptions) error {
	if options == nil {
		options = &LoadDataOptions{}
	}
	sourcePageMap := options.ForPages
	if sourcePageMap == nil {
		sourcePageMap = pageMap
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}

	lensIDs := PageIDsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pageId,lensId
		FROM lenses
		WHERE lensId IN`).AddArgsGroup(lensIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID, lensID string
		err := rows.Scan(&pageID, &lensID)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		sourcePageMap[lensID].LensParentID = pageID
		AddPageIDToMap(pageID, pageMap)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load lens parent ids: %v", err)
	}
	return nil
}

type ProcessPathPageCallback func(db *database.DB, pathPage *PathPage) error

// Load all path pages matching the given query
func LoadPathPages(db *database.DB, queryPart *database.QueryPart, resultData *CommonHandlerData, callback ProcessPathPageCallback) error {
	rows := database.NewQuery(`
		SELECT pathp.id,pathp.guideId,pathp.pathPageId,pathp.pathIndex,
			pathp.createdBy,pathp.createdAt,pathp.updatedBy,pathp.updatedAt
		FROM pathPages AS pathp`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pathPage PathPage
		err := rows.Scan(&pathPage.ID, &pathPage.GuideID, &pathPage.PathPageID, &pathPage.PathIndex,
			&pathPage.CreatedBy, &pathPage.CreatedAt, &pathPage.UpdatedBy, &pathPage.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		return callback(db, &pathPage)
	})
	if err != nil {
		return fmt.Errorf("Couldn't load path pages: %v", err)
	}
	return nil
}

// Load all path pages for the given pages
func LoadPathForPages(db *database.DB, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIDs := PageIDsListFromMap(sourcePageMap)
	queryPart := database.NewQuery(`
		WHERE pathp.guideId IN`).AddArgsGroup(pageIDs)
	err := LoadPathPages(db, queryPart, resultData, func(db *database.DB, pathPage *PathPage) error {
		sourcePageMap[pathPage.GuideID].PathPages = append(sourcePageMap[pathPage.GuideID].PathPages, pathPage)
		AddPageIDToMap(pathPage.PathPageID, resultData.PageMap)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load lenses: %v", err)
	}

	for _, p := range sourcePageMap {
		sort.Sort(p.PathPages)
	}
	return nil
}

// Load all path pages with the given id
func LoadPathPage(db *database.DB, id string) (*PathPage, error) {
	var pathPage *PathPage
	queryPart := database.NewQuery(`WHERE pathp.id=?`, id)
	err := LoadPathPages(db, queryPart, nil, func(db *database.DB, pp *PathPage) error {
		pathPage = pp
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load the path page: %v", err)
	}
	return pathPage, nil
}

type ProcessPathInstanceCallback func(db *database.DB, pathInstance *PathInstance) error

// Load path instances matching the giving query condition
func LoadPathInstances(db *database.DB, queryPart *database.QueryPart, u *CurrentUser, callback ProcessPathInstanceCallback) error {
	rows := database.NewQuery(`
		SELECT pathi.id,pathi.userId,pathi.guideId,pathi.pageIds,pathi.sourcePageIds,pathi.progress,
			pathi.createdAt,pathi.updatedAt,pathi.isFinished
		FROM pathInstances AS pathi`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		instance := NewPathInstance()
		var pageIDs, sourcePageIDs, userID string
		err := rows.Scan(&instance.ID, &userID, &instance.GuideID, &pageIDs, &sourcePageIDs,
			&instance.Progress, &instance.CreatedAt, &instance.UpdatedAt, &instance.IsFinished)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		instance.IsByCurrentUser = userID == u.ID
		pageIdsList := strings.Split(pageIDs, ",")
		sourceIdsList := strings.Split(sourcePageIDs, ",")
		for n, pageID := range pageIdsList {
			instance.Pages = append(instance.Pages, &PathInstancePage{pageID, sourceIdsList[n]})
		}
		return callback(db, instance)
	})
	return err
}

// Load path instance with the given id
func LoadPathInstance(db *database.DB, id string, u *CurrentUser) (*PathInstance, error) {
	var instance *PathInstance
	queryPart := database.NewQuery(`WHERE pathi.id=?`, id)
	err := LoadPathInstances(db, queryPart, u, func(db *database.DB, pathInstance *PathInstance) error {
		instance = pathInstance
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load path instance: %v", err)
	}
	return instance, nil
}
