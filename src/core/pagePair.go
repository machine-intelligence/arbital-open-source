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

// Load full edit for both the parent and the child of the given page pair.
func LoadFullEditsForPagePair(db *database.DB, pagePair *PagePair, u *CurrentUser) (*Page, *Page, error) {
	editLoadOptions := &LoadEditOptions{
		LoadNonliveEdit: true,
		PreferLiveEdit:  true,
	}
	parent, err := LoadFullEdit(db, pagePair.ParentId, u, editLoadOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while loading parent page: %v", err)
	} else if parent == nil {
		return nil, nil, fmt.Errorf("Parent page doesn't exist", nil)
	}
	child, err := LoadFullEdit(db, pagePair.ChildId, u, editLoadOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while loading child page: %v", err)
	} else if child == nil {
		return nil, nil, fmt.Errorf("Child page doesn't exist", nil)
	}
	return parent, child, nil
}

// Returns the ids of all the children of the given page (including deleted and unpublished children)
func _getChildren(db *database.DB, pageId string) ([]string, error) {
	children := make([]string, 0)
	rows := db.NewStatement(`
		SELECT childId
		FROM pagePairs
		WHERE parentId=? AND type=?`).Query(pageId, ParentPagePairType)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var childId string
		if err := rows.Scan(&childId); err != nil {
			return fmt.Errorf("failed to scan for childId: %v", err)
		}

		children = append(children, childId)
		return nil
	})
	return children, err
}

// Returns the ids of all the parents of the given page (does *not* include deleted and unpublished parents)
func _getParents(db *database.DB, pageId string) ([]string, error) {
	parents := make([]string, 0)

	rows := database.NewQuery(`
		SELECT parentId
		FROM pagePairs AS pp
		JOIN`).AddPart(PageInfosTable(nil)).Add(`AS pi
		ON pp.parentId=pi.pageId
		WHERE pp.childId=?`, pageId).Add(`AND pp.type=?`, ParentPagePairType).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId string
		if err := rows.Scan(&parentId); err != nil {
			return fmt.Errorf("failed to scan for parentId: %v", err)
		}

		parents = append(parents, parentId)
		return nil
	})
	return parents, err
}

type GetRelatedFunc func(db *database.DB, pageId string) ([]string, error)

// Finds all pages reachable from the source page by recursively following the given getRelated function
// (see: https://en.wikipedia.org/wiki/Transitive_closure)
func _getReachablePages(db *database.DB, sourceId string, getRelated GetRelatedFunc) (map[string]bool, error) {
	reachablePages := make(map[string]bool)
	toVisit := []string{sourceId}

	for len(toVisit) > 0 {
		currentId := toVisit[0]
		toVisit = toVisit[1:]

		reachablePages[currentId] = true

		relatedPages, err := getRelated(db, currentId)
		if err != nil {
			return nil, err
		}

		for _, relatedId := range relatedPages {
			if _, is_visited := reachablePages[relatedId]; !is_visited {
				toVisit = append(toVisit, relatedId)
			}
		}
	}

	return reachablePages, nil
}

// Returns the subgraph composed of the given page and all of its descendants
func GetDescendants(db *database.DB, pageId string) ([]string, error) {
	reachablePagesSet, err := _getReachablePages(db, pageId, _getChildren)
	if err != nil {
		return nil, err
	}

	reachablePagesArray := make([]string, 0)
	for id := range reachablePagesSet {
		reachablePagesArray = append(reachablePagesArray, id)
	}

	return reachablePagesArray, nil
}

// Checks to see if one page is an ancestor of another
func IsAncestor(db *database.DB, potentialAncestorId string, potentialDescendantId string) (bool, error) {
	ancestors, err := _getReachablePages(db, potentialDescendantId, _getParents)
	if err != nil {
		return false, err
	}
	_, isAncestor := ancestors[potentialAncestorId]
	return isAncestor, nil
}
