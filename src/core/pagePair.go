// pagePair.go contains all the helpers to work with page pairs (aka relationships)
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

const (
	// Various types of page connections.
	ParentPagePairType      = "parent"
	TagPagePairType         = "tag"
	RequirementPagePairType = "requirement"
	SubjectPagePairType     = "subject"
)

type PagePair struct {
	Id            string
	ParentId      string
	ChildId       string
	Type          string
	CreatorId     string
	CreatedAt     string
	EverPublished bool
}

type LoadChildIdsOptions struct {
	// If set, the children will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[string]*Page
	// Type of children to load
	Type string
	// Type of the child relationship to follow
	PagePairType string
	// Load options to set for the new pages
	LoadOptions *PageLoadOptions
}

type LoadParentIdsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[string]*Page
	// Load whether or not each parent has parents of its own.
	LoadHasParents bool
	// Type of the parent relationship to follow
	PagePairType string
	// Load options to set for the new pages
	LoadOptions *PageLoadOptions
	// Mastery map to populate with masteries necessary for a requirement
	MasteryMap map[string]*Mastery
}

type ProcessPagePairCallback func(db *database.DB, pagePair *PagePair) error

// LoadPagePairs loads page pairs matching the given query
func LoadPagePairs(db *database.DB, queryPart *database.QueryPart, callback ProcessPagePairCallback) error {
	rows := database.NewQuery(`
		SELECT id,type,childId,parentId,creatorId,createdAt,everPublished
		FROM pagePairs`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pagePair PagePair
		err := rows.Scan(&pagePair.Id, &pagePair.Type, &pagePair.ChildId, &pagePair.ParentId,
			&pagePair.CreatorId, &pagePair.CreatedAt, &pagePair.EverPublished)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		return callback(db, &pagePair)
	})
	return err
}

// LoadChildIds loads the page ids for all the children of the pages in the given pageMap.
func LoadChildIds(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadChildIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pairTypeFilter := ""
	if options.PagePairType != "" {
		pairTypeFilter = "type = '" + options.PagePairType + "' AND"
	}

	pageTypeFilter := ""
	if options.Type != "" {
		pageTypeFilter = "AND pi.type = '" + options.Type + "'"
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type,pi.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE`).Add(pairTypeFilter).Add(`parentId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON pi.pageId=pp.childId`).Add(pageTypeFilter).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		var ppType string
		var piType string
		err := rows.Scan(&parentId, &childId, &ppType, &piType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(childId, pageMap, options.LoadOptions)

		parent := sourcePageMap[parentId]
		if piType == LensPageType {
			parent.LensIds = append(parent.LensIds, newPage.PageId)
			newPage.ParentIds = append(newPage.ParentIds, parent.PageId)
		} else if piType == CommentPageType {
			parent.CommentIds = append(parent.CommentIds, newPage.PageId)
		} else if piType == QuestionPageType {
			parent.QuestionIds = append(parent.QuestionIds, newPage.PageId)
		} else if piType == WikiPageType && ppType == ParentPagePairType {
			parent.ChildIds = append(parent.ChildIds, childId)
			parent.HasChildren = true
			if parent.LoadOptions.HasGrandChildren {
				newPage.LoadOptions.SubpageCounts = true
			}
			if parent.LoadOptions.RedLinkCountForChildren {
				newPage.LoadOptions.RedLinkCount = true
			}
		} else if piType == WikiPageType && ppType == TagPagePairType {
			parent.RelatedIds = append(parent.RelatedIds, childId)
		}
		return nil
	})
	return err
}

// LoadParentIds loads the page ids for all the parents of the pages in the given pageMap.
func LoadParentIds(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadParentIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pairTypeFilter := ""
	if options.PagePairType != "" {
		pairTypeFilter = "type = '" + options.PagePairType + "' AND"
	}

	pageIds := PageIdsListFromMap(sourcePageMap)
	newPages := make(map[string]*Page)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE `).Add(pairTypeFilter).Add(`childId IN`).AddArgsGroup(pageIds).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (pi.pageId=pp.parentId)
		WHERE (pi.currentEdit>0 AND NOT pi.isDeleted) OR pp.parentId=pp.childId
		`).ToStatement(db).Query()

	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		var ppType string
		err := rows.Scan(&parentId, &childId, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(parentId, pageMap, options.LoadOptions)
		childPage := sourcePageMap[childId]

		if ppType == ParentPagePairType {
			childPage.ParentIds = append(childPage.ParentIds, parentId)
			childPage.HasParents = true
			newPages[newPage.PageId] = newPage
		} else if ppType == RequirementPagePairType {
			childPage.RequirementIds = append(childPage.RequirementIds, parentId)
			options.MasteryMap[parentId] = &Mastery{PageId: parentId}
		} else if ppType == TagPagePairType {
			childPage.TaggedAsIds = append(childPage.TaggedAsIds, parentId)
		} else if ppType == SubjectPagePairType {
			childPage.SubjectIds = append(childPage.SubjectIds, parentId)
			options.MasteryMap[parentId] = &Mastery{PageId: parentId}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}

	// Load if parents have parents
	if options.LoadHasParents && len(newPages) > 0 {
		pageIds = PageIdsListFromMap(newPages)
		rows := database.NewQuery(`
			SELECT childId,sum(1)
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND childId IN`).AddArgsGroup(pageIds).Add(`
			GROUP BY 1`).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId string
			var parents int
			err := rows.Scan(&pageId, &parents)
			if err != nil {
				return fmt.Errorf("failed to scan for grandparents: %v", err)
			}
			pageMap[pageId].HasParents = parents > 0
			return nil
		})
		if err != nil {
			return fmt.Errorf("Failed to load grandparents: %v", err)
		}
	}
	return nil
}

// Load full edit for both the parent and the child of the given page pair. Fall back
// to loading unpublished edits if the user is given
func LoadFullPagesForPair(db *database.DB, pagePair *PagePair, u *CurrentUser) (*Page, *Page, error) {
	parent, err := LoadFullEdit(db, pagePair.ParentId, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while loading parent page: %v", err)
	} else if parent == nil {
		parent, err = LoadFullEdit(db, pagePair.ParentId, u, &LoadEditOptions{LoadNonliveEdit: true})
		if err != nil {
			return nil, nil, fmt.Errorf("Error while loading parent page (2): %v", err)
		} else if parent == nil {
			return nil, nil, fmt.Errorf("Parent page doesn't exist", nil)
		}
	}
	child, err := LoadFullEdit(db, pagePair.ChildId, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while loading child page: %v", err)
	} else if child == nil {
		child, err = LoadFullEdit(db, pagePair.ChildId, u, &LoadEditOptions{LoadNonliveEdit: true})
		if err != nil {
			return nil, nil, fmt.Errorf("Error while loading child page (2): %v", err)
		} else if child == nil {
			return nil, nil, fmt.Errorf("Child page doesn't exist", nil)
		}
	}
	return parent, child, nil
}
