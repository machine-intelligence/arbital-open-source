package site

import (
	"fmt"
	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

// Create a statement for inserting new userTrustSnapshots rows for this user. Returns the
// id of the snapshot created.
func InsertUserTrustSnapshots(tx *database.Tx, u *user.User, pageId string) (int64, error) {
	now := database.Now()

	var domainIds []string
	var err error

	if pageId == "" {
		domainIds, err = user.LoadAllDomainIds(tx.DB)
	} else {
		domainIds, err = loadDomainsForPage(tx.DB, pageId)
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

// Look up the domains that this page is in
func loadDomainsForPage(db *database.DB, pageId string) ([]string, error) {
	domainIds := make([]string, 0)

	rows := database.NewQuery(`
		SELECT domainId
		FROM pageInfos
		JOIN pageDomainPairs
		ON (pageInfos.pageId=pageDomainPairs.pageId)
		WHERE pageInfos.pageId=?`).ToStatement(db).Query(pageId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainId string
		err := rows.Scan(&domainId)
		if err != nil {
			return fmt.Errorf("failed to scan for a domain: %v", err)
		}
		domainIds = append(domainIds, domainId)
		return nil
	})

	// For pages with no domain, we consider them to be in the "" domain.
	if len(domainIds) == 0 {
		domainIds = append(domainIds, "")
	}

	return domainIds, err
}
