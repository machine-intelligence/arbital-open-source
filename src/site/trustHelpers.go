package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// Insert userTrustSnapshots rows for this user for this page. Returns the
// id of the snapshot created.
func InsertUserTrustSnapshotsForPage(tx *database.Tx, u *core.CurrentUser, pageId string) (int64, error) {
	domainIds, err := core.LoadDomainsForPage(tx.DB, pageId)
	if err != nil {
		return 0, err
	}

	return insertUserTrustSnapshots(tx, u, domainIds)
}

// Insert userTrustSnapshots rows for this user for this changelog. Returns the
// id of the snapshot created.
func InsertUserTrustSnapshotsForChangelog(tx *database.Tx, u *core.CurrentUser, changelogId string) (int64, error) {
	var pageId string
	var auxPageId string
	row := tx.DB.NewStatement(`
			SELECT pageId,auxPageId
			FROM changelogs
			WHERE id=?`).WithTx(tx).QueryRow(changelogId)
	_, err := row.Scan(&pageId, &auxPageId)
	if err != nil {
		return 0, err
	}

	domainIds, err := core.LoadDomainsForPages(tx.DB, pageId, auxPageId)
	if err != nil {
		return 0, err
	}

	return insertUserTrustSnapshots(tx, u, domainIds)
}

// Insert userTrustSnapshots rows for this user for these domains. Returns the
// id of the snapshot created.
func insertUserTrustSnapshots(tx *database.Tx, u *core.CurrentUser, domainIds []string) (int64, error) {
	// Compute the next snapshot id we will use
	var snapshotId int64
	row := tx.DB.NewStatement(`
		SELECT IFNULL(max(id),0)
		FROM userTrustSnapshots
		`).WithTx(tx).QueryRow()
	_, err := row.Scan(&snapshotId)
	if err != nil {
		return 0, err
	}
	snapshotId++

	// Snapshot user's trust
	hashmaps := make(database.InsertMaps, 0)
	for _, domainId := range domainIds {
		hashmap := make(database.InsertMap)
		hashmap["id"] = snapshotId
		hashmap["userId"] = u.Id
		hashmap["domainId"] = domainId
		hashmap["generalTrust"] = u.TrustMap[domainId].GeneralTrust
		hashmap["editTrust"] = u.TrustMap[domainId].EditTrust
		hashmap["createdAt"] = database.Now()
		hashmaps = append(hashmaps, hashmap)
	}
	statement := tx.DB.NewMultipleInsertStatement("userTrustSnapshots", hashmaps)
	if _, err := statement.WithTx(tx).Exec(); err != nil {
		return 0, err
	}

	return snapshotId, err
}
