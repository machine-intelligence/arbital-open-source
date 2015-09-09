// page.go contains all the page stuff
package core

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	PageId int64  `json:"pageId,string"`
	Edit   int    `json:"edit"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	// Full text of the page. Not always sent to the FE.
	Text           string `json:"text"`
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
	// Map of page aliases/ids -> page title, so we can expand [alias] links
	Links map[string]string `json:"links"`
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

// LoadPageOptions describes options for loading page(s) from the db
type LoadPageOptions struct {
	LoadText    bool
	LoadSummary bool
	// If set to true, load snapshots and autosaves, not only current edits
	AllowUnpublished bool
}

// LoadPages loads the given pages.
func LoadPages(c sessions.Context, pageMap map[int64]*Page, userId int64, options *LoadPageOptions) error {
	if options == nil {
		options = &LoadPageOptions{}
	}
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := PageIdsStringFromMap(pageMap)
	textSelect := "\"\" AS text"
	if options.LoadText {
		textSelect = "text"
	}
	summarySelect := "\"\" AS summary"
	if options.LoadSummary {
		summarySelect = "summary"
	}
	publishedConstraint := "isCurrentEdit"
	if options.AllowUnpublished {
		publishedConstraint = fmt.Sprintf("(isCurrentEdit || creatorId=%d)", userId)
	}
	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT pageId,edit,type,creatorId,createdAt,title,%s,length(text),karmaLock,privacyKey,
				deletedBy,hasVote,voteType,%s,alias,sortChildrenBy,groupId,parents,
				isAutosave,isSnapshot,isCurrentEdit,todoCount,anchorContext,anchorText,anchorOffset
			FROM pages
			WHERE %s AND deletedBy=0 AND pageId IN (%s) AND
				(groupId=0 OR groupId IN (SELECT id FROM groups WHERE isVisible) OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=%d))
			ORDER BY edit DESC
		) AS p
		GROUP BY pageId`,
		textSelect, summarySelect, publishedConstraint, pageIds, userId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p Page
		err := rows.Scan(
			&p.PageId, &p.Edit, &p.Type, &p.CreatorId, &p.CreatedAt, &p.Title,
			&p.Text, &p.TextLength, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy, &p.HasVote,
			&p.VoteType, &p.Summary, &p.Alias, &p.SortChildrenBy, &p.GroupId,
			&p.ParentsStr, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit,
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
			op.Type = p.Type
			op.CreatorId = p.CreatorId
			op.CreatedAt = p.CreatedAt
			op.Title = p.Title
			op.Text = p.Text
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
			op.TodoCount = p.TodoCount
			op.AnchorContext = p.AnchorContext
			op.AnchorText = p.AnchorText
			op.AnchorOffset = p.AnchorOffset
			if err := op.ProcessParents(c, nil); err != nil {
				return fmt.Errorf("Couldn't process parents: %v", err)
			}
		}
		return nil
	})
	return err
}

func UpdatePageLinks(c sessions.Context, tx *sql.Tx, pageId int64, text string, configAddress string) error {
	// Delete old links.
	query := fmt.Sprintf("DELETE FROM links WHERE parentId=%d", pageId)
	_, err := tx.Exec(query)
	if err != nil {
		return fmt.Errorf("Couldn't delete old links: %v", err)
	}

	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	aliasesAndIds := make([]string, 0, 0)
	extractLinks := func(exp *regexp.Regexp) {
		submatches := exp.FindAllStringSubmatch(text, -1)
		for _, submatch := range submatches {
			aliasesAndIds = append(aliasesAndIds, submatch[1])
		}
	}
	// Find directly encoded urls
	extractLinks(regexp.MustCompile(regexp.QuoteMeta(configAddress) + "/pages/([0-9]+)"))
	// Find ids and aliases using [id/alias] syntax.
	extractLinks(regexp.MustCompile("\\[([A-Za-z0-9_-]+?)\\](?:[^(]|$)"))
	// Find ids and aliases using [text](id/alias) syntax.
	extractLinks(regexp.MustCompile("\\[.+?\\]\\(([A-Za-z0-9_-]+?)\\)"))
	// Find ids and aliases using [vote: id/alias] syntax.
	extractLinks(regexp.MustCompile("\\[vote: ?([A-Za-z0-9_-]+?)\\]"))
	if len(aliasesAndIds) > 0 {
		// Populate linkTuples
		linkMap := make(map[string]bool) // track which aliases we already added to the list
		linkTuples := make([]string, 0, 0)
		for _, alias := range aliasesAndIds {
			if linkMap[alias] {
				continue
			}
			insertValue := fmt.Sprintf("(%d, '%s')", pageId, alias)
			linkTuples = append(linkTuples, insertValue)
			linkMap[alias] = true
		}

		// Insert all the tuples into the links table.
		linkTuplesStr := strings.Join(linkTuples, ",")
		query = fmt.Sprintf(`
			INSERT INTO links (parentId,childAlias)
			VALUES %s`, linkTuplesStr)
		if _, err = tx.Exec(query); err != nil {
			return fmt.Errorf("Couldn't insert links: %v", err)
		}
	}
	return nil
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
