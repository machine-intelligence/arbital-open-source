// domains.go contains functions for dealing with domains (e.g. propagating domain changes)
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

var (
	NoDomain = &Domain{ID: "0"}
)

// Domain stores all information about Arbital domain (like Math or ValueAlignment)
type Domain struct {
	ID        string `json:"id"`
	PageID    string `json:"pageId"`
	CreatedAt string `json:"createdAt"`
	Alias     string `json:"alias"`

	// Settings
	CanUsersComment      bool `json:"canUsersComment"`
	CanUsersProposeEdits bool `json:"canUsersProposeEdits"`
}

// Information about a member of a domain
type DomainMember struct {
	DomainID     string `json:"domainId"`
	DomainPageID string `json:"domainPageId"`
	UserID       string `json:"userId"`
	CreatedAt    string `json:"createdAt"`
	Role         string `json:"role"`
}

// Returns true if this role is at least as high as the given role.
func (dm *DomainMember) AtLeast(asHighAs DomainRoleType) bool {
	userRole := DomainRoleType(dm.Role)
	return userRole.AtLeast(asHighAs)
}

type ProcessDomainCallback func(db *database.DB, domain *Domain) error

// LoadDomains loads the domains matching the given condition.
func LoadDomains(db *database.DB, queryPart *database.QueryPart, callback ProcessDomainCallback) error {
	rows := database.NewQuery(`
		SELECT d.id,d.pageId,d.createdAt,d.alias,d.canUsersComment,d.canUsersProposeEdits
		FROM domains AS d`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var d Domain
		err := rows.Scan(&d.ID, &d.PageID, &d.CreatedAt, &d.Alias, &d.CanUsersComment, &d.CanUsersProposeEdits)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		return callback(db, &d)
	})
	if err != nil {
		return fmt.Errorf("Couldn't load domains: %v", err)
	}
	return nil
}

// LoadDomain loads the domain info with the given alias.
func LoadDomainByAlias(db *database.DB, domainAlias string) (*Domain, error) {
	var resultDomain *Domain
	queryPart := database.NewQuery(`WHERE d.alias=?`, domainAlias)
	err := LoadDomains(db, queryPart, func(db *database.DB, domain *Domain) error {
		resultDomain = domain
		return nil
	})
	return resultDomain, err
}

// LoadRelevantDomains loads all domains associated with the given pages and users into the domainMap.
func LoadRelevantDomains(db *database.DB, u *CurrentUser, pageMap map[string]*Page, userMap map[string]*User, domainMap map[string]*Domain) error {
	domainIDs := make([]string, 0)
	for domainID := range domainMap {
		domainIDs = append(domainIDs, domainID)
	}
	for domainID := range u.DomainMembershipMap {
		domainIDs = append(domainIDs, domainID)
	}
	whereClause := database.NewQuery(`WHERE FALSE`)
	if len(domainIDs) > 0 {
		whereClause.Add(`OR d.id IN`).AddArgsGroupStr(domainIDs)
	}

	pageIDs := PageIDsListFromMap(pageMap)
	if len(pageIDs) > 0 {
		whereClause.Add(`
			OR d.id IN (
				SELECT seeDomainId FROM pageInfos WHERE pageId IN`).AddArgsGroup(pageIDs).Add(`
				UNION SELECT editDomainId FROM pageInfos WHERE pageId IN`).AddArgsGroup(pageIDs).Add(`
			)`)
	}

	err := LoadDomains(db, whereClause, func(db *database.DB, domain *Domain) error {
		domainMap[domain.ID] = domain
		AddPageIDToMap(domain.PageID, pageMap)
		return nil
	})
	return err
}
