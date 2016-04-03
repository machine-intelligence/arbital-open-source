// fixTextTask.go updates all pages' text fields to fix common mistakes
package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/database"
	//"zanaduu3/src/user"
)

// FixTextTask is the object that's put into the daemon queue.
type FixTextTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *FixTextTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *FixTextTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== FIX TEXT START ====")
	defer c.Debugf("==== FIX TEXT COMPLETED ====")

	rows := db.NewStatement(`
			SELECT pageId,edit,text
			FROM pages
			WHERE isLiveEdit`).Query()
	//if err = rows.Process(fixText1); err != nil {
	//if err = rows.Process(fixText2); err != nil {
	if err = rows.Process(fixText3); err != nil {
		c.Debugf("ERROR, failed to fix text: %v", err)
		return 0, err
	}

	return 0, err
}

func fixText1(db *database.DB, rows *database.Rows) error {
	var pageId, edit string
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	// Find and replace [token1 token2] with [ token1 token2]
	exp := regexp.MustCompile("(\\[[^ \\\\0-9:-\\]]+ [^\\]]*?\\])(?:[^(]|$)")
	newText := exp.ReplaceAllStringFunc(text, func(submatch string) string {
		parts := strings.Split(submatch, " ")
		parts[0] = "[ " + strings.Split(parts[0], "[")[1]
		return strings.Join(parts, " ")
	})
	if newText != text {
		db.C.Debugf("========================== %s", text)
		db.C.Debugf("========================== %s", newText)
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["edit"] = edit
		hashmap["text"] = newText
		statement := db.NewInsertStatement("pages", hashmap, "text")
		if _, err := statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't update pages table: %v", err)
		}
	}
	return nil
}

func fixText2(db *database.DB, rows *database.Rows) error {
	var pageId, edit string
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	// Find and replace [text](id/alias) links with [id/alias text]

	// First remove all instances of "http://zanaduu3.appspot.com/pages/" in the links, leaving just the pageId
	// On the first pass, accept anything inside the parentheses, since the text we want to remove isn't a valid alias
	exp := regexp.MustCompile("\\[([^\\]]+)\\]\\(([^\\)]+)\\)")
	newText := exp.ReplaceAllStringFunc(text, func(submatch string) string {
		result := submatch
		result = strings.Replace(result, "http://zanaduu3.appspot.com/pages/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/edit/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/pages/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/p/", "", -1)
		db.C.Debugf("submatch: %v", submatch)
		db.C.Debugf("result  : %v", result)
		return result
	})

	// Now convert from [text](id/alias) to [id/alias text]
	// On this pass, only accept valid aliases inside the parentheses, to prevent changing URL links
	exp = regexp.MustCompile("\\[([^\\]]+)\\]\\(([A-Za-z0-9_]+)\\)")
	newText = exp.ReplaceAllString(newText, "[$2 $1]")

	if newText != text {
		db.C.Debugf("========================== %s", text)
		db.C.Debugf("========================== %s", newText)
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["edit"] = edit
		hashmap["text"] = newText
		statement := db.NewInsertStatement("pages", hashmap, "text")
		if _, err := statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't update pages table: %v", err)
		}
	}

	return nil
}

func fixText3(db *database.DB, rows *database.Rows) error {
	var pageId, edit string
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	// Find and replace "Click [here to edit](http://arbital.com/edit/base10id"; links with "Click [here to edit](http://arbital.com/edit/base36id";

	// First remove all instances of "http://zanaduu3.appspot.com/pages/" in the links, leaving just the pageId
	// On the first pass, accept anything inside the parentheses, since the text we want to remove isn't a valid alias
	exp := regexp.MustCompile("Click \\[here to edit\\]\\(http\\:\\/\\/arbital\\.com\\/edit\\/([0-9]+)")

	submatches := exp.FindAllStringSubmatch(text, -1)
	base10Id := "0"
	base36Id := "0"
	for _, submatch := range submatches {
		db.C.Debugf("submatch: %v", submatch)
		db.C.Debugf("submatch[0]: %v", submatch[0])
		db.C.Debugf("submatch[1]: %v", submatch[1])
		//base10Id = submatch[1]

		rows = database.NewQuery(`
				SELECT base36id,base10id
				FROM base10tobase36
				WHERE base10id = `).AddArg(submatch[1]).ToStatement(db).Query()

		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			err := rows.Scan(&base36Id, &base10Id)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			db.C.Debugf("base10Id: %v", base10Id)
			db.C.Debugf("base36Id: %v", base36Id)
			return nil
		})
		if err != nil {
			return err
		}
	}

	newText := exp.ReplaceAllStringFunc(text, func(submatch string) string {
		//exp.ReplaceAllStringFunc(text, func(submatch string) string {

		result := submatch
		result = strings.Replace(result, base10Id, base36Id, -1)
		db.C.Debugf("submatch: %v", submatch)
		db.C.Debugf("result  : %v", result)

		return result
	})

	//db.C.Debugf("newText: %v", newText)

	if newText != text {
		db.C.Debugf("========================== %s", text)
		db.C.Debugf("========================== %s", newText)
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["edit"] = edit
		hashmap["text"] = newText
		statement := db.NewInsertStatement("pages", hashmap, "text")
		if _, err := statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't update pages table: %v", err)
		}
	}

	exp = regexp.MustCompile("Click \\[here to edit\\]\\(http\\:\\/\\/arbital\\.com\\/edit\\/" + pageId)

	submatches = exp.FindAllStringSubmatch(newText, -1)

	for _, submatch := range submatches {
		db.C.Debugf("correct submatch: %v", submatch)
	}

	return nil
}
