// page.go contains all the page stuff
package core

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

const (
	// Helpers for matching our markdown extensions
	SpacePrefix   = "(^| |\n)"
	NoParenSuffix = "($|[^(])"
)

// NewPage returns a pointer to a new page object created with the given page id
func NewPage(pageId string) *Page {
	p := &Page{corePageData: corePageData{PageId: pageId}}
	p.Votes = make([]*Vote, 0)
	p.Summaries = make(map[string]string)
	p.CreatorIds = make([]string, 0)
	p.AnswerIds = make([]string, 0)
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
	p.Members = make(map[string]*Member)
	return p
}

// AddPageIdToMap adds a new page with the given page id to the map if it's not
// in the map already.
// Returns the new/existing page.
func AddPageIdToMap(pageId string, pageMap map[string]*Page) *Page {
	if !IsIdValid(pageId) {
		return nil
	}
	if p, ok := pageMap[pageId]; ok {
		return p
	}
	p := NewPage(pageId)
	pageMap[pageId] = p
	return AddPageToMap(pageId, pageMap, EmptyLoadOptions)
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
			lowerCaseString := strings.ToLower(submatch[3][:1]) + submatch[3][1:]
			upperCaseString := strings.ToUpper(submatch[3][:1]) + submatch[3][1:]
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
		regexp.MustCompile(SpacePrefix + "(" + regexp.QuoteMeta(sessions.GetDomain()) + "/p(?:ages)?/)(" + AliasRegexpStr + ")"),
		// Find ids and aliases using [alias optional text] syntax.
		regexp.MustCompile(SpacePrefix + "(\\[\\-?)(" + AliasRegexpStr + ")( [^\\]]*?)?(\\])([^(]|$)"),
		// Find ids and aliases using [text](alias) syntax.
		regexp.MustCompile(SpacePrefix + "(\\[[^\\]]+?\\]\\()(" + AliasRegexpStr + ")(\\))"),
		// Find ids and aliases using [vote: alias] syntax.
		regexp.MustCompile(SpacePrefix + "(\\[vote: ?)(" + AliasRegexpStr + ")(\\])"),
		// Find ids and aliases using [@alias] syntax.
		regexp.MustCompile(SpacePrefix + "(\\[@)(" + AliasRegexpStr + ")(\\])([^(]|$)"),
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
		FROM pageInfos
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

		lowerCaseString := strings.ToLower(submatch[3][:1]) + submatch[3][1:]
		upperCaseString := strings.ToUpper(submatch[3][:1]) + submatch[3][1:]

		// Since ReplaceAllStringFunc gives us the whole match, rather than submatch
		// array, we have stored it earlier and can now piece it together
		if id, ok := aliasMap[lowerCaseString]; ok {
			return submatch[1] + submatch[2] + id + strings.Join(submatch[4:], "")
		}
		if id, ok := aliasMap[upperCaseString]; ok {
			return submatch[1] + submatch[2] + id + strings.Join(submatch[4:], "")
		}
		return match
	}
	for _, exp := range regexps {
		text = exp.ReplaceAllStringFunc(text, replaceAlias)
	}
	return text, nil
}

// UpdatePageLinks updates the links table for the given page by parsing the text.
func UpdatePageLinks(tx *database.Tx, pageId string, text string, configAddress string) error {
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
	extractLinks(regexp.MustCompile(regexp.QuoteMeta(configAddress) + "/p(?:ages)?/(" + AliasRegexpStr + ")"))
	// Find ids and aliases using [alias optional text] syntax.
	extractLinks(regexp.MustCompile("\\[\\-?(" + AliasRegexpStr + ")(?: [^\\]]*?)?\\](?:[^(]|$)"))
	// Find ids and aliases using [text](alias) syntax.
	extractLinks(regexp.MustCompile("\\[.+?\\]\\((" + AliasRegexpStr + ")\\)"))
	// Find ids and aliases using [vote: alias] syntax.
	extractLinks(regexp.MustCompile("\\[vote: ?(" + AliasRegexpStr + ")\\]"))
	// Find ids and aliases using [@alias] syntax.
	extractLinks(regexp.MustCompile("\\[@?(" + AliasRegexpStr + ")\\]"))
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
	// Match [todo: text]
	re := regexp.MustCompile("\\[todo: ?[^\\]]*\\]")
	submatches := re.FindAllString(text, -1)
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
	return fmt.Sprintf("/e/%s", pageId)
}

// GetEditPageFullUrl returns the full url for editing the given page.
func GetEditPageFullUrl(subdomain string, pageId string) string {
	if len(subdomain) > 0 {
		subdomain += "."
	}
	domain := strings.TrimPrefix(sessions.GetRawDomain(), "http://")
	return fmt.Sprintf("http://%s%s/e/%s", subdomain, domain, pageId)
}

// GetNewPageUrl returns the domain relative url for creating a page with a set alias.
func GetNewPageUrl(alias string) string {
	if alias != "" {
		alias = fmt.Sprintf("?alias=%s", alias)
	}
	return fmt.Sprintf("/e/%s", alias)
}

// GetEditLevel checks if the user can edit this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "comment" = can't perform action because this is a comment page the user doesn't own
// "###" = user doesn't have at least ### karma
func GetEditLevel(p *Page, u *user.User) string {
	karmaReq := p.EditKarmaLock
	if karmaReq < EditPageKarmaReq && p.WasPublished {
		karmaReq = EditPageKarmaReq
	}
	if u.Karma < karmaReq {
		if u.IsAdmin {
			return "admin"
		}
		return fmt.Sprintf("%d", karmaReq)
	}
	return ""
}

// GetDeleteLevel checks if the user can delete this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "###" = user doesn't have at least ### karma
func GetDeleteLevel(p *Page, u *user.User) string {
	karmaReq := p.EditKarmaLock
	if karmaReq < DeletePageKarmaReq {
		karmaReq = DeletePageKarmaReq
	}
	if u.Karma < karmaReq {
		if u.IsAdmin {
			return "admin"
		}
		return fmt.Sprintf("%d", karmaReq)
	}
	return ""
}

// CorrectPageType converts the page type to lowercase and checks that it's
// an actual page type we support.
func CorrectPageType(pageType string) (string, error) {
	pageType = strings.ToLower(pageType)
	if pageType != WikiPageType &&
		pageType != LensPageType &&
		pageType != QuestionPageType &&
		pageType != AnswerPageType &&
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
