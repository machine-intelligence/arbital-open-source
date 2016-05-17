// Contains helpers for lastViews.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	// Possible views
	LastAchievementsModeView = "lastAchievementsModeView"
	LastReadModeView         = "lastReadModeView"
	LastDiscussionModeView   = "lastDiscussionModeView"
)

// Both load and update the last time the user loaded the given view.
func LoadAndUpdateLastView(db *database.DB, u *core.CurrentUser, viewName string) (string, error) {
	var lastView string
	row := database.NewQuery(`
		SELECT viewedAt
		FROM lastViews
		WHERE userId=?`, u.Id).Add(`
			AND viewName=?`, viewName).ToStatement(db).QueryRow()
	_, err := row.Scan(&lastView)
	if err != nil {
		return "", err
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["viewName"] = viewName
	hashmap["viewedAt"] = database.Now()
	statement := db.NewInsertStatement("lastViews", hashmap, "viewedAt")
	_, err = statement.Exec()
	if err != nil {
		return "", err
	}

	return lastView, nil
}
