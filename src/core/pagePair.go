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
	ID            string `json:"id"`
	ParentID      string `json:"parentId"`
	ChildID       string `json:"childId"`
	Type          string `json:"type"`
	CreatorID     string `json:"creatorId"`
	CreatedAt     string `json:"createdAt"`
	Level         int    `json:"level"`
	IsStrong      bool   `json:"isStrong"`
	EverPublished bool   `json:"everPublished"`
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
		SELECT pp.id,pp.type,pp.childId,pp.parentId,pp.creatorId,pp.createdAt,pp.everPublished,
			pp.level,pp.isStrong
		FROM pagePairs AS pp`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pagePair PagePair
		err := rows.Scan(&pagePair.ID, &pagePair.Type, &pagePair.ChildID, &pagePair.ParentID,
			&pagePair.CreatorID, &pagePair.CreatedAt, &pagePair.EverPublished, &pagePair.Level,
			&pagePair.IsStrong)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		return callback(db, &pagePair)
	})
	return err
}
func LoadPagePair(db *database.DB, id string) (*PagePair, error) {
	var pagePair *PagePair
	queryPart := database.NewQuery(`WHERE pp.id=?`, id)
	err := LoadPagePairs(db, queryPart, func(db *database.DB, pp *PagePair) error {
		pagePair = pp
		return nil
	})
	return pagePair, err
}

// LoadChildIds loads the page ids for all the children of the pages in the given pageMap.
func LoadChildIDs(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadChildIdsOptions) error {
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

	pageIDs := PageIDsListFromMap(sourcePageMap)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type,pi.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE`).Add(pairTypeFilter).Add(`parentId IN`).AddArgsGroup(pageIDs).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTable(u)).Add(`AS pi
		ON pi.pageId=pp.childId`).Add(pageTypeFilter).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID, childID string
		var ppType string
		var piType string
		err := rows.Scan(&parentID, &childID, &ppType, &piType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(childID, pageMap, options.LoadOptions)

		parent := sourcePageMap[parentID]
		if piType == CommentPageType {
			parent.CommentIDs = append(parent.CommentIDs, newPage.PageID)
		} else if piType == QuestionPageType {
			parent.QuestionIDs = append(parent.QuestionIDs, newPage.PageID)
		} else if piType == WikiPageType && ppType == ParentPagePairType {
			parent.ChildIDs = append(parent.ChildIDs, childID)
			parent.HasChildren = true
			if parent.LoadOptions.HasGrandChildren {
				newPage.LoadOptions.SubpageCounts = true
			}
			if parent.LoadOptions.RedLinkCountForChildren {
				newPage.LoadOptions.RedLinkCount = true
			}
		} else if piType == WikiPageType && ppType == TagPagePairType {
			parent.RelatedIDs = append(parent.RelatedIDs, childID)
		}
		return nil
	})
	return err
}

// LoadParentIds loads the page ids for all the parents of the pages in the given pageMap.
func LoadParentIDs(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadParentIdsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pairTypeFilter := ""
	if options.PagePairType != "" {
		pairTypeFilter = "type = '" + options.PagePairType + "' AND"
	}

	pageIDs := PageIDsListFromMap(sourcePageMap)
	newPages := make(map[string]*Page)

	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pp.type
		FROM (
			SELECT id,parentId,childId,type
			FROM pagePairs
			WHERE `).Add(pairTypeFilter).Add(`childId IN`).AddArgsGroup(pageIDs).Add(`
		) AS pp
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (pi.pageId=pp.parentId)
		WHERE (pi.currentEdit>0 AND NOT pi.isDeleted) OR pp.parentId=pp.childId
		`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID, childID string
		var ppType string
		err := rows.Scan(&parentID, &childID, &ppType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage := AddPageToMap(parentID, pageMap, options.LoadOptions)
		childPage := sourcePageMap[childID]

		if ppType == ParentPagePairType {
			childPage.ParentIDs = append(childPage.ParentIDs, parentID)
			childPage.HasParents = true
			newPages[newPage.PageID] = newPage
		} else if ppType == TagPagePairType {
			childPage.TaggedAsIDs = append(childPage.TaggedAsIDs, parentID)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}

	// Load if parents have parents
	if options.LoadHasParents && len(newPages) > 0 {
		pageIDs = PageIDsListFromMap(newPages)
		rows := database.NewQuery(`
			SELECT childId,sum(1)
			FROM pagePairs
			WHERE type=?`, ParentPagePairType).Add(`AND childId IN`).AddArgsGroup(pageIDs).Add(`
			GROUP BY 1`).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageID string
			var parents int
			err := rows.Scan(&pageID, &parents)
			if err != nil {
				return fmt.Errorf("failed to scan for grandparents: %v", err)
			}
			pageMap[pageID].HasParents = parents > 0
			return nil
		})
		if err != nil {
			return fmt.Errorf("Failed to load grandparents: %v", err)
		}
	}
	return nil
}

type LoadReqsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[string]*Page
	// Mastery map to populate with masteries necessary for a requirement
	MasteryMap map[string]*Mastery
}

// LoadRequisites loads the subjects and requirements for all pages in the map
func LoadRequisites(db *database.DB, pageMap map[string]*Page, u *CurrentUser, options *LoadReqsOptions) error {
	sourcePageMap := options.ForPages
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIDs := PageIDsListFromMap(sourcePageMap)

	queryPart := database.NewQuery(`
		JOIN`).AddPart(PageInfosTableAll(u)).Add(`AS pi
		ON (pi.pageId=pp.parentId)
		WHERE ((pi.currentEdit>0 AND NOT pi.isDeleted) OR pp.parentId=pp.childId)
			AND pp.type IN (?,?)`, RequirementPagePairType, SubjectPagePairType).Add(`
			AND pp.childId IN`).AddArgsGroup(pageIDs).Add(`
		`)
	err := LoadPagePairs(db, queryPart, func(db *database.DB, pagePair *PagePair) error {
		childPage := sourcePageMap[pagePair.ChildID]
		if pagePair.Type == RequirementPagePairType {
			childPage.Requirements = append(childPage.Requirements, pagePair)
			options.MasteryMap[pagePair.ParentID] = &Mastery{PageID: pagePair.ParentID}
		} else if pagePair.Type == SubjectPagePairType {
			childPage.Subjects = append(childPage.Subjects, pagePair)
			options.MasteryMap[pagePair.ParentID] = &Mastery{PageID: pagePair.ParentID}
		}
		AddPageIDToMap(pagePair.ParentID, pageMap)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}

	return nil
}

// Load full edit for both the parent and the child of the given page pair.
func LoadFullEditsForPagePair(db *database.DB, pagePair *PagePair, u *CurrentUser) (*Page, *Page, error) {
	editLoadOptions := &LoadEditOptions{
		LoadNonliveEdit: true,
		PreferLiveEdit:  true,
	}
	parent, err := LoadFullEdit(db, pagePair.ParentID, u, editLoadOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while loading parent page: %v", err)
	} else if parent == nil {
		return nil, nil, fmt.Errorf("Parent page doesn't exist", nil)
	}
	child, err := LoadFullEdit(db, pagePair.ChildID, u, editLoadOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while loading child page: %v", err)
	} else if child == nil {
		return nil, nil, fmt.Errorf("Child page doesn't exist", nil)
	}
	return parent, child, nil
}

// Returns the ids of all the children of the given page (including deleted and unpublished children)
func _getChildren(db *database.DB, pageID string) ([]string, error) {
	children := make([]string, 0)
	rows := db.NewStatement(`
		SELECT childId
		FROM pagePairs
		WHERE parentId=? AND type=?`).Query(pageID, ParentPagePairType)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var childID string
		if err := rows.Scan(&childID); err != nil {
			return fmt.Errorf("failed to scan for childId: %v", err)
		}

		children = append(children, childID)
		return nil
	})
	return children, err
}

// Returns the ids of all the parents of the given page (does *not* include deleted and unpublished parents)
func _getParents(db *database.DB, pageID string) ([]string, error) {
	parents := make([]string, 0)

	rows := database.NewQuery(`
		SELECT parentId
		FROM pagePairs AS pp
		JOIN`).AddPart(PageInfosTable(nil)).Add(`AS pi
		ON pp.parentId=pi.pageId
		WHERE pp.childId=?`, pageID).Add(`AND pp.type=?`, ParentPagePairType).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID string
		if err := rows.Scan(&parentID); err != nil {
			return fmt.Errorf("failed to scan for parentId: %v", err)
		}

		parents = append(parents, parentID)
		return nil
	})
	return parents, err
}

type GetRelatedFunc func(db *database.DB, pageID string) ([]string, error)

// Finds all pages reachable from the source page by recursively following the given getRelated function
// (see: https://en.wikipedia.org/wiki/Transitive_closure)
func _getReachablePages(db *database.DB, sourceID string, getRelated GetRelatedFunc) (map[string]bool, error) {
	reachablePages := make(map[string]bool)
	toVisit := []string{sourceID}

	for len(toVisit) > 0 {
		currentID := toVisit[0]
		toVisit = toVisit[1:]

		reachablePages[currentID] = true

		relatedPages, err := getRelated(db, currentID)
		if err != nil {
			return nil, err
		}

		for _, relatedID := range relatedPages {
			if _, isVisited := reachablePages[relatedID]; !isVisited {
				toVisit = append(toVisit, relatedID)
			}
		}
	}

	return reachablePages, nil
}

// Returns the subgraph composed of the given page and all of its descendants
func GetDescendants(db *database.DB, pageID string) ([]string, error) {
	reachablePagesSet, err := _getReachablePages(db, pageID, _getChildren)
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
func IsAncestor(db *database.DB, potentialAncestorID string, potentialDescendantID string) (bool, error) {
	ancestors, err := _getReachablePages(db, potentialDescendantID, _getParents)
	if err != nil {
		return false, err
	}
	_, isAncestor := ancestors[potentialAncestorID]
	return isAncestor, nil
}
