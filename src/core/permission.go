// permission.go contains all the stuff related to user permissions
package core

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

type Permissions struct {
	DomainAccess Permission `json:"domainAccess"`
	Edit         Permission `json:"edit"`
	Delete       Permission `json:"delete"`

	// Note that for comments, this means "can reply to this comment"
	// Note that all users can always leave an editor-only comment
	Comment Permission `json:"comment"`
}

// Permission says whether the user can perform a certain action
type Permission struct {
	// We use 'has' AND 'reason' to make sure that if someone doesn't load permissions,
	// the defaults don't allow for anything.
	Has bool `json:"has"`
	// Reason why this action is not allowed.
	Reason string `json:"reason"`
}

func (p *Page) computeDomainPermissions(c sessions.Context, u *CurrentUser) {
	if len(p.DomainIds) <= 0 {
		p.Permissions.DomainAccess.Has = true
		return
	}
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.DomainAccess.Has {
			p.Permissions.DomainAccess.Has = true
		}
	}
	p.Permissions.DomainAccess.Reason = "You are not a member of the domain(s) this page belongs to"
}

func (p *Page) computeEditPermissions(c sessions.Context, u *CurrentUser) {
	// Check the page isn't locked by someone else
	if p.LockedUntil > database.Now() && p.LockedBy != u.Id {
		p.Permissions.Edit.Reason = "Can't change locked page"
		return
	}
	if IsIdValid(p.SeeGroupId) && !u.IsMemberOfGroup(p.SeeGroupId) {
		p.Permissions.Edit.Reason = "You don't have group permission to EVEN SEE this page"
		return
	}
	if IsIdValid(p.EditGroupId) && !u.IsMemberOfGroup(p.EditGroupId) {
		p.Permissions.Edit.Reason = "You don't have group permission to edit this page"
		return
	}
	// If a page hasn't been published, only the creator can edit it
	if !p.WasPublished {
		p.Permissions.Edit.Has = p.PageCreatorId == u.Id
		if !p.Permissions.Edit.Has {
			p.Permissions.Edit.Reason = "Can't edit an unpublished page you didn't create"
		}
		return
	}
	// If it's a comment, only the creator can edit it
	if p.Type == CommentPageType {
		p.Permissions.Edit.Has = p.PageCreatorId == u.Id
		if !p.Permissions.Edit.Has {
			p.Permissions.Edit.Reason = "Can't edit a comment you didn't create"
		}
		return
	}
	// If the page is part of the general domain, only the creator and domain members
	// can edit it.
	if len(p.DomainIds) <= 0 {
		p.Permissions.Edit.Has = p.PageCreatorId == u.Id || u.IsDomainMember
		if !p.Permissions.Edit.Has {
			p.Permissions.Edit.Reason = "Only the creator and domain members can edit an unlisted page"
		}
		return
	}
	// Compute whether the user can edit via any of the domains
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.Edit.Has {
			p.Permissions.Edit.Has = true
			return
		}
	}
	if !p.Permissions.Edit.Has {
		p.Permissions.Edit.Reason = "Not enough reputation to edit this page"
	}
}

func (p *Page) computeDeletePermissions(c sessions.Context, u *CurrentUser) {
	if !p.WasPublished {
		p.Permissions.Delete.Reason = "Can't delete an unpublished page"
		return
	}
	if !p.Permissions.Edit.Has {
		p.Permissions.Delete.Reason = p.Permissions.Edit.Reason
		return
	}
	// If it's a comment, only the creator can delete it
	if p.Type == CommentPageType {
		p.Permissions.Delete.Has = p.PageCreatorId == u.Id || u.IsAdmin
		if !p.Permissions.Delete.Has {
			p.Permissions.Delete.Reason = "Can't delete a comment you didn't create"
		}
		return
	}
	// If the page is part of the general domain, only the creator and domain members
	// can edit it.
	if len(p.DomainIds) <= 0 {
		p.Permissions.Delete.Has = p.PageCreatorId == u.Id || u.IsDomainMember
		if !p.Permissions.Delete.Has {
			p.Permissions.Delete.Reason = "Only the creator and domain members can delete an unlisted page"
		}
		return
	}
	// Compute whether the user can delete via any of the domains
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.Delete.Has {
			p.Permissions.Delete.Has = true
			return
		}
	}
	if !p.Permissions.Delete.Has {
		p.Permissions.Delete.Reason = "Not enough reputation to delete this page"
	}
}

func (p *Page) computeCommentPermissions(c sessions.Context, u *CurrentUser) {
	if !p.WasPublished {
		p.Permissions.Comment.Reason = "Can't comment on an unpublished page"
		return
	}
	// Anyone who can edit the page can also comment
	if p.Permissions.Edit.Has {
		p.Permissions.Comment.Has = true
		return
	}
	// Compute whether the user can comment via any of the domains
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.Comment.Has {
			p.Permissions.Comment.Has = true
			return
		}
	}
	if !p.Permissions.Comment.Has {
		p.Permissions.Comment.Reason = "Not enough reputation to comment"
	}
}

// ComputePermissions computes all the permissions for the given page.
func (p *Page) ComputePermissions(c sessions.Context, u *CurrentUser) {
	p.Permissions = &Permissions{}
	// Order is important
	p.computeDomainPermissions(c, u)
	p.computeEditPermissions(c, u)
	p.computeDeletePermissions(c, u)
	p.computeCommentPermissions(c, u)
}
func ComputePermissionsForMap(c sessions.Context, pageMap map[string]*Page, u *CurrentUser) {
	for _, p := range pageMap {
		p.ComputePermissions(c, u)
	}
}

// Verify that the user has edit permissions for all the pages in the map.
func VerifyEditPermissionsForMap(db *database.DB, pageMap map[string]*Page, u *CurrentUser) (string, error) {
	filteredPageMap := filterPageMap(pageMap, func(p *Page) bool { return len(p.DomainIds) <= 0 })
	err := LoadDomainIds(db, nil, &LoadDataOptions{
		ForPages: filteredPageMap,
	})
	if err != nil {
		return "", fmt.Errorf("Couldn't load domains: %v", err)
	}
	ComputePermissionsForMap(db.C, pageMap, u)
	for _, p := range pageMap {
		if !p.Permissions.Edit.Has {
			return fmt.Sprintf("Don't have edit access to page " + p.PageId + ": " + p.Permissions.Edit.Reason), nil
		}
	}
	return "", nil
}
func VerifyEditPermissionsForList(db *database.DB, pageIds []string, u *CurrentUser) (string, error) {
	pageMap := make(map[string]*Page)
	for _, pageId := range pageIds {
		AddPageIdToMap(pageId, pageMap)
	}
	return VerifyEditPermissionsForMap(db, pageMap, u)
}

// =========================== Relationships ==================================

// Check if the given user can affect a relationship between the two pages.
func CanAffectRelationship(c sessions.Context, parent *Page, child *Page, relationshipType string) (string, error) {
	// No intragroup links allowed.
	if child.SeeGroupId != parent.SeeGroupId {
		return "Parent and child need to have the same See Group", nil
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
	parentOk := parent.Type == DomainPageType || parent.Type == GroupPageType ||
		parent.Type == QuestionPageType || parent.Type == WikiPageType
	childOk := child.Type == LensPageType || child.Type == QuestionPageType || child.Type == WikiPageType
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
	childOk := child.Type == DomainPageType || child.Type == LensPageType || child.Type == WikiPageType
	return childOk
}

// Check if a subject relationship between the two pages would be valid.
func isSubjectRelationshipSupported(parent *Page, child *Page) bool {
	if child.Type == CommentPageType || parent.Type == CommentPageType {
		return false
	}
	parentOk := parent.Type == DomainPageType || parent.Type == LensPageType || parent.Type == WikiPageType
	childOk := child.Type == DomainPageType || child.Type == LensPageType || child.Type == WikiPageType
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
