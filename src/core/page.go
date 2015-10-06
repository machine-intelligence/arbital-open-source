// page.go contains all the page stuff
package core

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	// Various page types we have in our system.
	WikiPageType     = "wiki"
	CommentPageType  = "comment"
	QuestionPageType = "question"
	AnswerPageType   = "answer"
	LensPageType     = "lens"

	// Various types of updates a user can get.
	TopLevelCommentUpdateType = "topLevelComment"
	ReplyUpdateType           = "reply"
	PageEditUpdateType        = "pageEdit"
	CommentEditUpdateType     = "commentEdit"
	NewPageByUserUpdateType   = "newPageByUser"
	NewChildPageUpdateType    = "newChildPage"

	// Options for sorting page's children.
	ChronologicalChildSortingOption = "chronological"
	AlphabeticalChildSortingOption  = "alphabetical"
	LikesChildSortingOption         = "likes"

	// Options for vote types
	ProbabilityVoteType = "probability"
	ApprovalVoteType    = "approval"

	// Highest karma lock a user can create is equal to their karma * this constant.
	MaxKarmaLockFraction = 0.8

	// When encoding a page id into a compressed string, we use this base.
	PageIdEncodeBase = 36

	// How long the page lock lasts
	PageLockDuration = 30 * 60 // in seconds
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
	KarmaLock      int    `json:"karmaLock"`
	PrivacyKey     int64  `json:"privacyKey,string"`
	GroupId        int64  `json:"groupId,string"`
	ParentsStr     string `json:"parentsStr"`
	DeletedBy      int64  `json:"deletedBy,string"`
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
	//LinkedFrom   []string        `json:"linkedFrom"`
	RedLinkCount int `json:"redLinkCount"`
	// Set to pageId corresponding to the question/answer the user started creating for this page
	ChildDraftId int64 `json:"childDraftId,string"`
	// Page is locked by this user
	LockedBy int64 `json:"lockedBy,string"`
	// User has the page lock until this time
	LockedUntil string `json:"lockedUntil"`

	// === Other data ===
	// This data is included under "Full data", but can also be loaded along side "Auxillary data".

	// Comments.
	CommentIds []string `json:"commentIds"`

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

// PageMetaData contains all the meta date the user types in the meta-text field.
type PageMetaData struct {
	Mastery struct {
		Alias  string
		Levels map[float32]map[string]int
	}
}

// PagePair describes a parent child relationship, which are stored in pagePairs db table.
type PagePair struct {
	Id       int64 `json:"id,string"`
	ParentId int64 `json:"parentId,string"`
	ChildId  int64 `json:"childId,string"`
}

// ProcessParents converts ParentsStr from this page to the Parents array, and
// populates the given pageMap with the parents.
// pageMap can be nil.
func (p *Page) ProcessParents(c sessions.Context, pageMap map[int64]*Page) error {
	if len(p.ParentsStr) <= 0 {
		return nil
	}
	p.Parents = nil
	p.HasParents = false
	parentIds := strings.Split(p.ParentsStr, ",")
	for _, idStr := range parentIds {
		id, err := strconv.ParseInt(idStr, PageIdEncodeBase, 64)
		if err != nil {
			return err
		}
		pair := PagePair{ParentId: id, ChildId: p.PageId}
		if pageMap != nil {
			newPage, ok := pageMap[pair.ParentId]
			if !ok {
				newPage = &Page{PageId: pair.ParentId}
				pageMap[newPage.PageId] = newPage
			}
			newPage.Children = append(newPage.Children, &pair)
		}
		p.Parents = append(p.Parents, &pair)
		p.HasParents = true
	}
	return nil
}

// PageIdsStringFromMap returns a comma separated string of all pageIds in the given map.
func PageIdsStringFromMap(pageMap map[int64]*Page) string {
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

// PageIdsListFromMap returns a comma separated string of all pageIds in the given map.
func PageIdsListFromMap(pageMap map[int64]*Page) []interface{} {
	list := make([]interface{}, 0, len(pageMap))
	for id, _ := range pageMap {
		list = append(list, id)
	}
	return list
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
	textSelect := "\"\" AS text"
	if options.LoadText {
		textSelect = "text"
	}
	summarySelect := "\"\" AS summary"
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
				length(text),metaText,karmaLock,privacyKey,deletedBy,hasVote,voteType,` + summarySelect + `,
				alias,sortChildrenBy,groupId,parents,isAutosave,isSnapshot,isCurrentEdit,isMinorEdit,
				todoCount,anchorContext,anchorText,anchorOffset
			FROM pages
			WHERE`).AddPart(publishedConstraint).Add(`AND deletedBy=0 AND pageId IN`).AddArgsGroup(pageIds).Add(`
			AND (groupId=0 OR groupId IN (SELECT id FROM groups WHERE isVisible) OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=`).AddArg(userId).Add(`))
			ORDER BY edit DESC
		) AS p
		GROUP BY pageId`).ToStatement(db)
	rows := statement.Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p Page
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.PrevEdit, &p.Type, &p.CreatorId, &p.CreatedAt, &p.Title, &p.Clickbait,
			&p.Text, &p.TextLength, &p.MetaText, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.HasVote,
			&p.VoteType, &p.Summary, &p.Alias, &p.SortChildrenBy, &p.GroupId,
			&p.ParentsStr, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
			&p.TodoCount, &p.AnchorContext, &p.AnchorText, &p.AnchorOffset)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}
		if p.DeletedBy <= 0 {
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
			op.IsMinorEdit = p.IsMinorEdit
			op.TodoCount = p.TodoCount
			op.AnchorContext = p.AnchorContext
			op.AnchorText = p.AnchorText
			op.AnchorOffset = p.AnchorOffset
			if err := op.ProcessParents(db.C, nil); err != nil {
				return fmt.Errorf("Couldn't process parents: %v", err)
			}
		}
		return nil
	})
	return err
}

// LoadEditHistory loads the edit history for the given page.
func LoadEditHistory(db *database.DB, page *Page, userId int64) error {
	editHistoryMap := make(map[int]*Page)
	rows := db.NewStatement(`
		SELECT pageId,edit,prevEdit,creatorId,createdAt,deletedBy,isAutosave,
			isSnapshot,isCurrentEdit,isMinorEdit,title,clickbait,length(text)
		FROM pages
		WHERE pageId=? AND (creatorId=? OR NOT isAutosave)
		ORDER BY edit`).Query(page.PageId, userId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p Page
		err := rows.Scan(&p.PageId, &p.Edit, &p.PrevEdit, &p.CreatorId, &p.CreatedAt,
			&p.DeletedBy, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
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
			if edit.DeletedBy <= 0 {
				// Only add non-deleted, public edits or our snapshots to the map
				page.EditHistoryMap[fmt.Sprintf("%d", editNum)] = edit
			}
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

// StandardizeLinks converts all alias links into pageId links.
func StandardizeLinks(db *database.DB, text string) (string, error) {
	// Populate a list of all the links
	aliasesAndIds := make([]interface{}, 0)
	// Track regexp matches, because ReplaceAllStringFunc doesn't support matching groups
	matches := make(map[string][]string)
	extractLinks := func(exp *regexp.Regexp) {
		submatches := exp.FindAllStringSubmatch(text, -1)
		for _, submatch := range submatches {
			matches[submatch[0]] = submatch
			aliasesAndIds = append(aliasesAndIds, submatch[2])
		}
	}

	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	// NOTE: each regexp should have one group that captures stuff that comes before
	// the alias, and then 0 or more groups that capture everything after
	regexps := []*regexp.Regexp{
		// Find directly encoded urls
		regexp.MustCompile("(" + regexp.QuoteMeta(sessions.GetDomain()) + "/pages/)([A-Za-z0-9_-]+)"),
		// Find ids and aliases using [id/alias optional text] syntax.
		regexp.MustCompile("(\\[)([A-Za-z0-9_-]+)( [^\\]]*?)?(\\])([^(]|$)"),
		// Find ids and aliases using [text](id/alias) syntax.
		regexp.MustCompile("(\\[[^\\]]+?\\]\\()([A-Za-z0-9_-]+?)(\\))"),
		// Find ids and aliases using [vote: id/alias] syntax.
		regexp.MustCompile("(\\[vote: ?)([A-Za-z0-9_-]+?)(\\])"),
	}
	for _, exp := range regexps {
		extractLinks(exp)
	}

	if len(aliasesAndIds) <= 0 {
		return text, nil
	}

	// Populate alias -> pageId map
	aliasMap := make(map[string]string)
	rows := database.NewQuery(`
		SELECT pageId,alias
		FROM pages
		WHERE isCurrentEdit AND alias IN`).AddArgsGroup(aliasesAndIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, alias string
		err := rows.Scan(&pageId, &alias)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		aliasMap[alias] = pageId
		return nil
	})
	if err != nil {
		return "", err
	}

	// Perform replacement
	replaceAlias := func(match string) string {
		submatch := matches[match]
		if id, ok := aliasMap[submatch[2]]; ok {
			// Since ReplaceAllStringFunc gives us the whole match, rather than submatch
			// array, we have stored it earlier and can now piece it together
			return submatch[1] + id + strings.Join(submatch[3:], "")
		}
		return match
	}
	for _, exp := range regexps {
		text = exp.ReplaceAllStringFunc(text, replaceAlias)
	}
	return text, nil
}

// UpdatePageLinks updates the links table for the given page by parsing the text.
func UpdatePageLinks(tx *database.Tx, pageId int64, text string, configAddress string) error {
	// Delete old links.
	statement := tx.NewTxStatement("DELETE FROM links WHERE parentId=?")
	_, err := statement.Exec(pageId)
	if err != nil {
		return fmt.Errorf("Couldn't delete old links: %v", err)
	}

	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	aliasesAndIds := make([]string, 0)
	extractLinks := func(exp *regexp.Regexp) {
		submatches := exp.FindAllStringSubmatch(text, -1)
		for _, submatch := range submatches {
			aliasesAndIds = append(aliasesAndIds, submatch[1])
		}
	}
	// Find directly encoded urls
	extractLinks(regexp.MustCompile(regexp.QuoteMeta(configAddress) + "/pages/([0-9]+)"))
	// Find ids and aliases using [id optional text] syntax.
	extractLinks(regexp.MustCompile("\\[([0-9]+)[^\\]]*?\\](?:[^(]|$)"))
	// Find ids and aliases using [text](id) syntax.
	extractLinks(regexp.MustCompile("\\[.+?\\]\\(([0-9]+?)\\)"))
	// Find ids and aliases using [vote: id] syntax.
	extractLinks(regexp.MustCompile("\\[vote: ?([0-9]+?)\\]"))
	if len(aliasesAndIds) > 0 {
		// Populate linkTuples
		linkMap := make(map[string]bool) // track which aliases we already added to the list
		valuesList := make([]interface{}, 0)
		for _, alias := range aliasesAndIds {
			lowercaseAlias := strings.ToLower(alias)
			if linkMap[lowercaseAlias] {
				continue
			}
			valuesList = append(valuesList, pageId, lowercaseAlias)
			linkMap[lowercaseAlias] = true
		}

		// Insert all the tuples into the links table.
		statement := tx.NewTxStatement(`
			INSERT INTO links (parentId,childAlias)
			VALUES ` + database.ArgsPlaceholder(len(valuesList), 2))
		if _, err = statement.Exec(valuesList...); err != nil {
			return fmt.Errorf("Couldn't insert links: %v", err)
		}
	}
	return nil
}

// GetPageLockedUntilTime returns time until the user can have the lock if the locked
// the page right now.
func GetPageLockedUntilTime() string {
	return time.Now().UTC().Add(PageLockDuration * time.Second).Format(database.TimeLayout)
}

// ExtractSummary extracts the summary text from a page text.
func ExtractSummary(text string) string {
	re := regexp.MustCompile("(?ms)^ {0,3}Summary ?: *\n?(.+?)(\n$|\\z)")
	submatches := re.FindStringSubmatch(text)
	if len(submatches) > 0 {
		return strings.TrimSpace(submatches[1])
	}
	// If no summary paragraph, just extract the first line.
	re = regexp.MustCompile("^(.*)")
	submatches = re.FindStringSubmatch(text)
	return strings.TrimSpace(submatches[1])
}

// ExtractTodoCount extracts the number of todos from a page text.
func ExtractTodoCount(text string) int {
	re := regexp.MustCompile("\\[todo: ?[^\\]]*\\]")
	submatches := re.FindAllString(text, -1)
	return len(submatches)
}
