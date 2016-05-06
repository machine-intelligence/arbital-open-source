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
	if !p.WasPublished {
		p.Permissions.Edit.Has = true
		return
	}
	// Compute whether the user can edit via any of the domains
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.Edit.Has {
			p.Permissions.Edit.Has = true
			return
		}
	}
	// Check if the user is editing their own comment
	if !p.Permissions.Edit.Has {
		p.Permissions.Edit.Has = p.Type == CommentPageType && p.CreatorId == u.Id
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
	// Compute whether the user can delete via any of the domains
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.Delete.Has {
			p.Permissions.Delete.Has = true
			return
		}
	}
	// Check if the user is deleting their own comment
	if !p.Permissions.Delete.Has {
		p.Permissions.Delete.Has = p.Type == CommentPageType && p.CreatorId == u.Id
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
	// Compute whether the user can comment via any of the domains
	for _, domainId := range p.DomainIds {
		if u.TrustMap[domainId].Permissions.Comment.Has {
			p.Permissions.Comment.Has = true
			return
		}
	}
	// Check if this page is a comment owned by the user, which means they can reply
	if !p.Permissions.Comment.Has {
		p.Permissions.Comment.Has = p.Type == CommentPageType && p.CreatorId == u.Id
	}
	if !p.Permissions.Comment.Has {
		p.Permissions.Comment.Reason = "Not enough reputation to comment"
	}
}

// ComputePermissions computes all the permissions for the given page.
func (p *Page) ComputePermissions(c sessions.Context, u *CurrentUser) {
	p.Permissions = &Permissions{}
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
// TODO: get rid of this and verify permissions per page, with potential helpers,
// e.g. CanCreatePagePair
func VerifyPermissionsForMap(db *database.DB, pageMap map[string]*Page, u *CurrentUser) (string, error) {
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
func VerifyPermissionsForList(db *database.DB, pageIds []string, u *CurrentUser) (string, error) {
	pageMap := make(map[string]*Page)
	for _, pageId := range pageIds {
		AddPageIdToMap(pageId, pageMap)
	}
	return VerifyPermissionsForMap(db, pageMap, u)
}
