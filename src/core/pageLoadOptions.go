package core

import (
	"reflect"
)

// PageLoadOptions contains flags for what parts of the page to load. The flags
// are listed IN ORDER they are used for loading.
// When adding a new flag, make sure it goes in the right order, and make sure
// that the default (false) value results in minimal work. Don't forget to update
// existing loadOptions templates below.
// Keep in mind that we check each option only once for all pages.
type PageLoadOptions struct {
	// Load options for subpages
	Comments  bool
	Questions bool

	// Load options for titlePlus pages
	Children                bool
	HasGrandChildren        bool
	SubpageCounts           bool
	AnswerCounts            bool
	RedLinkCountForChildren bool
	Parents                 bool
	Tags                    bool
	Related                 bool
	Lenses                  bool
	Path                    bool
	Requisites              bool
	SubmittedTo             bool
	Answers                 bool
	UserMarks               bool // marks owned by the logged in user
	UnresolvedMarks         bool // all unresolved marks
	AllMarks                bool // just load all marks
	TrustMap                bool // trust map for this user

	// Load options for basic pages
	Edit                  bool // because otherwise a non-published page id will be deleted from the pageMap
	Links                 bool
	DomainsAndPermissions bool

	// Load options for loading a page for editing
	ChangeLogs    bool
	SearchStrings bool
	LensParentID  bool

	// Options for what data to load for the page itself
	HasDraft        bool
	Likes           bool
	IndividualLikes bool // load each user who liked
	Votes           bool
	LastVisit       bool
	IsSubscribed    bool
	SubscriberCount bool
	LinkedMarkCount bool
	RedLinkCount    bool
	Mastery         bool
	UsedAsMastery   bool
	Creators        bool
	EditHistory     bool
	ProposalEditNum bool

	// Options for what fields to load from pages table
	Text      bool
	Summaries bool

	// Options for what data to load after the page's data has been loaded
	NextPrevIDs bool
	PageObjects bool

	// Load the page even if it's deleted
	IncludeDeleted bool
}

// Here we define some commonly used loadOptions templates.
var (
	// Options for loading the primary page
	PrimaryPageLoadOptions = (&PageLoadOptions{
		Questions:       true,
		Children:        true,
		Parents:         true,
		Tags:            true,
		Related:         true,
		ChangeLogs:      true,
		Lenses:          true,
		Path:            true,
		Requisites:      true,
		SubmittedTo:     true,
		UserMarks:       true,
		UnresolvedMarks: true,
		TrustMap:        true,
		Answers:         true,
		LinkedMarkCount: true,
		Mastery:         true,
		UsedAsMastery:   true,
		Creators:        true,
		ProposalEditNum: true,
		NextPrevIDs:     true,
	}).Add(SubpageLoadOptions)
	// Options for full page edit
	PrimaryEditLoadOptions = (&PageLoadOptions{
		Children:              true,
		Parents:               true,
		Tags:                  true,
		Lenses:                true,
		Path:                  true,
		Requisites:            true,
		Answers:               true,
		DomainsAndPermissions: true,
		ChangeLogs:            true,
		SearchStrings:         true,
		LensParentID:          true,
		Links:                 true,
		LinkedMarkCount:       true,
		ProposalEditNum:       true,
		Text:                  true,
		IsSubscribed:          true,
	}).Add(EmptyLoadOptions)
	// Options for loading a full lens
	LensFullLoadOptions = (&PageLoadOptions{
		Questions:       true,
		Children:        true,
		Tags:            true,
		Path:            true,
		SubpageCounts:   true,
		Requisites:      true,
		SubmittedTo:     true,
		UserMarks:       true,
		UnresolvedMarks: true,
		Mastery:         true,
		Creators:        true,
		UsedAsMastery:   true,
		ProposalEditNum: true,
	}).Add(SubpageLoadOptions)
	// Options for loading a subpage (like a comment)
	SubpageLoadOptions = (&PageLoadOptions{
		Comments:              true,
		DomainsAndPermissions: true,
		Links:           true,
		HasDraft:        true,
		IndividualLikes: true,
		Votes:           true,
		SubscriberCount: true,
		IsSubscribed:    true,
		Text:            true,
		PageObjects:     true,
	}).Add(TitlePlusLoadOptions)
	// Options for loading info for an intrasite link popover
	IntrasitePopoverLoadOptions = (&PageLoadOptions{
		Links:         true,
		Votes:         true,
		IsSubscribed:  true,
		SubpageCounts: true,
		AnswerCounts:  true,
		Summaries:     true,
	}).Add(TitlePlusLoadOptions)
	// Options for loading info about a lens
	LensInfoLoadOptions = (&PageLoadOptions{}).Add(TitlePlusLoadOptions)
	// Options for loading an answer
	AnswerLoadOptions = (&PageLoadOptions{
		SubpageCounts: true,
	}).Add(TitlePlusLoadOptions)
	// Options for loading the title of a possibly-deleted page
	TitlePlusIncludeDeletedLoadOptions = (&PageLoadOptions{
		IncludeDeleted: true,
	}).Add(TitlePlusLoadOptions)
	// Options for loading a page to display the title + some additional info.
	TitlePlusLoadOptions = &PageLoadOptions{
		Likes:     true,
		LastVisit: true,
	}
	EmptyLoadOptions = &PageLoadOptions{}
)

// Add creates a union of the existing load options with the given ones.
func (o *PageLoadOptions) Add(with *PageLoadOptions) *PageLoadOptions {
	if with == nil {
		return o
	}
	oVal := reflect.ValueOf(o).Elem()
	withVal := reflect.ValueOf(with).Elem()
	for i := 0; i < oVal.NumField(); i++ {
		oField := oVal.Field(i)
		oField.SetBool(oField.Bool() || withVal.Field(i).Bool())
	}
	return o
}

// filterPageMap filters the given page map based on the given predicate.
func filterPageMap(pageMap map[string]*Page, filter func(*Page) bool) map[string]*Page {
	filteredMap := make(map[string]*Page)
	for id, p := range pageMap {
		if filter(p) {
			filteredMap[id] = p
		}
	}
	return filteredMap
}
