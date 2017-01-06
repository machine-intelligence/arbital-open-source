// pageUtils.go contains various helpers for dealing with pages
package core

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/sessions"
)

const (
	// Helpers for matching our markdown extensions
	SpacePrefix   = "(^| |\n)"
	NoParenSuffix = "($|[^(])"

	// Strings that can be used inside a regexp to match on a page alias or id.
	AliasOrPageIDRegexpStr          = "[0-9A-Za-z_]+\\.?[0-9A-Za-z_]*"
	SubdomainAliasOrPageIDRegexpStr = "[0-9A-Za-z_]*"

	// Used for replacing non-alias characters.
	ReplaceRegexpStr = "[^A-Za-z0-9_]"

	Base31Chars             = "0123456789bcdfghjklmnpqrstvwxyz"
	Base31CharsForFirstChar = "0123456789"

	StubPageID                    = "72"
	RequestForEditTagParentPageID = "3zj"
	QualityMetaTagsPageID         = "5dg"
	AClassPageID                  = "4yf"
	BClassPageID                  = "4yd"
	FeaturedClassPageID           = "4yl"
	HubPageID                     = "5ls"
	ConceptPageID                 = "6cc"

	MathDomainID = 1
)

// AddPageToMap adds a new page with the given page id to the map if it's not
// in the map already.
// Returns the new/existing page.
func AddPageToMap(pageID string, pageMap map[string]*Page, loadOptions *PageLoadOptions) *Page {
	if !IsIDValid(pageID) {
		return nil
	}
	if p, ok := pageMap[pageID]; ok {
		p.LoadOptions.Add(loadOptions)
		return p
	}
	p := NewPage(pageID)
	p.LoadOptions = *loadOptions
	pageMap[pageID] = p
	return p
}
func AddPageIDToMap(pageID string, pageMap map[string]*Page) *Page {
	return AddPageToMap(pageID, pageMap, EmptyLoadOptions)
}

// AddUserToMap adds a new user with the given user id to the map if it's not
// in the map already.
// Returns the new/existing user.
func AddUserToMap(userID string, userMap map[string]*User) *User {
	if !IsIDValid(userID) {
		return nil
	}
	if u, ok := userMap[userID]; ok {
		return u
	}
	u := NewUser()
	u.ID = userID
	userMap[userID] = u
	return u
}

// Add a markId to the mark map if it's not there already.
func AddMarkToMap(markID string, markMap map[string]*Mark) *Mark {
	mark, ok := markMap[markID]
	if !ok {
		mark = &Mark{ID: markID}
		markMap[markID] = mark
	}
	return mark
}

// PageIdsListFromMap returns a comma separated string of all pageIds in the given map.
func PageIDsListFromMap(pageMap map[string]*Page) []interface{} {
	list := make([]interface{}, 0, len(pageMap))
	for id := range pageMap {
		list = append(list, id)
	}
	return list
}

// StandardizeLinks converts all alias links into pageId links.
func StandardizeLinks(db *database.DB, text string) (string, error) {
	// Populate a list of all the links
	aliasesAndIDs := make([]string, 0)
	// Track regexp matches, because ReplaceAllStringFunc doesn't support matching groups
	matches := make(map[string][]string)
	extractLinks := func(exp *regexp.Regexp) {
		submatches := exp.FindAllStringSubmatch(text, -1)
		for _, submatch := range submatches {
			matches[submatch[0]] = submatch
			lowerCaseString := strings.ToLower(submatch[2][:1]) + submatch[2][1:]
			upperCaseString := strings.ToUpper(submatch[2][:1]) + submatch[2][1:]
			aliasesAndIDs = append(aliasesAndIDs, lowerCaseString)
			aliasesAndIDs = append(aliasesAndIDs, upperCaseString)
		}
	}

	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	// NOTE: each regexp should have two groups that captures stuff that comes before
	// the alias, and then 0 or more groups that capture everything after
	// NOTE: we have to be careful about capturing too much around the link, because
	// if we do, the second link in `[alias] [alias]` won't get captured. For this reason
	// we check for backtick only right at the end of the expression.
	notInBackticks := "([^`]|$)"
	regexps := []*regexp.Regexp{
		// Find directly encoded urls
		regexp.MustCompile("(/p/)(" + AliasOrPageIDRegexpStr + ")" + notInBackticks),
		// Find ids and aliases using [alias optional text] syntax.
		regexp.MustCompile("(\\[[\\-\\+]?)(" + AliasOrPageIDRegexpStr + ")( [^\\]]*?)?(\\])([^(`]|$)"),
		// Find ids and aliases using [text](alias) syntax.
		regexp.MustCompile("(\\[[^\\]]+?\\]\\()(" + AliasOrPageIDRegexpStr + ")(\\))" + notInBackticks),
		// Find ids and aliases using [vote: alias] syntax.
		regexp.MustCompile("(\\[vote: ?)(" + AliasOrPageIDRegexpStr + ")(\\])" + notInBackticks),
		// Find ids and aliases using [@alias] syntax.
		regexp.MustCompile("(\\[@)(" + AliasOrPageIDRegexpStr + ")(\\])([^(`]|$)"),
	}
	for _, exp := range regexps {
		extractLinks(exp)
	}

	if len(aliasesAndIDs) <= 0 {
		return text, nil
	}

	// Populate alias -> pageId map
	aliasMap := make(map[string]string)
	rows := database.NewQuery(`
		SELECT pageId,alias
		FROM pageInfos AS pi
		WHERE alias IN`).AddArgsGroupStr(aliasesAndIDs).Add(`
			AND`).AddPart(WherePageInfos(nil)).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID, alias string
		err := rows.Scan(&pageID, &alias)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		aliasMap[alias] = pageID
		return nil
	})
	if err != nil {
		return "", err
	}

	// Perform replacement
	replaceAlias := func(match string) string {
		submatch := matches[match]

		lowerCaseString := strings.ToLower(submatch[2][:1]) + submatch[2][1:]
		upperCaseString := strings.ToUpper(submatch[2][:1]) + submatch[2][1:]

		// Since ReplaceAllStringFunc gives us the whole match, rather than submatch
		// array, we have stored it earlier and can now piece it together
		if id, ok := aliasMap[lowerCaseString]; ok {
			return submatch[1] + id + strings.Join(submatch[3:], "")
		}
		if id, ok := aliasMap[upperCaseString]; ok {
			return submatch[1] + id + strings.Join(submatch[3:], "")
		}
		return match
	}
	for _, exp := range regexps {
		text = exp.ReplaceAllStringFunc(text, replaceAlias)
	}
	return text, nil
}

// Extract all links from the given page text.
func ExtractPageLinks(text string, configAddress string) []string {
	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	linkMap := make(map[string]bool)
	extractLinks := func(exp *regexp.Regexp) {
		submatches := exp.FindAllStringSubmatch(text, -1)
		for _, submatch := range submatches {
			linkMap[submatch[1]] = true
		}
	}
	// Find directly encoded urls
	extractLinks(regexp.MustCompile(regexp.QuoteMeta(configAddress) + "/p(?:ages)?/(" + AliasOrPageIDRegexpStr + ")"))
	// Find ids and aliases using [alias optional text] syntax.
	extractLinks(regexp.MustCompile("\\[[\\-\\+]?(" + AliasOrPageIDRegexpStr + ")(?: [^\\]]*?)?\\](?:[^(]|$)"))
	// Find ids and aliases using [text](alias) syntax.
	extractLinks(regexp.MustCompile("\\[.+?\\]\\((" + AliasOrPageIDRegexpStr + ")\\)"))
	// Find ids and aliases using [vote: alias] syntax.
	extractLinks(regexp.MustCompile("\\[vote: ?(" + AliasOrPageIDRegexpStr + ")\\]"))
	// Find ids and aliases using [@alias] syntax.
	extractLinks(regexp.MustCompile("\\[@?(" + AliasOrPageIDRegexpStr + ")\\]"))

	aliasesAndIDs := make([]string, 0)
	for alias := range linkMap {
		aliasesAndIDs = append(aliasesAndIDs, alias)
	}
	return aliasesAndIDs
}

// UpdatePageLinks updates the links table for the given page by parsing the text.
func UpdatePageLinks(tx *database.Tx, pageID string, text string, configAddress string) error {
	// Delete old links.
	statement := tx.DB.NewStatement("DELETE FROM links WHERE parentId=?").WithTx(tx)
	_, err := statement.Exec(pageID)
	if err != nil {
		return fmt.Errorf("Couldn't delete old links: %v", err)
	}

	aliasesAndIDs := ExtractPageLinks(text, configAddress)
	if len(aliasesAndIDs) > 0 {
		// Populate linkTuples
		linkMap := make(map[string]bool) // track which aliases we already added to the list
		valuesList := make([]interface{}, 0)
		for _, alias := range aliasesAndIDs {
			lowercaseAlias := strings.ToLower(alias)
			if linkMap[lowercaseAlias] {
				continue
			}
			valuesList = append(valuesList, pageID, lowercaseAlias)
			linkMap[lowercaseAlias] = true
		}

		// Insert all the tuples into the links table.
		statement := tx.DB.NewStatement(`
			INSERT INTO links (parentId,childAlias)
			VALUES ` + database.ArgsPlaceholder(len(valuesList), 2)).WithTx(tx)
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
func GetPageQuickLockedUntilTime() string {
	return time.Now().UTC().Add(PageQuickLockDuration * time.Second).Format(database.TimeLayout)
}

// ExtractSummaries extracts the summaries from the given page text.
func ExtractSummaries(pageID string, text string) (map[string]string, []interface{}) {
	const defaultSummary = "Summary"
	re := regexp.MustCompile("(?ms)^\\[summary(\\([^)]+\\))?: ?([\\s\\S]+?)\\] *(\\z|\n\\z|\n\n)")
	summaries := make(map[string]string)

	submatches := re.FindAllStringSubmatch(text, -1)
	for _, submatch := range submatches {
		name := strings.Trim(submatch[1], "()")
		text := submatch[2]
		if name == "" {
			name = defaultSummary
		}
		summaries[name] = strings.TrimSpace(text)
	}
	if _, ok := summaries[defaultSummary]; !ok {
		// If no summaries, just extract the first line.
		re = regexp.MustCompile("^(.*)")
		submatch := re.FindStringSubmatch(text)
		summaries[defaultSummary] = strings.TrimSpace(submatch[1])
	}

	// Compute values for doing INSERT
	summaryValues := make([]interface{}, 0)
	for name, text := range summaries {
		summaryValues = append(summaryValues, pageID, name, text)
	}
	return summaries, summaryValues
}

// ExtractTodoCount extracts the number of todos from a page text.
func ExtractTodoCount(text string) int {
	// Match [todo: text] or |todo: text| or ||todo: text|| (any number of vertical bars)

	// Regexp for todo with brackets, [todo: text]
	re := regexp.MustCompile("\\[todo: ?[^\\]]*\\]")
	submatches := re.FindAllString(text, -1)
	// Regexp for todo with vertical bars, |todo: text|, ||todo: text|| etc.
	re = regexp.MustCompile("\\|+?todo: ?[^\\|]*\\|+")
	submatches = append(submatches, re.FindAllString(text, -1)...)
	todoCount := len(submatches)

	// Match [ red link text]
	re = regexp.MustCompile("\\[ [^\\]]+\\]")
	submatches = re.FindAllString(text, -1)
	return todoCount + len(submatches)
}

// GetPageFullUrl returns the full url for accessing the given page.
func GetPageFullURL(subdomain string, pageID string) string {
	if len(subdomain) > 0 {
		subdomain += "."
	}
	domain := strings.TrimPrefix(sessions.GetRawDomain(), "http://")
	return fmt.Sprintf("http://%s%s/p/%s", subdomain, domain, pageID)
}

// GetEditPageFullUrl returns the full url for editing the given page.
func GetEditPageFullURL(subdomain string, pageID string) string {
	if len(subdomain) > 0 {
		subdomain += "."
	}
	domain := strings.TrimPrefix(sessions.GetRawDomain(), "http://")
	return fmt.Sprintf("http://%s%s/edit/%s", subdomain, domain, pageID)
}

// CorrectPageType converts the page type to lowercase and checks that it's
// an actual page type we support.
func CorrectPageType(pageType string) (string, error) {
	pageType = strings.ToLower(pageType)
	if pageType != WikiPageType &&
		pageType != QuestionPageType &&
		pageType != CommentPageType &&
		pageType != GroupPageType {
		return pageType, fmt.Errorf("Invalid page type: %s", pageType)
	}
	return pageType, nil
}
func CorrectPagePairType(pagePairType string) (string, error) {
	pagePairType = strings.ToLower(pagePairType)
	if pagePairType != ParentPagePairType &&
		pagePairType != TagPagePairType &&
		pagePairType != RequirementPagePairType &&
		pagePairType != SubjectPagePairType {
		return pagePairType, fmt.Errorf("Incorrect type: %s", pagePairType)
	}
	return pagePairType, nil
}

func IsIDValid(pageID string) bool {
	if len(pageID) > 0 && pageID[0] > '0' && pageID[0] <= '9' {
		return true
	}
	return false
}

// We store int64 ids in strings. Because of that invalid ids can have two values: "" and "0".
// Check if the given int64 id is valid.
func IsIntIDValid(id string) bool {
	return id != "" && id != "0"
}

// Check if the given alias is valid
func IsAliasValid(alias string) bool {
	return regexp.MustCompile("^" + AliasOrPageIDRegexpStr + "$").MatchString(alias)
}

func GetCommentParents(db *database.DB, pageID string) (string, string, error) {
	var commentParentID string
	var commentPrimaryPageID string
	rows := database.NewQuery(`
		SELECT pi.pageId,pi.type
		FROM pageInfos AS pi
		JOIN pagePairs AS pp
		ON (pi.pageId=pp.parentId)
		WHERE pp.type=?`, ParentPagePairType).Add(`
			AND pp.childId=?`, pageID).Add(`
			AND`).AddPart(WherePageInfos(nil)).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID string
		var pageType string
		err := rows.Scan(&parentID, &pageType)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		if pageType == CommentPageType {
			if IsIDValid(commentParentID) {
				return fmt.Errorf("Can't have more than one comment parent")
			}
			commentParentID = parentID
		} else {
			if IsIDValid(commentPrimaryPageID) {
				return fmt.Errorf("Can't have more than one non-comment parent for a comment")
			}
			commentPrimaryPageID = parentID
		}
		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to process rows: %v", err)
	} else if !IsIDValid(commentPrimaryPageID) {
		return "", "", fmt.Errorf("Comment pages need at least one normal page parent")
	}

	return commentParentID, commentPrimaryPageID, nil
}

// Return true iff the string is in the list
func IsStringInList(str string, list []string) bool {
	for _, listID := range list {
		if str == listID {
			return true
		}
	}
	return false
}

type CreateNewPageOptions struct {
	// If PageID isn't given, one will be created
	PageID           string
	Alias            string
	Type             string
	EditDomainID     string
	SeeDomainID      string
	SubmitToDomainID string
	Title            string
	Clickbait        string
	Text             string
	ExternalUrl      string
	IsEditorComment  bool
	IsPublished      bool

	// Additional options
	ParentIDs []string
	// If creating a new comment, this is the id of the page to which the comment belongs...
	CommentPrimaryPageID string
	Tx                   *database.Tx
}

func CreateNewPage(db *database.DB, u *CurrentUser, options *CreateNewPageOptions) (string, error) {
	// Error checking
	if options.Alias != "" && !IsAliasValid(options.Alias) {
		return "", fmt.Errorf("Invalid alias")
	}
	if options.IsEditorComment && options.Type != CommentPageType {
		return "", fmt.Errorf("Can't set isEditorComment for non-comment pages")
	}
	if options.Type == CommentPageType && len(options.ParentIDs) <= 0 {
		return "", fmt.Errorf("Comments should have a page parent")
	}

	// For new comments, check if the user has the permissions to create an approved comment
	isApprovedComment := false
	if options.Type == CommentPageType {
		if !IsIDValid(options.CommentPrimaryPageID) {
			return "", fmt.Errorf("Creating a new comment without a primary page parent id")
		}
		domainMap := make(map[string]*Domain)
		p, err := LoadFullEdit(db, options.CommentPrimaryPageID, u, domainMap, nil)
		if err != nil {
			return "", fmt.Errorf("Error while loading full edit", err)
		}
		isApprovedComment = p.Permissions.Comment.Has
	}

	// Check that the external url is unique
	if len(options.ExternalUrl) > 0 {
		isDupe, originalPageID, err := IsDuplicateExternalUrl(db, u, options.ExternalUrl)
		if err != nil {
			return "", fmt.Errorf("Couldn't check if external url is already in use: %v", err)
		}
		if isDupe {
			if len(originalPageID) == 0 {
				return "", fmt.Errorf("This external url is already in use.")
			}
			return "", fmt.Errorf("This external url is already in use. See: %v", GetPageFullURL("", originalPageID))
		}
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		if options.Tx != nil {
			tx = options.Tx
		}

		if options.PageID == "" {
			var err error
			options.PageID, err = GetNextAvailableID(tx)
			if err != nil {
				return sessions.NewError("Couldn't get next available id", err)
			}
		}

		// Fill in the defaults
		if options.Alias == "" {
			options.Alias = options.PageID
		}
		if options.Type == "" {
			options.Type = WikiPageType
		}
		if !IsIntIDValid(options.EditDomainID) {
			options.EditDomainID = u.MyDomainID()
		}

		// Update pageInfos
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = options.PageID
		hashmap["alias"] = options.Alias
		hashmap["type"] = options.Type
		hashmap["maxEdit"] = 1
		hashmap["createdBy"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["seeDomainId"] = options.SeeDomainID
		hashmap["editDomainId"] = options.EditDomainID
		hashmap["submitToDomainId"] = options.SubmitToDomainID
		hashmap["lockedBy"] = u.ID
		hashmap["lockedUntil"] = GetPageQuickLockedUntilTime()
		hashmap["externalUrl"] = options.ExternalUrl
		hashmap["sortChildrenBy"] = LikesChildSortingOption
		hashmap["isEditorComment"] = options.IsEditorComment
		hashmap["isApprovedComment"] = isApprovedComment
		if options.IsPublished {
			hashmap["currentEdit"] = 1
		}
		statement := db.NewInsertStatement("pageInfos", hashmap).WithTx(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos", err)
		}

		// Update pages
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = options.PageID
		hashmap["edit"] = 1
		hashmap["title"] = options.Title
		hashmap["clickbait"] = options.Clickbait
		hashmap["text"] = options.Text
		hashmap["creatorId"] = u.ID
		hashmap["createdAt"] = database.Now()
		if options.IsPublished {
			hashmap["isLiveEdit"] = true
		} else {
			hashmap["isAutosave"] = true
		}
		statement = db.NewInsertStatement("pages", hashmap).WithTx(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pages", err)
		}

		if options.IsPublished {
			// Add a summary for the page
			hashmap = make(database.InsertMap)
			hashmap["pageId"] = options.PageID
			hashmap["name"] = "Summary"
			hashmap["text"] = options.Text
			statement = tx.DB.NewInsertStatement("pageSummaries", hashmap).WithTx(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't create a new page summary", err)
			}
		}

		// Subscribe this user to the page that they just created.
		err2 := AddSubscription(tx, u.ID, options.PageID, true)
		if err2 != nil {
			return err2
		}

		return nil
	})
	if err2 != nil {
		return "", sessions.ToError(err2)
	}

	// Add parents
	for _, parentIDStr := range options.ParentIDs {
		_, err := CreateNewPagePair(db, u, &CreateNewPagePairOptions{
			ParentID: parentIDStr,
			ChildID:  options.PageID,
			Type:     ParentPagePairType,
		})
		if err != nil {
			return "", fmt.Errorf("Couldn't create a new page pair: %v", err)
		}
	}

	// Update elastic search index.
	if options.IsPublished {
		doc := &elastic.Document{
			PageID:    options.PageID,
			Type:      options.Type,
			Title:     options.Title,
			Clickbait: options.Clickbait,
			Text:      options.Text,
			Alias:     options.Alias,
			CreatorID: u.ID,
		}
		err := elastic.AddPageToIndex(db.C, doc)
		if err != nil {
			return "", fmt.Errorf("Failed to update index: %v", err)
		}
	}

	return options.PageID, nil
}

// Create / update a subscription
func AddSubscription(tx *database.Tx, userID string, toPageID string, asMaintainer bool) sessions.Error {
	hashmap := make(database.InsertMap)
	hashmap["userId"] = userID
	hashmap["toId"] = toPageID
	hashmap["createdAt"] = database.Now()
	hashmap["asMaintainer"] = asMaintainer
	statement := tx.DB.NewInsertStatement("subscriptions", hashmap, "asMaintainer").WithTx(tx)
	_, err := statement.Exec()
	if err != nil {
		return sessions.NewError("Couldn't subscribe", err)
	}
	return nil
}

// Return true if the user (can create approved comments, can approve comments)
func CanUserApproveComment(db *database.DB, u *CurrentUser, parentIDs []string) (bool, error) {
	var domainID string
	row := database.NewQuery(`
		SELECT pi.editDomainId
		FROM pageInfos AS pi
		WHERE pi.pageId IN`).AddArgsGroupStr(parentIDs).Add(`
			AND pi.type!=?`, CommentPageType).Add(`
			AND`).AddPart(WherePageInfos(u)).ToStatement(db).QueryRow()
	exists, err := row.Scan(&domainID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}
	if dm, ok := u.DomainMembershipMap[domainID]; !ok {
		return false, nil
	} else {
		return dm.CanApproveComments, nil
	}
}
