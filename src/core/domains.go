// domains.go contains functions for dealing with domains (e.g. propagating domain changes)
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

// Domain stores all information about Arbital domain (like Math or ValueAlignment)
type Domain struct {
	ID        string `json:"id"`
	PageID    string `json:"pageId"`
	CreatedAt string `json:"createdAt"`
	Alias     string `json:"alias"`

	// Settings
	CanUsersComment      bool `json:"canUsersComment"` // misleading name: should be CanUsersProposeComments
	CanUsersProposeEdits bool `json:"canUsersProposeEdits"`

	// Additional data loaded from other tables
	FriendDomainIDs []string `json:"friendDomainIds"`
}

func NewDomain() *Domain {
	var d Domain
	d.ID = "0"
	d.FriendDomainIDs = make([]string, 0)
	return &d
}

func NewDomainWithID(id string) *Domain {
	d := NewDomain()
	d.ID = id
	return d
}

// Information about a member of a domain
type DomainMember struct {
	DomainID     string `json:"domainId"`
	DomainPageID string `json:"domainPageId"`
	UserID       string `json:"userId"`
	CreatedAt    string `json:"createdAt"`
	Role         string `json:"role"`

	CanApproveComments bool `json:"canApproveComments"`
	CanSubmitLinks     bool `json:"canSubmitLinks"`
}

// Returns true if this role is at least as high as the given role.
func (dm *DomainMember) AtLeast(asHighAs string) bool {
	return RoleAtLeast(dm.Role, asHighAs)
}

type ProcessDomainCallback func(db *database.DB, domain *Domain) error

// LoadDomains loads the domains matching the given condition.
func LoadDomains(db *database.DB, queryPart *database.QueryPart, callback ProcessDomainCallback) error {
	rows := database.NewQuery(`
		SELECT d.id,d.pageId,d.createdAt,d.alias,d.canUsersComment,d.canUsersProposeEdits
		FROM domains AS d`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		d := NewDomain()
		err := rows.Scan(&d.ID, &d.PageID, &d.CreatedAt, &d.Alias, &d.CanUsersComment, &d.CanUsersProposeEdits)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		return callback(db, d)
	})
	if err != nil {
		return fmt.Errorf("Couldn't load domains: %v", err)
	}
	return nil
}

// LoadDomainByID loads the domain info with the given id.
func LoadDomainByID(db *database.DB, domainID string) (*Domain, error) {
	resultDomain := NewDomain()
	queryPart := database.NewQuery(`WHERE d.id=?`, domainID)
	err := LoadDomains(db, queryPart, func(db *database.DB, domain *Domain) error {
		resultDomain = domain
		return nil
	})
	return resultDomain, err
}

// LoadDomain loads the domain info with the given alias.
func LoadDomainByAlias(db *database.DB, domainAlias string) (*Domain, error) {
	resultDomain := NewDomain()
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

	loadedDomainIDs := make([]string, 0)
	err := LoadDomains(db, whereClause, func(db *database.DB, domain *Domain) error {
		domainMap[domain.ID] = domain
		AddPageIDToMap(domain.PageID, pageMap)
		loadedDomainIDs = append(loadedDomainIDs, domain.ID)
		return nil
	})
	if err != nil {
		return err
	}

	// Load domain friends
	rows := database.NewQuery(`
		SELECT df.domainId,df.friendId
		FROM domainFriends AS df`).Add(`
		WHERE df.domainId IN`).AddArgsGroupStr(loadedDomainIDs).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainID, friendID string
		err := rows.Scan(&domainID, &friendID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		domainMap[domainID].FriendDomainIDs = append(domainMap[domainID].FriendDomainIDs, friendID)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load domains friends: %v", err)
	}
	return nil
}
