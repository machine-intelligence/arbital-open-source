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
	DomainPageType   = "domain"

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

	// Mark types
	QueryMarkType     = "query"
	TypoMarkType      = "typo"
	ConfusionMarkType = "confusion"

	// How long the page lock lasts
	PageQuickLockDuration = 5 * 60  // in seconds
	PageLockDuration      = 30 * 60 // in seconds

	RequestForEditTagParentPageId = "3zj"
	MathDomainId                  = "1lw"
)

var (
	// Regexp that strictly matches an alias, and not a page id
	StrictAliasRegexp = regexp.MustCompile("^[0-9A-Za-z_]*[A-Za-z_][0-9A-Za-z_]*$")
)

// corePageData has data we load directly from the pages and pageInfos tables.
type corePageData struct {
	Likeable

	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageId                   string `json:"pageId"`
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
	EditCreatorId            string `json:"editCreatorId"`
	EditCreatedAt            string `json:"editCreatedAt"`
	PageCreatorId            string `json:"pageCreatorId"`
	PageCreatedAt            string `json:"pageCreatedAt"`
	SeeGroupId               string `json:"seeGroupId"`
	EditGroupId              string `json:"editGroupId"`
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

	// The following data is filled on demand.
	Text     string `json:"text"`
	MetaText string `json:"metaText"`
}

// NewCorePageData returns a pointer to a new corePageData object created with the given page id
func NewCorePageData(pageId string) *corePageData {
	data := &corePageData{PageId: pageId}
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
	ViewCount                int  `json:"viewCount"`
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
	TaggedAsIds    []string `json:"taggedAsIds"`
	RelatedIds     []string `json:"relatedIds"`
	RequirementIds []string `json:"requirementIds"`
	SubjectIds     []string `json:"subjectIds"`
	DomainIds      []string `json:"domainIds"`
	ChildIds       []string `json:"childIds"`
	ParentIds      []string `json:"parentIds"`
	MarkIds        []string `json:"markIds"`

	// Lens stuff
	Lenses       LensList `json:"lenses"`
	LensParentId string   `json:"lensParentId"`

	// TODO: eventually move this to the user object (once we have load
	// options + pipeline for users)
	// For user pages, this is the domains user has access to
	DomainMembershipIds []string `json:"domainMembershipIds"`

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

	// Populated for groups
	Members map[string]*Member `json:"members"`

	// === FE data ===
	// This data isn't loaded on the BE, but populated on the FE
	// Map: red alias -> pretty text
	RedAliases map[string]string `json:"redAliases"`
	// List of page's meta tags that indicate the page should be improved
	ImprovementTagIds []string `json:"improvementTagIds"`
	// List of page's tags that are not meta tags
	NonMetaTagIds []string `json:"nonMetaTagIds"`
	// TODOs extracted from the page's text
	Todos []string `json:"todos"`
}

// NewPage returns a pointer to a new page object created with the given page id
func NewPage(pageId string) *Page {
	p := &Page{corePageData: *NewCorePageData(pageId)}
	p.Votes = make([]*Vote, 0)
	p.Summaries = make(map[string]string)
	p.CreatorIds = make([]string, 0)
	p.CommentIds = make([]string, 0)
	p.QuestionIds = make([]string, 0)
	p.TaggedAsIds = make([]string, 0)
	p.RelatedIds = make([]string, 0)
	p.RequirementIds = make([]string, 0)
	p.SubjectIds = make([]string, 0)
	p.DomainIds = make([]string, 0)
	p.ChangeLogs = make([]*ChangeLog, 0)
	p.ChildIds = make([]string, 0)
	p.ParentIds = make([]string, 0)
	p.MarkIds = make([]string, 0)
	p.DomainMembershipIds = make([]string, 0)
	p.Lenses = make(LensList, 0)
	p.DomainSubmissions = make(map[string]*PageToDomainSubmission)
	p.Answers = make([]*Answer, 0)
	p.SearchStrings = make(map[string]string)
	p.Members = make(map[string]*Member)
	p.EditHistory = make(map[string]*EditInfo)
	p.RedAliases = make(map[string]string)
	p.ImprovementTagIds = make([]string, 0)
	p.NonMetaTagIds = make([]string, 0)
	p.Todos = make([]string, 0)

	// NOTE: we want permissions to be explicitly null so that if someone refers to them
	// they get an error. The permissions are only set when they are also fully computed.
	p.Permissions = nil
	return p
}

// Lens connection
type Lens struct {
	Id        int64  `json:"id,string"`
	PageId    string `json:"pageId"`
	LensId    string `json:"lensId"`
	LensIndex int    `json:"lensIndex"`
	LensName  string `json:"lensName"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
	UpdatedBy string `json:"updatedBy"`
	UpdatedAt string `json:"updatedAt"`
}

type LensList []*Lens

func (a LensList) Len() int           { return len(a) }
func (a LensList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a LensList) Less(i, j int) bool { return a[i].LensIndex < a[j].LensIndex }

type Vote struct {
	Value     int    `json:"value"`
	UserId    string `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

type Likeable struct {
	LikeableId   int64  `json:"likeableId,string"`
	LikeableType string `json:"likeableType"`
	MyLikeValue  int    `json:"myLikeValue"`
	LikeCount    int    `json:"likeCount"`
	DislikeCount int    `json:"dislikeCount"`
	// Computed from LikeCount and DislikeCount
	LikeScore int `json:"likeScore"`

	// List of user ids who liked this page
	IndividualLikes []string `json:"individualLikes"`
}

func NewLikeable(likeableType string) *Likeable {
	return &Likeable{
		LikeableType:    likeableType,
		IndividualLikes: make([]string, 0),
	}
}

// ChangeLog describes a row from changeLogs table.
type ChangeLog struct {
	Likeable

	Id               string `json:"id"`
	PageId           string `json:"pageId"`
	UserId           string `json:"userId"`
	Edit             int    `json:"edit"`
	Type             string `json:"type"`
	CreatedAt        string `json:"createdAt"`
	AuxPageId        string `json:"auxPageId"`
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
	IsSubmitted         bool   `json:"isSubmitted"`
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

// Information about a page which was submitted to a domain
type PageToDomainSubmission struct {
	PageId      string `json:"pageId"`
	DomainId    string `json:"domainId"`
	CreatedAt   string `json:"createdAt"`
	SubmitterId string `json:"submitterId"`
	ApprovedAt  string `json:"approvedAt"`
	ApproverId  string `json:"approverId"`
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

// GlobalHandlerData includes misc data that we send to the FE with each request
// that resets FE data.
type GlobalHandlerData struct {
	// Id of the private group the current user is in
	PrivateGroupId string `json:"privateGroupId"`
	// List of all domains
	DomainIds []string `json:"domainIds"`
	// List of tags ids that mean a page should be improved
	ImprovementTagIds []string `json:"improvementTagIds"`
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

	GlobalData *GlobalHandlerData `json:"globalData"`
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
	err = LoadLensesForPages(db, data, &LoadDataOptions{
		ForPages: filteredPageMap,
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

	if u.Id != "" {
		// Load user's marks
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
	err = LoadChangeLogsForPages(db, u, data, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadChangeLogsForPages failed: %v", err)
	}

	// Load search strings
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.SearchStrings })
	err = LoadSearchStrings(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSearchStrings failed: %v", err)
	}

	// Load whether the pages are lenses for other pages
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.LensParentId })
	err = LoadLensParentIds(db, pageMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadLensParentIds failed: %v", err)
	}

	// Load whether or not the pages have an unpublished draft
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.HasDraft })
	err = LoadDraftExistence(db, u.Id, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadDraftExistence failed: %v", err)
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
		return fmt.Errorf("LoadVotes failed: %v", err)
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

	// Load pages' edit history
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.EditHistory })
	err = LoadEditHistory(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadEditHistory failed: %v", err)
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
		AddUserIdToMap(p.PageCreatorId, userMap)
		AddUserIdToMap(p.EditCreatorId, userMap)
		if IsIdValid(p.LockedBy) {
			AddUserIdToMap(p.LockedBy, userMap)
		}
	}
	err = LoadUsers(db, userMap, u.Id)
	if err != nil {
		return fmt.Errorf("LoadUsers failed: %v", err)
	}

	// Computed which pages count as visited.
	// TODO: refactor this code to use multiple-maps insert wrapper
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
			pi.isEditorComment,pi.isEditorCommentIntention,pi.isResolved,
			pi.isRequisite,pi.indirectTeacher,pi.currentEdit,pi.likeableId,p.isAutosave,p.isSnapshot,
			p.isLiveEdit,p.isMinorEdit,p.editSummary,pi.isDeleted,pi.mergedInto,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset
		FROM pages AS p
		JOIN`).AddPart(pageInfosTable).Add(`AS pi
		ON (p.pageId = pi.pageId AND p.edit = pi.currentEdit)
		WHERE p.pageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		p := NewCorePageData("")
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.PrevEdit, &p.EditCreatorId, &p.EditCreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.Type, &p.HasVote,
			&p.VoteType, &p.Alias, &p.PageCreatedAt, &p.PageCreatorId, &p.SortChildrenBy,
			&p.SeeGroupId, &p.EditGroupId, &p.IsEditorComment, &p.IsEditorCommentIntention,
			&p.IsResolved, &p.IsRequisite, &p.IndirectTeacher, &p.CurrentEdit, &p.LikeableId,
			&p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit, &p.EditSummary, &p.IsDeleted, &p.MergedInto,
			&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("Failed to scan a page: %v", err)
		}
		pageMap[p.PageId].corePageData = *p
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

// LoadEditHistory loads edit histories for given pages
func LoadEditHistory(db *database.DB, pageMap map[string]*Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsListFromMap(pageMap)

	rows := database.NewQuery(`
		SELECT pageId,edit,prevEdit
		FROM pages
		WHERE pageId IN`).AddArgsGroup(pageIds).Add(`
			AND NOT isSnapshot AND NOT isAutosave`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		var editInfo EditInfo
		err := rows.Scan(&pageId, &editInfo.Edit, &editInfo.PrevEdit)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		pageMap[pageId].EditHistory[fmt.Sprintf("%d", editInfo.Edit)] = &editInfo
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

type ProcessChangeLogCallback func(db *database.DB, changeLog *ChangeLog) error

// LoadChangeLogs loads the change logs matching the given condition.
func LoadChangeLogs(db *database.DB, queryPart *database.QueryPart, resultData *CommonHandlerData, callback ProcessChangeLogCallback) error {
	rows := database.NewQuery(`
			SELECT id,pageId,userId,edit,type,createdAt,auxPageId,likeableId,oldSettingsValue,newSettingsValue
			FROM changeLogs`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		l := NewChangeLog()
		err := rows.Scan(&l.Id, &l.PageId, &l.UserId, &l.Edit, &l.Type, &l.CreatedAt,
			&l.AuxPageId, &l.LikeableId, &l.OldSettingsValue, &l.NewSettingsValue)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		if resultData != nil {
			AddUserToMap(l.UserId, resultData.UserMap)
			AddPageToMap(l.AuxPageId, resultData.PageMap, TitlePlusIncludeDeletedLoadOptions)
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
			WHERE pageId=?`, p.PageId).Add(`
				AND (userId=? OR type!=?)`, u.Id, NewSnapshotChangeLog).Add(`
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
func LoadChangeLogsByIds(db *database.DB, ids []string, typeConstraint string) (map[string]*ChangeLog, error) {
	changeLogs := make(map[string]*ChangeLog)
	queryPart := database.NewQuery(`
			WHERE id IN`).AddArgsGroupStr(ids).Add(`
				AND type=?`, typeConstraint)
	err := LoadChangeLogs(db, queryPart, nil, func(db *database.DB, changeLog *ChangeLog) error {
		changeLogs[changeLog.Id] = changeLog
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
		if changeLog.LikeableId != 0 {
			likeablesMap[changeLog.LikeableId] = &changeLog.Likeable
		}
	}
	return LoadLikes(db, u, likeablesMap, nil, nil)
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
				WHERE p.pageId=?`, pageId).Add(`
					AND (p.creatorId=? OR (NOT p.isSnapshot AND NOT p.isAutosave))`, u.Id).Add(`
					/* To consider a snapshot, it has to be based on the current edit */
					AND (NOT p.isSnapshot OR pi.currentEdit=0 OR p.prevEdit=pi.currentEdit)`).AddPart(orderClause).Add(`
				LIMIT 1
			)`)
	}
	statement := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,pi.type,p.title,p.clickbait,p.text,p.metaText,
			pi.alias,p.creatorId,pi.sortChildrenBy,pi.hasVote,pi.voteType,
			p.createdAt,pi.seeGroupId,pi.editGroupId,pi.createdAt,
			pi.createdBy,pi.isEditorComment,pi.isEditorCommentIntention,
			pi.isResolved,pi.likeableId,p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,p.editSummary,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset,
			pi.currentEdit>0,pi.isDeleted,pi.mergedInto,pi.currentEdit,pi.maxEdit,pi.lockedBy,pi.lockedUntil,
			pi.voteType,pi.isRequisite,pi.indirectTeacher
		FROM pages AS p
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (p.pageId=pi.pageId AND p.pageId=?)`, pageId).Add(`
		WHERE`).AddPart(whereClause).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageId, &p.Edit, &p.PrevEdit, &p.Type, &p.Title, &p.Clickbait,
		&p.Text, &p.MetaText, &p.Alias, &p.EditCreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.EditCreatedAt, &p.SeeGroupId,
		&p.EditGroupId, &p.PageCreatedAt, &p.PageCreatorId,
		&p.IsEditorComment, &p.IsEditorCommentIntention, &p.IsResolved, &p.LikeableId,
		&p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit, &p.EditSummary,
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

// LoadLikes loads likes corresponding to the given likeable objects.
func LoadLikes(db *database.DB, u *CurrentUser, likeablesMap map[int64]*Likeable, individualLikesPageMap map[int64]*Likeable, userMap map[string]*User) error {
	if len(likeablesMap) <= 0 {
		return nil
	}

	likeableIds := make([]interface{}, 0)
	for id, _ := range likeablesMap {
		likeableIds = append(likeableIds, id)
	}

	rows := database.NewQuery(`
		SELECT likeableId,userId,value
		FROM likes
		WHERE likeableId IN`).AddArgsGroup(likeableIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likeableId int64
		var userId string
		var value int
		err := rows.Scan(&likeableId, &userId, &value)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		likeable := likeablesMap[likeableId]
		// We count the current user's like value towards the sum here in the FE.
		if userId == u.Id {
			likeable.MyLikeValue = value
		} else if value > 0 {
			if likeable.LikeCount >= likeable.DislikeCount {
				likeable.LikeScore++
			} else {
				likeable.LikeScore += 2
			}
			likeable.LikeCount++
		} else if value < 0 {
			if likeable.DislikeCount >= likeable.LikeCount {
				likeable.LikeScore--
			}
			likeable.DislikeCount++
		}

		// Store the like itself for pages that want it
		if individualLikesPageMap != nil {
			if likeable, ok := individualLikesPageMap[likeableId]; ok {
				likeable.IndividualLikes = append(likeable.IndividualLikes, userId)
				AddUserIdToMap(userId, userMap)
			}
		}
		return nil
	})
	return err
}

// LoadLikesForPages loads likes corresponding to the given pages and updates the pages.
func LoadLikesForPages(db *database.DB, u *CurrentUser, pageMap map[string]*Page, individualLikesPageMap map[string]*Page, userMap map[string]*User) error {
	likeablesMap := make(map[int64]*Likeable)
	for _, page := range pageMap {
		if page.LikeableId != 0 {
			likeablesMap[page.LikeableId] = &page.Likeable
		}
	}
	individualLikeablesMap := make(map[int64]*Likeable)
	for _, page := range individualLikesPageMap {
		if page.LikeableId != 0 {
			individualLikeablesMap[page.LikeableId] = &page.Likeable
		}
	}
	return LoadLikes(db, u, likeablesMap, individualLikeablesMap, userMap)
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

type ProcessLoadPageToDomainSubmissionCallback func(db *database.DB, submission *PageToDomainSubmission) error

// LoadPageToDomainSubmissions loads information the pages that have been submitted to a domain
func LoadPageToDomainSubmissions(db *database.DB, queryPart *database.QueryPart, callback ProcessLoadPageToDomainSubmissionCallback) error {
	rows := database.NewQuery(`
		SELECT pageId,domainId,createdAt,submitterId,approvedAt,approverId
		FROM pageToDomainSubmissions`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var submission PageToDomainSubmission
		err := rows.Scan(&submission.PageId, &submission.DomainId, &submission.CreatedAt,
			&submission.SubmitterId, &submission.ApprovedAt, &submission.ApproverId)
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

	pageIds := PageIdsListFromMap(sourceMap)
	if len(pageIds) <= 0 {
		return nil
	}

	queryPart := database.NewQuery(`WHERE pageId IN`).AddArgsGroup(pageIds)
	err := LoadPageToDomainSubmissions(db, queryPart, func(db *database.DB, submission *PageToDomainSubmission) error {
		AddPageIdToMap(submission.DomainId, pageMap)
		AddUserToMap(submission.SubmitterId, userMap)
		AddUserToMap(submission.ApproverId, userMap)
		pageMap[submission.PageId].DomainSubmissions[submission.DomainId] = submission
		return nil
	})
	return err
}

// LoadPageToDomainSubmission loads information about a specific page that was submitted to a specific domain
func LoadPageToDomainSubmission(db *database.DB, pageId, domainId string) (*PageToDomainSubmission, error) {
	queryPart := database.NewQuery(`
		WHERE pageId=?`, pageId).Add(`
			AND domainId=?`, domainId).Add(`
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

	// Only load the marks the current user created
	constraint := database.NewQuery(``)
	if options.CurrentUserConstraint {
		constraint = database.NewQuery(`AND m.creatorId=?`, u.Id)
	}

	// Only load for pages in which current user is an author
	pageIdsPart := database.NewQuery(``).AddArgsGroup(pageIds)
	if options.EditorConstraint {
		pageIdsPart = database.NewQuery(`(
			SELECT p.pageId
			FROM pages AS p
			WHERE p.pageId IN`).AddArgsGroup(pageIds).Add(`
				AND NOT p.isSnapshot AND NOT p.isAutosave
				AND p.creatorId=?`, u.Id).Add(`
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
			text,requisiteSnapshotId,resolvedPageId,resolvedBy,answered,isSubmitted
		FROM marks
		WHERE id IN` + database.InArgsPlaceholder(len(markIds))).Query(markIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var creatorId string
		mark := &Mark{}
		err := rows.Scan(&mark.Id, &mark.Type, &mark.PageId, &mark.CreatorId, &mark.CreatedAt,
			&mark.AnchorContext, &mark.AnchorText, &mark.AnchorOffset, &mark.Text,
			&mark.RequisiteSnapshotId, &mark.ResolvedPageId, &mark.ResolvedBy, &mark.Answered,
			&mark.IsSubmitted)
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

// Given a parent page that collects meta tags, load all its children.
func LoadMetaTags(db *database.DB, parentId string) ([]string, error) {
	pageMap := make(map[string]*Page)
	page := AddPageIdToMap(parentId, pageMap)
	options := &LoadChildIdsOptions{
		ForPages:     pageMap,
		Type:         WikiPageType,
		PagePairType: ParentPagePairType,
		LoadOptions:  EmptyLoadOptions,
	}
	err := LoadChildIds(db, pageMap, nil, options)
	if err != nil {
		return nil, err
	}
	return page.ChildIds, nil
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
			AND pi.type!=? AND pi.type!=?`, CommentPageType, QuestionPageType).Add(`
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
		SELECT toId,asMaintainer
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageId string
		var asMaintainer bool
		err := rows.Scan(&toPageId, &asMaintainer)
		if err != nil {
			return fmt.Errorf("Failed to scan for a subscription: %v", err)
		}
		pageMap[toPageId].IsSubscribed = true
		pageMap[toPageId].IsSubscribedAsMaintainer = asMaintainer
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
		SELECT toId,COUNT(*),SUM(asMaintainer)
		FROM subscriptions
		WHERE userId!=?`, currentUserId).Add(`
			AND toId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageId string
		var subscriberCount, maintainerCount int
		err := rows.Scan(&toPageId, &subscriberCount, &maintainerCount)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		pageMap[toPageId].SubscriberCount = subscriberCount
		pageMap[toPageId].MaintainerCount = maintainerCount
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
	if len(aliases) <= 0 {
		return aliasToIdMap, nil
	}

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
	// TODO: refactor these queries into one query + additional parts
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
	return aliasToIdMap, nil
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

type ProcessLensCallback func(db *database.DB, lens *Lens) error

// Load all lenses for the given pages
func LoadLenses(db *database.DB, queryPart *database.QueryPart, resultData *CommonHandlerData, callback ProcessLensCallback) error {
	rows := database.NewQuery(`
		SELECT l.id,l.pageId,l.lensId,l.lensIndex,l.lensName,l.createdBy,l.createdAt,l.updatedBy,l.updatedAt
		FROM lenses AS l`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var lens Lens
		err := rows.Scan(&lens.Id, &lens.PageId, &lens.LensId, &lens.LensIndex, &lens.LensName,
			&lens.CreatedBy, &lens.CreatedAt, &lens.UpdatedBy, &lens.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		if resultData != nil {
			AddPageToMap(lens.LensId, resultData.PageMap, LensInfoLoadOptions)
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

	pageIds := PageIdsListFromMap(sourcePageMap)
	queryPart := database.NewQuery(`
		JOIN`).AddPart(PageInfosTable(resultData.User)).Add(`AS pi
		ON (l.pageId=pi.pageId)`).Add(`
		WHERE l.pageId IN`).AddArgsGroup(pageIds)
	err := LoadLenses(db, queryPart, resultData, func(db *database.DB, lens *Lens) error {
		sourcePageMap[lens.PageId].Lenses = append(sourcePageMap[lens.PageId].Lenses, lens)
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

// Load parent pages for which the given pages are lenses
func LoadLensParentIds(db *database.DB, pageMap map[string]*Page, options *LoadDataOptions) error {
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

	lensIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pageId,lensId
		FROM lenses
		WHERE lensId IN`).AddArgsGroup(lensIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, lensId string
		err := rows.Scan(&pageId, &lensId)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		sourcePageMap[lensId].LensParentId = pageId
		AddPageIdToMap(pageId, pageMap)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load lens parent ids: %v", err)
	}
	return nil
}
