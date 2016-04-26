// Contains helpers for likeable things, like pages and changeLogs.
package site

import (
	"fmt"
	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	// Possible likeable types
	ChangelogLikeableType = "changelog"
	PageLikeableType      = "page"
)

// Get the likeableId of the given likeable. If it doesn't have one, create one for it.
// Returns the likeableId of the likeable.
func GetOrCreateLikeableId(tx *database.Tx, likeableType string, id string) (int64, error) {
	// Figure out which table to look into
	tableName, idField, err := getTableAndIdFieldForLikeable(likeableType)
	if err != nil {
		return 0, err
	}

	// Look up the likeable
	var likeableId int64
	row := tx.DB.NewStatement(`
		SELECT likeableId
		FROM ` + tableName + `
		WHERE ` + idField + `=?`).WithTx(tx).QueryRow(id)
	_, err = row.Scan(&likeableId)
	if err != nil {
		return 0, fmt.Errorf("Couldn't look up likeableId: %v", err)
	}

	// If it already has a likeableId, return that
	if likeableId != 0 {
		return likeableId, nil
	}

	// Otherwise, insert a new likeableId
	result, err := tx.DB.NewInsertStatement("likeableIds", make(database.InsertMap)).WithTx(tx).Exec()
	if err != nil {
		return 0, fmt.Errorf("Couldn't insert new likeableId", err)
	}
	likeableId, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Couldn't retrieve new likeableId", err)
	}

	// Update the likeable with the likeableId
	hashmap := make(database.InsertMap)
	hashmap[idField] = id
	hashmap["likeableId"] = likeableId
	result, err = tx.DB.NewInsertStatement(tableName, hashmap, "likeableId").WithTx(tx).Exec()
	if err != nil {
		return 0, err
	}

	return likeableId, nil
}

// Insert the userTrustSnapshots for the given likeable. Returns the id of the snapshots.
func InsertUserTrustSnapshotsForLikeable(tx *database.Tx, u *core.CurrentUser, likeableType string, id string) (int64, error) {
	if likeableType == PageLikeableType {
		return InsertUserTrustSnapshotsForPage(tx, u, id)
	}
	if likeableType == ChangelogLikeableType {
		return InsertUserTrustSnapshotsForChangelog(tx, u, id)
	}

	return 0, fmt.Errorf("invalid likeableType")
}

// Get the name of the table and id field for the given likeableType.
func getTableAndIdFieldForLikeable(likeableType string) (string, string, error) {
	switch likeableType {
	case PageLikeableType:
		return "pageInfos", "pageId", nil
	case ChangelogLikeableType:
		return "changeLogs", "id", nil
	default:
		return "", "", fmt.Errorf("invalid likeableType")
	}
}
