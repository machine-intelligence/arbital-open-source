// pageUtils.go contains various helpers for dealing with pages
package core

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	// Helpers for matching our markdown extensions
	SpacePrefix   = "(^| |\n)"
	NoParenSuffix = "($|[^(])"

	// String that can be used inside a regexp to match an a page alias or id
	AliasRegexpStr          = "[A-Za-z0-9_]+\\.?[A-Za-z0-9_]*"
	SubdomainAliasRegexpStr = "[A-Za-z0-9_]*"
	ReplaceRegexpStr        = "[^A-Za-z0-9_]" // used for replacing non-alias characters
)

// NewPage returns a pointer to a new page object created with the given page id
func NewPage(pageId string) *Page {
	p := &Page{corePageData: corePageData{PageId: pageId}}
	p.Votes = make([]*Vote, 0)
	p.Summaries = make(map[string]string)
	p.CreatorIds = make([]string, 0)
	p.CommentIds = make([]string, 0)
	p.QuestionIds = make([]string, 0)
	p.LensIds = make([]string, 0)
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
	p.IndividualLikes = make([]string, 0)
	p.Members = make(map[string]*Member)
	p.Answers = make([]*Answer, 0)
	p.SearchStrings = make(map[string]string)

	// NOTE: we want permissions to be explicitly null so that if someone refers to them
	// they get an error. The permissions are only set when they are also fully computed.
	p.Permissions = nil
	return p
}

// AddPageToMap adds a new page with the given page id to the map if it's not
// in the map already.
// Returns the new/existing page.
func AddPageToMap(pageId string, pageMap map[string]*Page, loadOptions *PageLoadOptions) *Page {
	if !IsIdValid(pageId) {
		return nil
	}
	if p, ok := pageMap[pageId]; ok {
		p.LoadOptions.Add(loadOptions)
		return p
	}
	p := NewPage(pageId)
	p.LoadOptions = *loadOptions
	pageMap[pageId] = p
	return p
}
func AddPageIdToMap(pageId string, pageMap map[string]*Page) *Page {
	return AddPageToMap(pageId, pageMap, EmptyLoadOptions)
}

// AddUserToMap adds a new user with the given user id to the map if it's not
// in the map already.
// Returns the new/existing user.
func AddUserToMap(userId string, userMap map[string]*User) *User {
	if !IsIdValid(userId) {
		return nil
	}
	if u, ok := userMap[userId]; ok {
		return u
	}
	u := &User{Id: userId}
	userMap[userId] = u
	return u
}

// PageIdsStringFromMap returns a comma separated string of all pageIds in the given map.
func PageIdsStringFromMap(pageMap map[string]*Page) string {
	var buffer bytes.Buffer
	for id, _ := range pageMap {
		buffer.WriteString(fmt.Sprintf("%s,", id))
	}
	str := buffer.String()
	if len(str) >= 1 {
		str = str[0 : len(str)-1]
	}
	return str
}

// PageIdsListFromMap returns a comma separated string of all pageIds in the given map.
func PageIdsListFromMap(pageMap map[string]*Page) []interface{} {
	list := make([]interface{}, 0, len(pageMap))
	for id, _ := range pageMap {
		list = append(list, id)
	}
	return list
}

// StandardizeLinks converts all alias links into pageId links.
func StandardizeLinks(db *database.DB, text string) (string, error) {
	// Populate a list of all the links
	aliasesAndIds := make([]string, 0)
	// Track regexp matches, because ReplaceAllStringFunc doesn't support matching groups
	matches := make(map[string][]string)
	extractLinks := func(exp *regexp.Regexp) {
		submatches := exp.FindAllStringSubmatch(text, -1)
		for _, submatch := range submatches {
			matches[submatch[0]] = submatch
			lowerCaseString := strings.ToLower(submatch[2][:1]) + submatch[2][1:]
			upperCaseString := strings.ToUpper(submatch[2][:1]) + submatch[2][1:]
			aliasesAndIds = append(aliasesAndIds, lowerCaseString)
			aliasesAndIds = append(aliasesAndIds, upperCaseString)
		}
	}

	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	// NOTE: each regexp should have two groups that captures stuff that comes before
	// the alias, and then 0 or more groups that capture everything after
	regexps := []*regexp.Regexp{
		// Find directly encoded urls
		regexp.MustCompile("(/p(?:ages)?/)(" + AliasRegexpStr + ")"),
		// Find ids and aliases using [alias optional text] syntax.
		regexp.MustCompile("(\\[[\\-\\+]?)(" + AliasRegexpStr + ")( [^\\]]*?)?(\\])([^(]|$)"),
		// Find ids and aliases using [text](alias) syntax.
		regexp.MustCompile("(\\[[^\\]]+?\\]\\()(" + AliasRegexpStr + ")(\\))"),
		// Find ids and aliases using [vote: alias] syntax.
		regexp.MustCompile("(\\[vote: ?)(" + AliasRegexpStr + ")(\\])"),
		// Find ids and aliases using [@alias] syntax.
		regexp.MustCompile("(\\[@)(" + AliasRegexpStr + ")(\\])([^(]|$)"),
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
		FROM`).AddPart(PageInfosTable(nil)).Add(`AS pi
		WHERE alias IN`).AddArgsGroupStr(aliasesAndIds).ToStatement(db).Query()
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
	extractLinks(regexp.MustCompile(regexp.QuoteMeta(configAddress) + "/p(?:ages)?/(" + AliasRegexpStr + ")"))
	// Find ids and aliases using [alias optional text] syntax.
	extractLinks(regexp.MustCompile("\\[[\\-\\+]?(" + AliasRegexpStr + ")(?: [^\\]]*?)?\\](?:[^(]|$)"))
	// Find ids and aliases using [text](alias) syntax.
	extractLinks(regexp.MustCompile("\\[.+?\\]\\((" + AliasRegexpStr + ")\\)"))
	// Find ids and aliases using [vote: alias] syntax.
	extractLinks(regexp.MustCompile("\\[vote: ?(" + AliasRegexpStr + ")\\]"))
	// Find ids and aliases using [@alias] syntax.
	extractLinks(regexp.MustCompile("\\[@?(" + AliasRegexpStr + ")\\]"))

	aliasesAndIds := make([]string, 0)
	for alias, _ := range linkMap {
		aliasesAndIds = append(aliasesAndIds, alias)
	}
	return aliasesAndIds
}

// UpdatePageLinks updates the links table for the given page by parsing the text.
func UpdatePageLinks(tx *database.Tx, pageId string, text string, configAddress string) error {
	// Delete old links.
	statement := tx.DB.NewStatement("DELETE FROM links WHERE parentId=?").WithTx(tx)
	_, err := statement.Exec(pageId)
	if err != nil {
		return fmt.Errorf("Couldn't delete old links: %v", err)
	}

	aliasesAndIds := ExtractPageLinks(text, configAddress)
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
func ExtractSummaries(pageId string, text string) (map[string]string, []interface{}) {
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
		summaryValues = append(summaryValues, pageId, name, text)
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

// GetPageUrl returns the domain relative url for accessing the given page.
func GetPageUrl(pageId string) string {
	return fmt.Sprintf("/p/%s", pageId)
}

// GetPageFullUrl returns the full url for accessing the given page.
func GetPageFullUrl(subdomain string, pageId string) string {
	if len(subdomain) > 0 {
		subdomain += "."
	}
	domain := strings.TrimPrefix(sessions.GetRawDomain(), "http://")
	return fmt.Sprintf("http://%s%s/p/%s", subdomain, domain, pageId)
}

// GetEditPageUrl returns the domain relative url for editing the given page.
func GetEditPageUrl(pageId string) string {
	return fmt.Sprintf("/edit/%s", pageId)
}

// GetEditPageFullUrl returns the full url for editing the given page.
func GetEditPageFullUrl(subdomain string, pageId string) string {
	if len(subdomain) > 0 {
		subdomain += "."
	}
	domain := strings.TrimPrefix(sessions.GetRawDomain(), "http://")
	return fmt.Sprintf("http://%s%s/edit/%s", subdomain, domain, pageId)
}

// GetNewPageUrl returns the domain relative url for creating a page with a set alias.
func GetNewPageUrl(alias string) string {
	if alias != "" {
		alias = fmt.Sprintf("?alias=%s", alias)
	}
	return fmt.Sprintf("/edit/%s", alias)
}

// CorrectPageType converts the page type to lowercase and checks that it's
// an actual page type we support.
func CorrectPageType(pageType string) (string, error) {
	pageType = strings.ToLower(pageType)
	if pageType != WikiPageType &&
		pageType != LensPageType &&
		pageType != QuestionPageType &&
		pageType != CommentPageType &&
		pageType != GroupPageType &&
		pageType != DomainPageType {
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

func IsIdValid(pageId string) bool {
	if len(pageId) > 0 && pageId[0] > '0' && pageId[0] <= '9' {
		return true
	}
	return false
}

// Check if the given alias is valid
func IsAliasValid(alias string) bool {
	return regexp.MustCompile("^" + AliasRegexpStr + "$").MatchString(alias)
}

func IsUser(db *database.DB, userId string) bool {
	var userCount int
	row := db.NewStatement(`
		SELECT COUNT(id)
		FROM users
		WHERE id=?`).QueryRow(userId)
	row.Scan(&userCount)
	return userCount > 0
}

func GetCommentParents(db *database.DB, pageId string) (string, string, error) {
	var commentParentId string
	var commentPrimaryPageId string
	rows := database.NewQuery(`
		SELECT pi.pageId,pi.type
		FROM`).AddPart(PageInfosTable(nil)).Add(`AS pi
		JOIN pagePairs AS pp
		ON (pi.pageId=pp.parentId)
		WHERE pp.type=?`, ParentPagePairType).Add(`
			AND pp.childId=?`, pageId).Add(`
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId string
		var pageType string
		err := rows.Scan(&parentId, &pageType)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		if pageType == CommentPageType {
			if IsIdValid(commentParentId) {
				return fmt.Errorf("Can't have more than one comment parent")
			}
			commentParentId = parentId
		} else {
			if IsIdValid(commentPrimaryPageId) {
				return fmt.Errorf("Can't have more than one non-comment parent for a comment")
			}
			commentPrimaryPageId = parentId
		}
		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to process rows: %v", err)
	}
	if !IsIdValid(commentPrimaryPageId) {
		return "", "", fmt.Errorf("Comment pages need at least one normal page parent")
	}

	return commentParentId, commentPrimaryPageId, nil
}

func GetPrimaryParentTitle(db *database.DB, u *CurrentUser, pageId string) (string, error) {
	parentTitle := ""
	found, err := database.NewQuery(`
		SELECT primaryParents.title
		FROM`).AddPart(PageInfosTable(u)).Add(`AS primaryParentInfos
		JOIN pagePairs AS pp
		ON primaryParentInfos.pageId=pp.parentId
		JOIN pages AS primaryParents
		ON primaryParentInfos.pageId=primaryParents.pageId
		WHERE primaryParentInfos.type!=?`, CommentPageType).Add(`
			AND primaryParents.isLiveEdit AND pp.type=?`, ParentPagePairType).Add(`
			AND pp.childId=?`, pageId).ToStatement(db).QueryRow().Scan(&parentTitle)
	if err != nil {
		return "", fmt.Errorf("Couldn't load primary parent", err)
	} else if !found {
		return "", fmt.Errorf("Couldn't find a primary parent")
	} else {
		return parentTitle, nil
	}
}

// Look up the domains that this page is in
func LoadDomainsForPage(db *database.DB, pageId string) ([]string, error) {
	return LoadDomainsForPages(db, pageId)
}

// Look up the domains that these pages are in
func LoadDomainsForPages(db *database.DB, pageIds ...interface{}) ([]string, error) {
	domainIds := make([]string, 0)

	rows := database.NewQuery(`
		SELECT pdp.domainId
		FROM`).AddPart(PageInfosTable(nil)).Add(`AS pi
		JOIN pageDomainPairs AS pdp
		ON (pi.pageId=pdp.pageId)
		WHERE pi.pageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainId string
		err := rows.Scan(&domainId)
		if err != nil {
			return fmt.Errorf("failed to scan for a domain: %v", err)
		}
		domainIds = append(domainIds, domainId)
		return nil
	})

	// For pages with no domain, we consider them to be in the "" domain.
	if len(domainIds) == 0 {
		domainIds = append(domainIds, "")
	}

	return domainIds, err
}

// LoadAllDomainIds loads all the domains that currently exist on Arbital.
// If pageMap is given, it also adds them to the pageMap.
func LoadAllDomainIds(db *database.DB, pageMap map[string]*Page) ([]string, error) {
	domainIds := make([]string, 0)
	rows := database.NewQuery(`
		SELECT DISTINCT domainId
		FROM pageDomainPairs`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainId string
		err := rows.Scan(&domainId)
		if err != nil {
			return fmt.Errorf("failed to scan for a domain: %v", err)
		}
		domainIds = append(domainIds, domainId)
		if pageMap != nil {
			AddPageToMap(domainId, pageMap, TitlePlusLoadOptions)
		}
		return nil
	})
	// We also have a "" domain for pages with no domain.
	domainIds = append(domainIds, "")
	return domainIds, err
}

// Return true iff the string is in the list
func IsStringInList(str string, list []string) bool {
	for _, listId := range list {
		if str == listId {
			return true
		}
	}
	return false
}
