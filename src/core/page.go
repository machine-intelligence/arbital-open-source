// page.go contains all the page stuff
package core

import (
	"fmt"
	"regexp"
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
	NewParentChangeLog         = "newParent"
	DeleteParentChangeLog      = "deleteParent"
	NewChildChangeLog          = "newChild"
	DeleteChildChangeLog       = "deleteChild"
	NewTagChangeLog            = "newTag"
	DeleteTagChangeLog         = "deleteTag"
	NewUsedAsTagChangeLog      = "newUsedAsTag"
	DeleteUsedAsTagChangeLog   = "deleteUsedAsTag"
	NewRequirementChangeLog    = "newRequirement"
	DeleteRequirementChangeLog = "deleteRequirement"
	NewRequiredByChangeLog     = "newRequiredBy"
	DeleteRequiredByChangeLog  = "deleteRequiredBy"
	NewSubjectChangeLog        = "newSubject"
	DeleteSubjectChangeLog     = "deleteSubject"
	NewTeacherChangeLog        = "newTeacher"
	DeleteTeacherChangeLog     = "deleteTeacher"
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

// corePageData has data we load directly from pages table.
type corePageData struct {
	// === Basic data. ===
	// Any time we load a page, you can at least expect all this data.
	PageId            string `json:"pageId"`
	Edit              int    `json:"edit"`
	PrevEdit          int    `json:"prevEdit"`
	Type              string `json:"type"`
	Title             string `json:"title"`
	Clickbait         string `json:"clickbait"`
	TextLength        int    `json:"textLength"` // number of characters
	Alias             string `json:"alias"`
	SortChildrenBy    string `json:"sortChildrenBy"`
	HasVote           bool   `json:"hasVote"`
	VoteType          string `json:"voteType"`
	CreatorId         string `json:"creatorId"`
	CreatedAt         string `json:"createdAt"`
	OriginalCreatedAt string `json:"originalCreatedAt"`
	OriginalCreatedBy string `json:"originalCreatedBy"`
	EditKarmaLock     int    `json:"editKarmaLock"`
	SeeGroupId        string `json:"seeGroupId"`
	EditGroupId       string `json:"editGroupId"`
	IsAutosave        bool   `json:"isAutosave"`
	IsSnapshot        bool   `json:"isSnapshot"`
	IsLiveEdit        bool   `json:"isLiveEdit"`
	IsMinorEdit       bool   `json:"isMinorEdit"`
	IsRequisite       bool   `json:"isRequisite"`
	IndirectTeacher   bool   `json:"indirectTeacher"`
	TodoCount         int    `json:"todoCount"`
	LensIndex         int    `json:"lensIndex"`
	IsEditorComment   bool   `json:"isEditorComment"`
	SnapshotText      string `json:"snapshotText"`
	AnchorContext     string `json:"anchorContext"`
	AnchorText        string `json:"anchorText"`
	AnchorOffset      int    `json:"anchorOffset"`

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
	// True iff there is an edit that has isLiveEdit set for this page
	WasPublished bool    `json:"wasPublished"`
	Votes        []*Vote `json:"votes"`
	// We don't allow users to change the vote type once a page has been published
	// with a voteType!="" even once. If it has, this is the vote type it shall
	// always have.
	LockedVoteType string `json:"lockedVoteType"`
	// Highest edit number used for this page for all users
	MaxEditEver  int `json:"maxEditEver"`
	RedLinkCount int `json:"redLinkCount"`
	// Set to pageId corresponding to the question/answer the user started creating for this page
	ChildDraftId string `json:"childDraftId"`
	// Page is locked by this user
	LockedBy string `json:"lockedBy"`
	// User has the page lock until this time
	LockedUntil string `json:"lockedUntil"`
	NextPageId  string `json:"nextPageId"`
	PrevPageId  string `json:"prevPageId"`
	// Whether or not the page is used as a requirement
	UsedAsMastery bool `json:"usedAsMastery"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".
	Summaries map[string]string `json:"summaries"`
	// Ids of the users who edited this page. Ordered by how much they contributed.
	CreatorIds []string `json:"creatorIds"`

	// Subpages.
	AnswerIds      []string `json:"answerIds"`
	CommentIds     []string `json:"commentIds"`
	QuestionIds    []string `json:"questionIds"`
	LensIds        []string `json:"lensIds"`
	TaggedAsIds    []string `json:"taggedAsIds"`
	RelatedIds     []string `json:"relatedIds"`
	RequirementIds []string `json:"requirementIds"`
	SubjectIds     []string `json:"subjectIds"`

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
	UserId           string `json:"userId"`
	Edit             int    `json:"edit"`
	Type             string `json:"type"`
	CreatedAt        string `json:"createdAt"`
	AuxPageId        string `json:"auxPageId"`
	OldSettingsValue string `json:"oldSettingsValue"`
	NewSettingsValue string `json:"newSettingsValue"`
}

// Mastery is a page you should have mastered before you can understand another page.
type Mastery struct {
	PageId    string `json:"pageId"`
	Has       bool   `json:"has"`
	Wants     bool   `json:"wants"`
	UpdatedAt string `json:"updatedAt"`
}

// PageObject stores some information for an object embedded in a page
type PageObject struct {
	PageId string `json:"pageId"`
	Edit   int    `json:"edit"`
	Object string `json:"object"`
	Value  string `json:"value"`
}

// CommonHandlerData is what handlers fill out and return
type CommonHandlerData struct {
	// If set, then this packet should reset everything on the FE
	ResetEverything bool
	// Optional user object with the current user's data
	User *user.User
	// Map of page id -> currently live version of the page
	PageMap map[string]*Page
	// Map of page id -> some edit of the page
	EditMap    map[string]*Page
	UserMap    map[string]*User
	MasteryMap map[string]*Mastery
	// Page id -> {object alias -> object}
	PageObjectMap map[string]map[string]*PageObject
	// ResultMap contains various data the specific handler returns
	ResultMap map[string]interface{}
}

// NewHandlerData creates and initializes a new commonHandlerData object.
func NewHandlerData(u *user.User, resetEverything bool) *CommonHandlerData {
	var data CommonHandlerData
	data.User = u
	data.ResetEverything = resetEverything
	data.PageMap = make(map[string]*Page)
	data.EditMap = make(map[string]*Page)
	data.UserMap = make(map[string]*User)
	data.MasteryMap = make(map[string]*Mastery)
	data.PageObjectMap = make(map[string]map[string]*PageObject)
	data.ResultMap = make(map[string]interface{})
	return &data
}

// ToJson puts together the data into one "json" object, so we
// can send it to the front-end.
func (data *CommonHandlerData) ToJson() map[string]interface{} {
	jsonData := make(map[string]interface{})

	jsonData["resetEverything"] = data.ResetEverything

	if data.User != nil {
		jsonData["user"] = data.User
	}

	jsonData["pages"] = data.PageMap
	jsonData["edits"] = data.EditMap
	jsonData["users"] = data.UserMap
	jsonData["masteries"] = data.MasteryMap
	jsonData["pageObjects"] = data.PageObjectMap
	jsonData["result"] = data.ResultMap
	return jsonData
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

	// Load answers
	filteredPageMap := filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Answers })
	err := LoadChildIds(db, pageMap, u, &LoadChildIdsOptions{
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

	// Load whether the page is used as a requirement
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.UsedAsMastery })
	err = LoadUsedAsMastery(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadUsedAsMastery failed: %v", err)
	}

	// Load pages' creator's ids
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Creators })
	err = LoadCreatorIds(db, pageMap, userMap, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return fmt.Errorf("LoadCreatorIds failed: %v", err)
	}

	// Add other pages we'll need
	AddUserGroupIdsToPageMap(u, pageMap)

	// Load page data
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return !p.LoadOptions.Edit })
	err = LoadPages(db, u, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadPages failed: %v", err)
	}
	// Load summaries
	filteredPageMap = filterPageMap(pageMap, func(p *Page) bool { return p.LoadOptions.Summaries })
	err = LoadSummaries(db, filteredPageMap)
	if err != nil {
		return fmt.Errorf("LoadSummaries failed: %v", err)
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
		for id, p := range pageMap {
			if p.Text != "" {
				visitedValues = append(visitedValues, visitorId, id, database.Now())
			}
		}
	}

	// Add a visit to pages for which we loaded text.
	if len(visitedValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO visits (userId, pageId, createdAt)
			VALUES ` + database.ArgsPlaceholder(len(visitedValues), 3))
		if _, err = statement.Exec(visitedValues...); err != nil {
			return fmt.Errorf("Couldn't update visits", err)
		}
	}

	return nil
}

// LoadMasteries loads the masteries.
func LoadMasteries(db *database.DB, u *user.User, masteryMap map[string]*Mastery) error {
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
func LoadPageObjects(db *database.DB, u *user.User, pageMap map[string]*Page, pageObjectMap map[string]map[string]*PageObject) error {
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
func LoadPages(db *database.DB, u *user.User, pageMap map[string]*Page) error {
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
			length(p.text),p.metaText,pi.type,pi.editKarmaLock,pi.hasVote,pi.voteType,
			pi.alias,pi.createdAt,pi.createdBy,pi.sortChildrenBy,pi.seeGroupId,pi.editGroupId,
			pi.lensIndex,pi.isEditorComment,pi.isRequisite,pi.indirectTeacher,
			p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId = pi.pageId AND p.isLiveEdit)
		WHERE p.pageId IN`).AddArgsGroup(pageIds).Add(`
			AND (pi.seeGroupId=0 OR pi.seeGroupId IN`).AddIdsGroupStr(u.GroupIds).Add(`)
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p corePageData
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.PrevEdit, &p.CreatorId, &p.CreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.Type, &p.EditKarmaLock, &p.HasVote,
			&p.VoteType, &p.Alias, &p.OriginalCreatedAt, &p.OriginalCreatedBy, &p.SortChildrenBy,
			&p.SeeGroupId, &p.EditGroupId, &p.LensIndex, &p.IsEditorComment, &p.IsRequisite, &p.IndirectTeacher,
			&p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit,
			&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("Failed to scan a page: %v", err)
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
			return fmt.Errorf("failed to scan a page: %v", err)
		}
		pageMap[pageId].Summaries[name] = text
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
		SELECT userId,edit,type,createdAt,auxPageId,oldSettingsValue,newSettingsValue
		FROM changeLogs
		WHERE pageId=?`, p.PageId).Add(`
			AND (userId=? OR type!=?)`, userId, NewSnapshotChangeLog).Add(`
		ORDER BY createdAt DESC`).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var l ChangeLog
			err := rows.Scan(&l.UserId, &l.Edit, &l.Type, &l.CreatedAt, &l.AuxPageId, &l.OldSettingsValue, &l.NewSettingsValue)
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
			AddPageToMap(log.AuxPageId, pageMap, TitlePlusLoadOptions)
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
func LoadFullEdit(db *database.DB, pageId, userId string, options *LoadEditOptions) (*Page, error) {
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
				JOIN pageInfos AS pi
				ON (p.pageId=pi.pageId)
				WHERE p.pageId=? AND (p.prevEdit=pi.currentEdit OR p.isLiveEdit OR p.isAutosave) AND
					(p.creatorId=? OR NOT (p.isSnapshot OR p.isAutosave))
				ORDER BY IF(p.isAutosave,"z",p.createdAt) DESC
				LIMIT 1
			)`, pageId, userId)
	}
	statement := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,pi.type,p.title,p.clickbait,p.text,p.metaText,
			pi.alias,p.creatorId,pi.sortChildrenBy,pi.hasVote,pi.voteType,
			p.createdAt,pi.editKarmaLock,pi.seeGroupId,pi.editGroupId,pi.createdAt,
			pi.createdBy,pi.lensIndex,pi.isEditorComment,p.isAutosave,p.isSnapshot,p.isLiveEdit,p.isMinorEdit,
			p.todoCount,p.snapshotText,p.anchorContext,p.anchorText,p.anchorOffset,
			pi.currentEdit>0,pi.currentEdit,pi.maxEdit,pi.lockedBy,pi.lockedUntil,
			pi.voteType,pi.isRequisite,pi.indirectTeacher
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId AND p.pageId=?)`, pageId).Add(`
		WHERE`).AddPart(whereClause).Add(`AND
			(pi.seeGroupId=0 OR pi.seeGroupId IN (SELECT groupId FROM groupMembers WHERE userId=?))`, userId).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageId, &p.Edit, &p.PrevEdit, &p.Type, &p.Title, &p.Clickbait,
		&p.Text, &p.MetaText, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.EditKarmaLock, &p.SeeGroupId,
		&p.EditGroupId, &p.OriginalCreatedAt, &p.OriginalCreatedBy, &p.LensIndex,
		&p.IsEditorComment, &p.IsAutosave, &p.IsSnapshot, &p.IsLiveEdit, &p.IsMinorEdit,
		&p.TodoCount, &p.SnapshotText, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset, &p.WasPublished,
		&p.CurrentEdit, &p.MaxEditEver, &p.LockedBy, &p.LockedUntil, &p.LockedVoteType,
		&p.IsRequisite, &p.IndirectTeacher)
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

// LoadChildDrafts loads a potentially existing draft for the given page. If it's
// loaded, it'll be added to the given map.
func LoadChildDrafts(db *database.DB, userId string, options *LoadDataOptions) error {
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
					HAVING SUM(p.isLiveEdit)<=0
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
func LoadLikes(db *database.DB, currentUserId string, pageMap map[string]*Page) error {
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
		var userId string
		var pageId string
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
			return fmt.Errorf("failed to scan for a like: %v", err)
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
func LoadRedLinkCount(db *database.DB, pageMap map[string]*Page) error {
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
func LoadCreatorIds(db *database.DB, pageMap map[string]*Page, userMap map[string]*User, options *LoadDataOptions) error {
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
		SELECT pageId,editGroupId
		FROM pageInfos
		WHERE pageId IN`).AddArgsGroup(pageIdsList).Add(`
			AND editGroupId!=""`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, editGroupId string
		err := rows.Scan(&pageId, &editGroupId)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		AddPageToMap(editGroupId, pageMap, TitlePlusLoadOptions)
		return nil
	})
	return err
}

// LoadLinks loads the links for the given pages, and adds them to the pageMap.
func LoadLinks(db *database.DB, pageMap map[string]*Page, options *LoadDataOptions) error {
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
			FROM pageInfos
			WHERE currentEdit>0 AND alias IN`).AddArgsGroup(aliasesList).ToStatement(db).Query()
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
func LoadChildIds(db *database.DB, pageMap map[string]*Page, u *user.User, options *LoadChildIdsOptions) error {
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
		ON (pi.pageId=pp.childId AND pi.currentEdit>0 AND pi.type=?)`, options.Type).Add(`
		WHERE (pi.seeGroupId=0 OR pi.seeGroupId IN`).AddIdsGroupStr(u.GroupIds).Add(`)
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		var ppType string
		err := rows.Scan(&parentId, &childId, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(childId, pageMap, options.LoadOptions)

		parent := sourcePageMap[parentId]
		if options.Type == LensPageType {
			parent.LensIds = append(parent.LensIds, newPage.PageId)
			newPage.ParentIds = append(newPage.ParentIds, parent.PageId)
		} else if options.Type == AnswerPageType {
			parent.AnswerIds = append(parent.AnswerIds, newPage.PageId)
		} else if options.Type == CommentPageType {
			parent.CommentIds = append(parent.CommentIds, newPage.PageId)
		} else if options.Type == QuestionPageType {
			parent.QuestionIds = append(parent.QuestionIds, newPage.PageId)
		} else if options.Type == WikiPageType && options.PagePairType == ParentPagePairType {
			parent.ChildIds = append(parent.ChildIds, childId)
			parent.HasChildren = true
			if parent.LoadOptions.HasGrandChildren {
				newPage.LoadOptions.SubpageCounts = true
			}
			if parent.LoadOptions.RedLinkCountForChildren {
				newPage.LoadOptions.RedLinkCount = true
			}
		} else if options.Type == WikiPageType && options.PagePairType == TagPagePairType {
			parent.RelatedIds = append(parent.RelatedIds, childId)
		}
		return nil
	})
	return err
}

// LoadSubpageCounts loads the number of various types of children the pages have
func LoadSubpageCounts(db *database.DB, pageMap map[string]*Page) error {
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
		var pageId string
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
func LoadParentIds(db *database.DB, pageMap map[string]*Page, u *user.User, options *LoadParentIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	newPages := make(map[string]*Page)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE type=?`, options.PagePairType).Add(`AND childId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN pageInfos AS pi
		ON (pi.pageId=pp.parentId)`).Add(`
		WHERE (pi.seeGroupId=0 OR pi.seeGroupId IN`).AddIdsGroupStr(u.GroupIds).Add(`)
			AND (pi.currentEdit>0 || pp.parentId=pp.childId)
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		var ppType string
		err := rows.Scan(&parentId, &childId, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(parentId, pageMap, options.LoadOptions)
		childPage := pageMap[childId]

		if options.PagePairType == ParentPagePairType {
			childPage.ParentIds = append(childPage.ParentIds, parentId)
			childPage.HasParents = true
			newPages[newPage.PageId] = newPage
		} else if options.PagePairType == RequirementPagePairType {
			childPage.RequirementIds = append(childPage.RequirementIds, parentId)
			options.MasteryMap[parentId] = &Mastery{PageId: parentId}
		} else if options.PagePairType == TagPagePairType {
			childPage.TaggedAsIds = append(childPage.TaggedAsIds, parentId)
		} else if options.PagePairType == SubjectPagePairType {
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
func LoadCommentIds(db *database.DB, pageMap map[string]*Page, options *LoadDataOptions) error {
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
func loadOrderedChildrenIds(db *database.DB, parentId string, sortType string) ([]string, error) {
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
		JOIN pageInfos AS pi
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
func loadSiblingId(db *database.DB, pageId string, useNextSibling bool) (string, error) {
	// Load the parent and the sorting order
	var parentId string
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
		return "", err
	} else if !found || parentCount != 1 {
		return "", nil
	}

	// Load the sibling pages in order
	orderedSiblingIds, err := loadOrderedChildrenIds(db, parentId, sortType)
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
	return "", nil
}

// LoadNextPrevPageIds loads the pages that come before / after the given page
// in the learning list.
func LoadNextPrevPageIds(db *database.DB, userId string, options *LoadDataOptions) error {
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
		if !IsIdValid(p.NextPageId) {
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

// LoadAliasToPageIdMap loads the mapping from aliases to page ids.
func LoadAliasToPageIdMap(db *database.DB, aliases []string) (map[string]string, error) {
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
			FROM pageInfos
			WHERE currentEdit>0 AND (alias IN`).AddArgsGroupStr(strictAliases).Add(`)`).ToStatement(db)
		} else if len(strictAliases) <= 0 {
			query = database.NewQuery(`
			SELECT pageId,alias
			FROM pageInfos
			WHERE currentEdit>0 AND (pageId IN`).AddArgsGroupStr(strictPageIds).Add(`)`).ToStatement(db)
		} else {
			query = database.NewQuery(`
			SELECT pageId,alias
			FROM pageInfos
			WHERE currentEdit>0 AND (pageId IN`).AddArgsGroupStr(strictPageIds).Add(`OR alias IN`).AddArgsGroupStr(strictAliases).Add(`)`).ToStatement(db)
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

		// The query only gets results for when currentEdit > 0
		// We also want to return the pageIds even if they aren't for valid pages
		for _, pageId := range strictPageIds {
			aliasToIdMap[strings.ToLower(pageId)] = strings.ToLower(pageId)
		}
	}
	return aliasToIdMap, nil
}

// LoadOldAliasToPageId converts the given old (base 10) page alias to page id.
func LoadOldAliasToPageId(db *database.DB, alias string) (string, bool, error) {
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

	pageId, ok, err := LoadAliasToPageId(db, aliasToUse)
	if err != nil {
		return "", false, fmt.Errorf("Couldn't convert alias", err)
	}

	return pageId, ok, nil
}

// LoadAliasToPageId converts the given page alias to page id.
func LoadAliasToPageId(db *database.DB, alias string) (string, bool, error) {
	aliasToIdMap, err := LoadAliasToPageIdMap(db, []string{alias})
	if err != nil {
		return "", false, err
	}
	pageId, ok := aliasToIdMap[strings.ToLower(alias)]
	return pageId, ok, nil
}

// LoadAliasAndPageId returns both the alias and the pageId for a given alias or pageId.
func LoadAliasAndPageId(db *database.DB, alias string) (string, string, bool, error) {
	aliasToIdMap, err := LoadAliasToPageIdMap(db, []string{alias})
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
