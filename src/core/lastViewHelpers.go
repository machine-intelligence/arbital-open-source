// Contains helpers for lastViews.
package core

import (
	"zanaduu3/src/database"
)

const (
	// Possible views
	LastAchievementsModeView = "lastAchievementsModeView"
	LastBellUpdatesView      = "lastBellUpdatesView"
	LastDiscussionModeView   = "lastDiscussionModeView"
	LastMaintenanceModeView  = "lastMaintenanceModeView"
	LastReadModeView         = "lastReadModeView"
	LastRecentChangesView    = "lastRecentChangesView"
)

// Just load the last time the user loaded the given view.
func LoadLastView(db *database.DB, u *CurrentUser, viewName string) (string, error) {
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

	return lastView, nil
}

// Both load and update the last time the user loaded the given view.
func LoadAndUpdateLastView(db *database.DB, u *CurrentUser, viewName string) (string, error) {
	lastView, err := LoadLastView(db, u, viewName)
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
