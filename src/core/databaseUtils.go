// databaseUtils.go contains various helpers for dealing with database and tables
package core

import (
	"fmt"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// PageInfosTable is a wrapper for loading data from the pageInfos table.
// It filters all the pages to make sure the current can actually access them.
// It also filters out any pages that are deleted or aren't published.
func PageInfosTable(u *CurrentUser) *database.QueryPart {
	if u == nil {
		return database.NewQuery(`SELECT * FROM pageInfos WHERE currentEdit>0 AND NOT isDeleted`)
	}
	allowedGroups := append(u.GroupIds, "")
	return database.NewQuery(`(
		SELECT *
		FROM pageInfos
		WHERE seeGroupId IN`).AddArgsGroupStr(allowedGroups).Add(`
			AND currentEdit>0 AND NOT isDeleted
	)`)
}

// Like PageInfosTable but allows for autosaves, snapshots, and deleted pages.
func PageInfosTableAll(u *CurrentUser) *database.QueryPart {
	if u == nil {
		return database.NewQuery(`pageInfos`)
	}
	allowedGroups := append(u.GroupIds, "")
	return database.NewQuery(`(
		SELECT *
		FROM pageInfos
		WHERE seeGroupId IN`).AddArgsGroupStr(allowedGroups).Add(`
	)`)
}

// Replace a rune at a specific index in a string
func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

// Get the next highest base36 character, without vowels
// Returns the character, and true if it wrapped around to 0
// Since we decided that ids must begin with a digit, only allow characters 0-9 for the first character index
func GetNextBase31Char(c sessions.Context, char rune, isFirstChar bool) (rune, bool, error) {
	validChars := Base31Chars
	if isFirstChar {
		validChars = Base31CharsForFirstChar
	}
	index := strings.Index(validChars, strings.ToLower(string(char)))
	if index < 0 {
		return '0', false, fmt.Errorf("invalid character")
	}
	if index < len(validChars)-1 {
		nextChar := rune(validChars[index+1])
		return nextChar, false, nil
	} else {
		nextChar := rune(validChars[0])
		return nextChar, true, nil
	}
}

// Increment a base31 Id string
func IncrementBase31Id(c sessions.Context, previousId string) (string, error) {
	// Add 1 to the base36 value, skipping vowels
	// Start at the last character in the Id string, carrying the 1 as many times as necessary
	nextAvailableId := previousId
	index := len(nextAvailableId) - 1
	var newChar rune
	var err error
	processNextChar := true
	for processNextChar {
		// If we need to carry the 1 all the way to the beginning, then add a 1 at the beginning of the string
		if index < 0 {
			nextAvailableId = "1" + nextAvailableId
			processNextChar = false
		} else {
			// Increment the character at the current index in the Id string
			newChar, processNextChar, err = GetNextBase31Char(c, rune(nextAvailableId[index]), index == 0)
			if err != nil {
				return "", fmt.Errorf("Error processing id: %v", err)
			}
			nextAvailableId = replaceAtIndex(nextAvailableId, newChar, index)
			index = index - 1
		}
	}

	return nextAvailableId, nil
}

// Call GetNextAvailableId in a new transaction
func GetNextAvailableIdInNewTransaction(db *database.DB) (string, error) {
	return db.Transaction(func(tx *database.Tx) (string, error) {
		return GetNextAvailableId(tx)
	})
}

// Get the next available base36 Id string that doesn't contain vowels
func GetNextAvailableId(tx *database.Tx) (string, error) {
	// Query for the highest used pageId or userId
	var highestUsedId string
	row := database.NewQuery(`
		SELECT MAX(pageId)
		FROM (
			SELECT pageId
			FROM`).AddPart(PageInfosTableAll(nil)).Add(`AS pi
			UNION
			SELECT id
			FROM users
		) AS combined
		WHERE char_length(pageId) = (
			SELECT MAX(char_length(pageId))
			FROM (
				SELECT pageId
				FROM`).AddPart(PageInfosTableAll(nil)).Add(`AS pi
				UNION
				SELECT id
				FROM users
			) AS combined2
    )
		`).ToTxStatement(tx).QueryRow()
	_, err := row.Scan(&highestUsedId)
	if err != nil {
		return "", fmt.Errorf("Couldn't load id: %v", err)
	}
	return IncrementBase31Id(tx.DB.C, highestUsedId)
}
