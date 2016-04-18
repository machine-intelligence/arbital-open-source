package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// Create a statement for inserting new userTrustSnapshots rows for this user. Returns the
// id of the snapshot created.
func InsertUserTrustSnapshots(tx *database.Tx, u *core.CurrentUser, pageId string) (int64, error) {
	now := database.Now()

	var domainIds []string
	var err error

	if pageId == "" {
		domainIds, err = core.LoadAllDomainIds(tx.DB)
	} else {
		domainIds, err = core.LoadDomainsForPage(tx.DB, pageId)
	}
	if err != nil {
		return 0, err
	}

	// Compute the next snapshot id we will use
	var snapshotId int64
	row := tx.DB.NewStatement(`
		SELECT IFNULL(max(id),0)
		FROM userTrustSnapshots
		`).WithTx(tx).QueryRow()
	_, err = row.Scan(&snapshotId)
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
		hashmap["createdAt"] = now
		hashmaps = append(hashmaps, hashmap)
	}
	statement := tx.DB.NewMultipleInsertStatement("userTrustSnapshots", hashmaps)
	if _, err := statement.WithTx(tx).Exec(); err != nil {
		return 0, err
	}

	return snapshotId, err
}
