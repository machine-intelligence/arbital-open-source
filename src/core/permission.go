// permission.go contains all the stuff related to user permissions
package core

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	BannedDomainRole     string = "banned"
	NoDomainRole         string = ""        // aka, not a member
	DefaultDomainRole    string = "default" // aka, is a member, can comment, vote, and propose edits
	TrustedDomainRole    string = "trusted"
	ReviewerDomainRole   string = "reviewer"
	ArbiterDomainRole    string = "arbiter"
	ArbitratorDomainRole string = "arbitrator"
)

// List of roles ordered from least to most priveledged
var _allDomainRoles = []string{
	BannedDomainRole,
	NoDomainRole,
	DefaultDomainRole,
	TrustedDomainRole,
	ReviewerDomainRole,
	ArbiterDomainRole,
	ArbitratorDomainRole,
}

// Returns true if this role is at least as high as the given role. (>=)
func RoleAtLeast(role, atLeast string) bool {
	metThreshold := false
	for _, domainRole := range _allDomainRoles {
		metThreshold = metThreshold || (domainRole == atLeast)
		if domainRole == role {
			return metThreshold
		}
	}
	return false
}

// Return true if the given domain role is a valid one.
func IsDomainRoleValid(role string) bool {
	for _, v := range _allDomainRoles {
		if role == v {
			return true
		}
	}
	return false
}

// Return true iff current user has permission to given the given role to another user
// in the given domain.
func CanCurrentUserGiveRole(u *CurrentUser, domainID string, role string) bool {
	// TODO: check the current role of the user we are changing, since they might have a higher role than us
	currentUserRole := u.GetDomainMembershipRole(domainID)
	if !RoleAtLeast(currentUserRole, ArbiterDomainRole) {
		return false
	}
	// User can give any role up to, but not including Arbiter
	if !RoleAtLeast(role, ArbiterDomainRole) {
		return true
	}
	if !RoleAtLeast(currentUserRole, ArbitratorDomainRole) {
		return false
	}
	// User can give any role up to and including Arbiter
	return role == ArbiterDomainRole
}

type Permissions struct {
	Edit        Permission `json:"edit"`
	ProposeEdit Permission `json:"proposeEdit"`
	Delete      Permission `json:"delete"`

	// Note that for comment a comment page, this means "can reply to this comment"
	Comment        Permission `json:"comment"`
	ProposeComment Permission `json:"proposeComment"`
}

// Permission says whether the user can perform a certain action
type Permission struct {
	// We use 'has' AND 'reason' to make sure that if someone doesn't load permissions,
	// the defaults don't allow for anything.
	Has bool `json:"has"`
	// Reason why this action is not allowed.
	Reason string `json:"reason"`
}

// Return true iff the user has the permission to view the given domain
func CanUserSeeDomain(u *CurrentUser, domainID string) bool {
	return RoleAtLeast(u.GetDomainMembershipRole(domainID), DefaultDomainRole) || u.IsAdmin
}

func (p *Page) computeEditPermissions(c sessions.Context, u *CurrentUser, domainMap map[string]*Domain) {
	// Compute proposeEdit reason after we compute edit permission
	defer func() {
		if p.IsDeleted {
			p.Permissions.ProposeEdit.Has = false
			p.Permissions.ProposeEdit.Reason = "This page is deleted"
			return
		}
		if p.Permissions.Edit.Has {
			p.Permissions.ProposeEdit.Has = true
		}
		p.Permissions.ProposeEdit.Has = domainMap[p.EditDomainID].CanUsersProposeEdits
		if !p.Permissions.ProposeEdit.Has {
			p.Permissions.ProposeEdit.Reason = p.Permissions.Edit.Reason
		}
	}()

	if p.LockedUntil > database.Now() && p.LockedBy != u.ID {
		p.Permissions.Edit.Reason = fmt.Sprintf(
			"Another editor is currently working on the page. It will automatically unlock within half an hour.")
		return
	}

	if IsIntIDValid(p.SeeDomainID) && !CanUserSeeDomain(u, p.SeeDomainID) {
		p.Permissions.Edit.Reason = "You don't have domain permission to EVEN SEE this page"
		return
	}

	if !RoleAtLeast(u.GetDomainMembershipRole(p.EditDomainID), DefaultDomainRole) {
		p.Permissions.Edit.Reason = "You don't have domain permission to edit this page"
		return
	}

	if !RoleAtLeast(u.GetDomainMembershipRole(p.EditDomainID), TrustedDomainRole) {
		p.Permissions.Edit.Reason = "You don't have domain permission to edit this page, but you can propose edits"
		p.Permissions.ProposeEdit.Has = true
		return
	}

	p.Permissions.Edit.Has = true
}

func (p *Page) computeDeletePermissions(c sessions.Context, u *CurrentUser, domainMap map[string]*Domain) {
	if u.IsAdmin {
		p.Permissions.Delete.Has = true
		return
	}
	if !p.WasPublished {
		p.Permissions.Delete.Reason = "Can't delete an unpublished page"
		return
	}
	if !RoleAtLeast(u.GetDomainMembershipRole(p.EditDomainID), ReviewerDomainRole) {
		p.Permissions.Delete.Reason = "You don't have domain permission to delete this page"
		return
	}

	p.Permissions.Delete.Has = true
}

func (p *Page) computeCommentPermissions(c sessions.Context, u *CurrentUser, domainMap map[string]*Domain) {
	if !p.WasPublished {
		p.Permissions.Comment.Reason = "Can't comment on an unpublished page"
		p.Permissions.ProposeComment.Reason = "Can't comment on an unpublished page"
		return
	}

	// Compute proposeComment reason after we compute comment permission
	defer func() {
		if p.Permissions.Comment.Has {
			p.Permissions.ProposeComment.Has = true
			return
		}
		if domainMap[p.EditDomainID].CanUsersProposeComment {
			p.Permissions.ProposeComment.Has = true
			return
		}
		p.Permissions.ProposeComment.Reason = "Sorry, this domain only allows members to comment"
	}()

	// Allowed through the domain directly?
	if RoleAtLeast(u.GetDomainMembershipRole(p.EditDomainID), DefaultDomainRole) {
		p.Permissions.Comment.Has = true
		return
	}
	// Allowed through a friend of the domain?
	for _, friendID := range domainMap[p.EditDomainID].FriendDomainIDs {
		if RoleAtLeast(u.GetDomainMembershipRole(friendID), DefaultDomainRole) {
			p.Permissions.Comment.Has = true
			return
		}
	}
	p.Permissions.Comment.Reason = "You can't comment in this domain because you are not a member"
}

// ComputePermissions computes all the permissions for the given page.
func (p *Page) ComputePermissions(c sessions.Context, u *CurrentUser, domainMap map[string]*Domain) {
	p.Permissions = &Permissions{}
	// Order is important
	p.computeEditPermissions(c, u, domainMap)
	p.computeDeletePermissions(c, u, domainMap)
	p.computeCommentPermissions(c, u, domainMap)
}
func ComputePermissionsForMap(c sessions.Context, u *CurrentUser, pageMap map[string]*Page, domainMap map[string]*Domain) {
	for _, p := range pageMap {
		p.ComputePermissions(c, u, domainMap)
	}
}

// Verify that the user has edit permissions for all the pages in the map.
func VerifyEditPermissionsForMap(db *database.DB, u *CurrentUser, pageMap map[string]*Page) (string, error) {
	err := LoadPages(db, u, pageMap)
	if err != nil {
		return "", fmt.Errorf("Couldn't load pages: %v", err)
	}

	domainIDs := make([]string, 0)
	for _, p := range pageMap {
		domainIDs = append(domainIDs, p.EditDomainID)
	}

	// Load relevant domains
	domainMap := make(map[string]*Domain)
	queryPart := database.NewQuery(`WHERE d.id IN`).AddArgsGroupStr(domainIDs)
	err = LoadDomains(db, queryPart, func(db *database.DB, domain *Domain) error {
		domainMap[domain.ID] = domain
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("Couldn't load domains: %v", err)
	}

	ComputePermissionsForMap(db.C, u, pageMap, domainMap)
	for _, p := range pageMap {
		if !p.Permissions.Edit.Has {
			return fmt.Sprintf("Don't have edit access to page " + p.PageID + ": " + p.Permissions.Edit.Reason), nil
		}
	}
	return "", nil
}
func VerifyEditPermissionsForList(db *database.DB, u *CurrentUser, pageIDs []string) (string, error) {
	pageMap := make(map[string]*Page)
	for _, pageID := range pageIDs {
		AddPageIDToMap(pageID, pageMap)
	}
	return VerifyEditPermissionsForMap(db, u, pageMap)
}

// =========================== Relationships ==================================

// Check if the given user can affect a relationship between the two pages.
func CanAffectRelationship(c sessions.Context, parent *Page, child *Page, relationshipType string) (string, error) {
	// No intra-domain links allowed.
	if child.SeeDomainID != parent.SeeDomainID {
		return "Parent and child need to have the same domain", nil
	}

	// Check if this is a pairing we support
	if relationshipType == ParentPagePairType && !isParentRelationshipSupported(parent, child) {
		return "Parent relationship between these pages is not supported", nil
	} else if relationshipType == TagPagePairType && !isTagRelationshipSupported(parent, child) {
		return "Tag relationship between these pages is not supported", nil
	} else if relationshipType == RequirementPagePairType && !isRequirementRelationshipSupported(parent, child) {
		return "Requirement relationship between these pages is not supported", nil
	} else if relationshipType == SubjectPagePairType && !isSubjectRelationshipSupported(parent, child) {
		return "Subject relationship between these pages is not supported", nil
	}

	// Check if the user has the right permissions
	if relationshipType == ParentPagePairType && !hasParentRelationshipPermissions(parent, child) {
		return "Don't have permission to create a parent relationship between these pages", nil
	} else if relationshipType == TagPagePairType && !hasTagRelationshipPermissions(parent, child) {
		return "Don't have permission to create a tag relationship between these pages", nil
	} else if relationshipType == RequirementPagePairType && !hasRequirementRelationshipPermissions(parent, child) {
		return "Don't have permission to create a requirement relationship between these pages", nil
	} else if relationshipType == SubjectPagePairType && !hasSubjectRelationshipPermissions(parent, child) {
		return "Don't have permission to create a subject relationship between these pages", nil
	}
	return "", nil
}

// Check if a parent relationship between the two pages would be valid.
func isParentRelationshipSupported(parent *Page, child *Page) bool {
	if child.Type == CommentPageType {
		return true
	}
	parentOk := parent.Type == GroupPageType ||
		parent.Type == QuestionPageType || parent.Type == WikiPageType
	childOk := child.Type == QuestionPageType || child.Type == WikiPageType
	return parentOk && childOk
}

// Check if a tag relationship between the two pages would be valid.
func isTagRelationshipSupported(parent *Page, child *Page) bool {
	return child.Type != CommentPageType && parent.Type != CommentPageType
}

// Check if a requirement relationship between the two pages would be valid.
func isRequirementRelationshipSupported(parent *Page, child *Page) bool {
	if child.Type == CommentPageType || parent.Type == CommentPageType {
		return false
	}
	childOk := child.Type == WikiPageType
	return childOk
}

// Check if a subject relationship between the two pages would be valid.
func isSubjectRelationshipSupported(parent *Page, child *Page) bool {
	if child.Type == CommentPageType || parent.Type == CommentPageType {
		return false
	}
	parentOk := parent.Type == WikiPageType
	childOk := child.Type == WikiPageType
	return parentOk && childOk
}

// Check if the current user can create a parent relationship between the two pages.
func hasParentRelationshipPermissions(parent *Page, child *Page) bool {
	if child.Type == CommentPageType {
		return parent.Permissions.Comment.Has || child.IsEditorComment
	}
	return parent.Permissions.Edit.Has && child.Permissions.Edit.Has
}

// Check if the current user can create a tag relationship between the two pages.
func hasTagRelationshipPermissions(parent *Page, child *Page) bool {
	return child.Permissions.Edit.Has
}

// Check if the current user can create a requirement relationship between the two pages.
func hasRequirementRelationshipPermissions(parent *Page, child *Page) bool {
	return child.Permissions.Edit.Has
}

// Check if the current user can create a subject relationship between the two pages.
func hasSubjectRelationshipPermissions(parent *Page, child *Page) bool {
	return child.Permissions.Edit.Has && parent.Permissions.Edit.Has
}
