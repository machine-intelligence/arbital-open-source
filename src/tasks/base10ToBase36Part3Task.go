// base10ToBase36Part3Task.go does part 3 of converting all the ids from base 10 to base 36
package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// Base10ToBase36Part3Task is the object that's put into the daemon queue.
type Base10ToBase36Part3Task struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *Base10ToBase36Part3Task) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *Base10ToBase36Part3Task) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== PART 3 START ====")
	defer c.Debugf("==== PART 3 COMPLETED ====")

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		// Scan all pages and replace the links
		rows := db.NewStatement(`
				SELECT pageId,edit,text
				FROM pages
				WHERE 1`).Query()

		if err = rows.Process(updatePageTextBase10ToBase36); err != nil {
			c.Debugf("ERROR, failed to update page text: %v", err)
			return "", err
		}

		doOneQuery(db, `ALTER TABLE links DROP INDEX parentId ;`)

		doOneQuery(db, `UPDATE pages SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE pages SET creatorId = CONCAT("zzz", creatorId) WHERE 1;`)

		doOneQuery(db, `UPDATE changeLogs SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE changeLogs SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE changeLogs SET auxPageId = CONCAT("zzz", auxPageId) WHERE 1;`)

		doOneQuery(db, `UPDATE groupMembers SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE groupMembers SET groupId = CONCAT("zzz", groupId) WHERE 1;`)

		doOneQuery(db, `UPDATE likes SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE likes SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE links SET parentId = CONCAT("zzz", parentId) WHERE 1;`)
		//doOneQuery(db, `UPDATE links SET childAlias = CONCAT("zzz", childAlias) WHERE 1;`)

		doOneQuery(db, `UPDATE pageDomainPairs SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageDomainPairs SET domainId = CONCAT("zzz", domainId) WHERE 1;`)

		doOneQuery(db, `UPDATE pageInfos SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET lockedBy = CONCAT("zzz", lockedBy) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET seeGroupId = CONCAT("zzz", seeGroupId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET editGroupId = CONCAT("zzz", editGroupId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET createdBy = CONCAT("zzz", createdBy) WHERE 1;`)
		//doOneQuery(db, `UPDATE pageInfos SET alias = CONCAT("zzz", alias) WHERE 1;`)

		doOneQuery(db, `UPDATE pagePairs SET parentId = CONCAT("zzz", parentId) WHERE 1;`)
		doOneQuery(db, `UPDATE pagePairs SET childId = CONCAT("zzz", childId) WHERE 1;`)

		doOneQuery(db, `UPDATE pageSummaries SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE subscriptions SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE subscriptions SET toId = CONCAT("zzz", toId) WHERE 1;`)

		doOneQuery(db, `UPDATE updates SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByPageId = CONCAT("zzz", groupByPageId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByUserId = CONCAT("zzz", groupByUserId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET subscribedToId = CONCAT("zzz", subscribedToId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET goToPageId = CONCAT("zzz", goToPageId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET byUserId = CONCAT("zzz", byUserId) WHERE 1;`)

		doOneQuery(db, `UPDATE userMasteryPairs SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE userMasteryPairs SET masteryId = CONCAT("zzz", masteryId) WHERE 1;`)

		doOneQuery(db, `UPDATE users SET id = CONCAT("zzz", id) WHERE 1;`)

		doOneQuery(db, `UPDATE visits SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE visits SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE votes SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE votes SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE pages SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pages SET creatorId = creatorIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE changeLogs SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE changeLogs SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE changeLogs SET auxPageId = auxPageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE groupMembers SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE groupMembers SET groupId = groupIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE likes SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE likes SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE links SET parentId = parentIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE links SET childAlias = childAliasBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pageDomainPairs SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageDomainPairs SET domainId = domainIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pageInfos SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET lockedBy = lockedByBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET seeGroupId = seeGroupIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET editGroupId = editGroupIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET createdBy = createdByBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET alias = aliasBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pagePairs SET parentId = parentIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pagePairs SET childId = childIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pageSummaries SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE subscriptions SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE subscriptions SET toId = toIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE updates SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByPageId = groupByPageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByUserId = groupByUserIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET subscribedToId = subscribedToIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET goToPageId = goToPageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET byUserId = byUserIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE userMasteryPairs SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE userMasteryPairs SET masteryId = masteryIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE users SET id = idBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE visits SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE visits SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE votes SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE votes SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `ALTER TABLE links ADD UNIQUE (parentId , childAlias);`)

		return "", nil
	})
	if errMessage != "" {
		return 0, err
	}

	return 0, err
}

func updatePageTextBase10ToBase36(db *database.DB, rows *database.Rows) error {
	var pageId, edit string
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	text, err := StandardizeLinksFromBase10ToBase36(db, text)
	if err != nil {
		return fmt.Errorf("failed to standardize links: %v", err)
	}

	// Update pages table
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageId
	hashmap["edit"] = edit
	hashmap["text"] = text
	statement := db.NewInsertStatement("pages", hashmap, "text")
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update pages table: %v", err)
	}

	/*
		// Update page links table
		err := core.UpdatePageLinks(tx, pageId, text, sessions.GetDomain())
		if err != nil {
			return fmt.Errorf("Couldn't update links: %v", err)
		}
	*/

	return nil
}

// StandardizeLinksFromBase10ToBase36 converts all base10 links into base36 links.
func StandardizeLinksFromBase10ToBase36(db *database.DB, text string) (string, error) {

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
		regexp.MustCompile(core.SpacePrefix + "(" + regexp.QuoteMeta(sessions.GetDomain()) + "/p(?:ages)?/)(" + core.AliasRegexpStr + ")"),
		// Find ids and aliases using [alias optional text] syntax.
		regexp.MustCompile(core.SpacePrefix + "(\\[\\-?)(" + core.AliasRegexpStr + ")( [^\\]]*?)?(\\])([^(]|$)"),
		// Find ids and aliases using [text](alias) syntax.
		regexp.MustCompile(core.SpacePrefix + "(\\[[^\\]]+?\\]\\()(" + core.AliasRegexpStr + ")(\\))"),
		// Find ids and aliases using [vote: alias] syntax.
		regexp.MustCompile(core.SpacePrefix + "(\\[vote: ?)(" + core.AliasRegexpStr + ")(\\])"),
		// Find ids and aliases using [@alias] syntax.
		regexp.MustCompile(core.SpacePrefix + "(\\[@)(" + core.AliasRegexpStr + ")(\\])([^(]|$)"),
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
		SELECT base36id,base10id
		FROM base10tobase36
		WHERE base10id IN`).AddArgsGroupStr(aliasesAndIds).ToStatement(db).Query()
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
