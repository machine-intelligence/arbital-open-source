// page.go contains all the page stuff
package core

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

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
	extractLinks(regexp.MustCompile(regexp.QuoteMeta(configAddress) + "/pages/(" + AliasRegexpStr + ")"))
	// Find ids and aliases using [alias optional text] syntax.
	extractLinks(regexp.MustCompile("\\[(" + AliasRegexpStr + ")(?: [^\\]]*?)?\\](?:[^(]|$)"))
	// Find ids and aliases using [text](alias) syntax.
	extractLinks(regexp.MustCompile("\\[.+?\\]\\((" + AliasRegexpStr + ")\\)"))
	// Find ids and aliases using [vote: alias] syntax.
	extractLinks(regexp.MustCompile("\\[vote: ?(" + AliasRegexpStr + ")\\]"))
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

// GetPageUrl returns the domain relative url for accessing the given page.
func GetPageUrl(p *Page) string {
	return fmt.Sprintf("/pages/%s", p.Alias)
}

// GetEditPageUrl returns the domain relative url for editing the given page.
func GetEditPageUrl(p *Page) string {
	return fmt.Sprintf("/edit/%d", p.PageId)
}

// GetEditLevel checks if the user can edit this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "comment" = can't perform action because this is a comment page the user doesn't own
// "###" = user doesn't have at least ### karma
func GetEditLevel(p *Page, u *user.User) string {
	if p.Type == CommentPageType {
		if p.CreatorId == u.Id {
			return ""
		} else {
			return p.Type
		}
	}
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
	if p.Type == CommentPageType {
		if p.CreatorId == u.Id {
			return ""
		} else if u.IsAdmin {
			return "admin"
		} else {
			return p.Type
		}
	}
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
