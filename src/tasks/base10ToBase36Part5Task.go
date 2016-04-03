// base10ToBase36Part5Task.go does part 3 of converting all the ids from base 10 to base 36
package tasks

import (
	"fmt"
	"regexp"
	"strings"

	//"zanaduu3/src/core"
	"zanaduu3/src/database"
	//"zanaduu3/src/sessions"
)

// Base10ToBase36Part5Task is the object that's put into the daemon queue.
type Base10ToBase36Part5Task struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *Base10ToBase36Part5Task) IsValid() error {
	return nil
}

var updatedPageIds map[string]int
var currentPageId string
var currentEdit string
var alreadyPrintedHeader bool

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *Base10ToBase36Part5Task) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== PART 5 START ====")
	defer c.Debugf("==== PART 5 COMPLETED ====")

	updatedPageIds = make(map[string]int)
	currentPageId = "0"
	currentEdit = "0"
	alreadyPrintedHeader = false

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		// Scan all pages and replace the links
		rows := db.NewStatement(`
					SELECT pageId,edit,text
					FROM pages
					WHERE 1`).Query()

		/*
			// Scan all pages and replace the links
			rows := db.NewStatement(`
						SELECT pageId,edit,text
						FROM pages
						WHERE pageId = "1m6" and isLiveEdit`).Query()
		*/
		if err = rows.Process(updatePageTextBase10ToBase36again); err != nil {
			c.Debugf("ERROR, failed to update page text: %v", err)
			return "", err
		}

		db.C.Debugf("updatedPageIds: %v", updatedPageIds)

		/*
			rows := db.NewStatement(`
					SELECT base36id
					FROM base10tobase36
					WHERE 1`).Query()

			if err = rows.Process(test1); err != nil {
				c.Debugf("ERROR, failed to update page text: %v", err)
				return "", err
			}
		*/
		return "", nil
	})
	if errMessage != "" {
		return 0, err
	}

	return 0, err
}

func test1(db *database.DB, rows *database.Rows) error {
	var pageId string
	if err := rows.Scan(&pageId); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	rows = database.NewQuery(`
		SELECT base36id,base10id
		FROM base10tobase36
		WHERE base10id = `).AddArg(pageId).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, alias string
		err := rows.Scan(&pageId, &alias)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		db.C.Debugf("pageId: %v", pageId)
		db.C.Debugf("alias: %v", alias)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func updatePageTextBase10ToBase36again(db *database.DB, rows *database.Rows) error {
	var pageId, edit string
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	currentPageId = pageId
	currentEdit = edit
	alreadyPrintedHeader = false

	text, err := standardizeLinksFromBase10ToBase36(db, text)
	if err != nil {
		return fmt.Errorf("failed to standardize links: %v", err)
	}

	// Update pages table
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageId
	hashmap["edit"] = edit
	hashmap["text"] = text

	statement := db.NewInsertStatement("pages", hashmap, "text")

	//db.C.Debugf("statement: %v", statement)
	//db.C.Debugf("hashmap: %v", hashmap)

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

// standardizeLinksFromBase10ToBase36 converts all base10 links into base36 links.
func standardizeLinksFromBase10ToBase36(db *database.DB, text string) (string, error) {

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

			//db.C.Debugf("submatch: %v", submatch)
			//db.C.Debugf("aliasesAndIds: %v", aliasesAndIds)
		}
	}

	fourOrMoreDigitsRegexpStr := "[0-9]{4,}"

	// NOTE: these regexps are waaaay too simplistic and don't account for the
	// entire complexity of Markdown, like 4 spaces, backticks, and escaped
	// brackets / parens.
	// NOTE: each regexp should have two groups that captures stuff that comes before
	// the alias, and then 0 or more groups that capture everything after
	regexps := []*regexp.Regexp{
		// Find directly encoded urls
		regexp.MustCompile("(/p/)(" + fourOrMoreDigitsRegexpStr + ")"),
		regexp.MustCompile("(/pages/)(" + fourOrMoreDigitsRegexpStr + ")"),
		regexp.MustCompile("(/path/)(" + fourOrMoreDigitsRegexpStr + ")"),
		regexp.MustCompile("(/domains/)(" + fourOrMoreDigitsRegexpStr + ")"),
		//regexp.MustCompile("(/user/)(" + fourOrMoreDigitsRegexpStr + ")"),
		//regexp.MustCompile("(/groups/)(" + fourOrMoreDigitsRegexpStr + ")"),
		// Find ids and aliases using [alias optional text] syntax.
		regexp.MustCompile("(\\[\\-?)(" + fourOrMoreDigitsRegexpStr + ")( [^\\]]*?)?(\\])([^(]|$)"),
		// Find ids and aliases using [text](alias) syntax.
		regexp.MustCompile("(\\[[^\\]]+?\\]\\()(" + fourOrMoreDigitsRegexpStr + ")(\\))"),
		// Find ids and aliases using [vote: alias] syntax.
		regexp.MustCompile("(\\[vote: ?)(" + fourOrMoreDigitsRegexpStr + ")(\\])"),
		// Find ids and aliases using [@alias] syntax.
		regexp.MustCompile("(\\[@)(" + fourOrMoreDigitsRegexpStr + ")(\\])([^(]|$)"),
	}
	for _, exp := range regexps {
		extractLinks(exp)
	}

	if len(aliasesAndIds) <= 0 {
		return text, nil
	}

	// Populate alias -> pageId map
	aliasMap := make(map[string]string)

	if !alreadyPrintedHeader && len(aliasesAndIds) > 0 {
		db.C.Debugf("*** PAGEID: %v, edit %v ***", currentPageId, currentEdit)
		alreadyPrintedHeader = true
	}

	db.C.Debugf("aliasesAndIds: %v", aliasesAndIds)
	//db.C.Debugf("aliasMap: %v", aliasMap)

	updatedPageIds[currentPageId] = len(aliasesAndIds)

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
		//db.C.Debugf("pageId: %v", pageId)
		//db.C.Debugf("aliasMap: %v", aliasMap)
		return nil
	})
	if err != nil {
		return "", err
	}

	db.C.Debugf("aliasMap: %v", aliasMap)

	// Perform replacement
	replaceAlias := func(match string) string {
		submatch := matches[match]

		//db.C.Debugf("submatch: %v", submatch)
		db.C.Debugf("submatch[0]: %v", submatch[0])
		//db.C.Debugf("submatch[1]: %v", submatch[1])
		//db.C.Debugf("submatch[2]: %v", submatch[2])
		//db.C.Debugf("submatch[3]: %v", submatch[3])
		//db.C.Debugf("submatch[4]: %v", submatch[4])

		lowerCaseString := strings.ToLower(submatch[2][:1]) + submatch[2][1:]
		upperCaseString := strings.ToUpper(submatch[2][:1]) + submatch[2][1:]

		// Since ReplaceAllStringFunc gives us the whole match, rather than submatch
		// array, we have stored it earlier and can now piece it together
		if id, ok := aliasMap[lowerCaseString]; ok {
			returnval := submatch[1] + id + strings.Join(submatch[3:], "")
			db.C.Debugf("returnval: %v", returnval)
			return returnval
		}
		if id, ok := aliasMap[upperCaseString]; ok {
			returnval := submatch[1] + id + strings.Join(submatch[3:], "")
			db.C.Debugf("returnval: %v", returnval)
			return returnval
		}

		db.C.Debugf("match: %v", match)

		return match
	}
	for _, exp := range regexps {
		text = exp.ReplaceAllStringFunc(text, replaceAlias)
		//db.C.Debugf("text: %v", text)
	}

	return text, nil
}
