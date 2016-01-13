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
	Answers   bool
	Comments  bool
	Questions bool

	// Load options for titlePlus pages
	Children                bool
	HasGrandChildren        bool
	SubpageCounts           bool
	RedLinkCountForChildren bool
	Parents                 bool
	Tags                    bool
	Related                 bool
	Lenses                  bool
	Requirements            bool
	Subjects                bool

	// Load options for basic pages
	Links   bool
	Domains bool

	// Load options for loading a page for editing
	Edit       bool
	ChangeLogs bool

	// Options for what data to load for the page itself
	ChildDraftId    bool
	HasDraft        bool
	Likes           bool
	Votes           bool
	LastVisit       bool
	IsSubscribed    bool
	SubscriberCount bool
	RedLinkCount    bool
	Mastery         bool
	UsedAsMastery   bool
	Creators        bool

	// Options for what fields to load from pages table
	Text      bool
	Summaries bool

	// Options for what data to load after the page's data has been loaded
	NextPrevIds bool
}

// Here we define some commonly used loadOptions templates.
var (
	// Options for loading the primary page
	PrimaryPageLoadOptions = (&PageLoadOptions{
		Answers:       true,
		Questions:     true,
		Children:      true,
		Parents:       true,
		Tags:          true,
		Related:       true,
		Lenses:        true,
		Requirements:  true,
		Subjects:      true,
		Domains:       true,
		ChildDraftId:  true,
		Mastery:       true,
		UsedAsMastery: true,
		Creators:      true,
		NextPrevIds:   true,
	}).Add(SubpageLoadOptions)
	// Options for full page edit
	PrimaryEditLoadOptions = (&PageLoadOptions{
		Children:     true,
		Parents:      true,
		Tags:         true,
		Lenses:       true,
		Requirements: true,
		Subjects:     true,
		ChangeLogs:   true,
		Links:        true,
		Text:         true,
	}).Add(EmptyLoadOptions)
	// Options for loading a full lens
	LensFullLoadOptions = (&PageLoadOptions{
		Questions:     true,
		SubpageCounts: true,
		Requirements:  true,
		Subjects:      true,
		ChildDraftId:  true,
		Mastery:       true,
		Creators:      true,
		UsedAsMastery: true,
	}).Add(SubpageLoadOptions)
	// Options for loading a subpage (like a comment or answer)
	SubpageLoadOptions = (&PageLoadOptions{
		Comments:        true,
		Links:           true,
		HasDraft:        true,
		Votes:           true,
		SubscriberCount: true,
		IsSubscribed:    true,
		Text:            true,
	}).Add(TitlePlusLoadOptions)
	// Options for loading info for an intrasite link popover
	IntrasitePopoverLoadOptions = (&PageLoadOptions{
		Links:         true,
		Votes:         true,
		IsSubscribed:  true,
		SubpageCounts: true,
		Summaries:     true,
	}).Add(TitlePlusLoadOptions)
	// Options for loading info about a lens
	LensInfoLoadOptions = (&PageLoadOptions{
		Requirements: true,
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
func filterPageMap(pageMap map[int64]*Page, filter func(*Page) bool) map[int64]*Page {
	filteredMap := make(map[int64]*Page)
	for id, p := range pageMap {
		if filter(p) {
			filteredMap[id] = p
		}
	}
	return filteredMap
}
