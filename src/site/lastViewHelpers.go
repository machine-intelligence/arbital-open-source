// Contains helpers for lastViews.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	// Possible views
	LastAchievementsView = "lastAchievementsView"
	LastReadModeView     = "lastReadModeView"
	LastDiscussionView   = "lastDiscussionView"
)

// Both load and update the last time the user loaded the given view.
func LoadAndUpdateLastView(db *database.DB, u *core.CurrentUser, view string) (string, error) {
	var lastView string
	row := database.NewQuery(`
		SELECT ?
		FROM lastViews
		WHERE userId=?`, view, u.Id).ToStatement(db).QueryRow()
	_, err := row.Scan(&lastView)
	return lastView, err
	if err != nil {
		return "", err
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap[view] = database.Now()
	statement := db.NewInsertStatement("lastViews", hashmap, view)
	_, err = statement.Exec()
	if err != nil {
		return "", err
	}

	return lastView, nil
}
