// propagateDomainTask.go updates all the page's children to have the right domains.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// PropagateDomainTask is the object that's put into the daemon queue.
type PropagateDomainTask struct {
	PageId string
	// If true, the page was deleted and we should update children + parents
	Deleted bool
}

func (task PropagateDomainTask) Tag() string {
	return "propagateDomain"
}

type domainFlags struct {
	Has        bool
	ShouldHave bool
}

// Check if this task is valid, and we can safely execute it.
func (task PropagateDomainTask) IsValid() error {
	if !core.IsIdValid(task.PageId) {
		return fmt.Errorf("PageId needs to be set")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task PropagateDomainTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("==== PROPAGATE DOMAIN START ====")
	defer c.Infof("==== PROPAGATE DOMAIN COMPLETED ====")

	err = propagateDomainsToPageAndDescendants(db, task.PageId)
	if err != nil {
		return -1, fmt.Errorf("Error propagating domain: %v", err)
	}

	return 0, nil
}

func propagateDomainsToPageAndDescendants(db *database.DB, pageId string) error {
	// all the descendants of the page (plus the page itself)
	pagesToUpdate, err := core.GetDescendants(db, pageId)
	if err != nil {
		return err
	}

	// map from each page-to-be-updated to its parents
	parentMap, err := _getParentMap(db, pagesToUpdate)
	if err != nil {
		return fmt.Errorf("Faled to load parents: %v", err)
	}

	// the set of all the parents
	allParents := make(map[string]bool)
	for _, parents := range parentMap {
		for parentId := range parents {
			allParents[parentId] = true
		}
	}

	// both the pages-to-update and their parents
	allPages := make(map[string]bool)
	for id := range pagesToUpdate {
		allPages[id] = true
	}
	for id := range allParents {
		allPages[id] = true
	}

	// set of pages in our subgraph that are themselves domain pages
	domainPages, err := _getDomainPages(db, allPages)
	if err != nil {
		return fmt.Errorf("Faled to load domain pages: %v", err)
	}

	// set of pages we're assuming already have the right domains
	// (that is, the parents of the to-be-updated pages that are not
	// also to-be-updated pages themselves)
	pagesWithValidDomains := make(map[string]bool)
	for parentId := range allParents {
		if _, toBeUpdated := pagesToUpdate[parentId]; !toBeUpdated {
			pagesWithValidDomains[parentId] = true
		}
	}

	// map from pages to their current domains
	// (used to get the domains of the pages-with-valid-domains, and also
	// so that we can diff with our computed domains for the to-be-updated pages)
	originalDomainsMap, err := _getOriginalDomains(db, allPages)
	if err != nil {
		return fmt.Errorf("Faled to load original domains: %v", err)
	}

	// figure out what the new domains should be
	computedDomainsMap := _getComputedDomains(db, originalDomainsMap, pagesToUpdate, pagesWithValidDomains, parentMap, domainPages)

	// diff with the original domains, so we can do minimal db writes
	domainsToAddMap, domainsToRemoveMap := _getDomainsToAddRemove(originalDomainsMap, computedDomainsMap)

	// update the domains in the db
	err = _updateDomains(db, domainsToAddMap, domainsToRemoveMap)
	if err != nil {
		return fmt.Errorf("Faled to update domains: %v", err)
	}

	return nil
}

// gets a map from a set of pages to sets of their parents
func _getParentMap(db *database.DB, pageIds map[string]bool) (map[string]map[string]bool, error) {
	parentMap := make(map[string]map[string]bool)

	pageIdsArray := make([]string, 0)
	for id := range pageIds {
		pageIdsArray = append(pageIdsArray, id)
	}

	rows := database.NewQuery(`
		SELECT childId, parentId
		FROM pagePairs AS pp
		JOIN`).AddPart(core.PageInfosTable(nil)).Add(`AS pi
		ON pp.parentId=pi.pageId
		WHERE pp.type=?`, core.ParentPagePairType).Add(`
			AND childId IN`).AddArgsGroupStr(pageIdsArray).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var childId, parentId string
		if err := rows.Scan(&childId, &parentId); err != nil {
			return fmt.Errorf("failed to scan for pagePair: %v", err)
		}
		if _, ok := parentMap[childId]; !ok {
			parentMap[childId] = make(map[string]bool)
		}
		// add the parent to this page's set of parents
		parentMap[childId][parentId] = true
		return nil
	})

	return parentMap, err
}

// gets the set of pages, from among those given, that are domain pages
func _getDomainPages(db *database.DB, pageIds map[string]bool) (map[string]bool, error) {
	pageIdsArray := make([]string, 0)
	for id := range pageIds {
		pageIdsArray = append(pageIdsArray, id)
	}

	domainPages := make(map[string]bool)
	rows := database.NewQuery(`
		SELECT pageId
		FROM`).AddPart(core.PageInfosTable(nil)).Add(`AS pi
		WHERE pageId IN`).AddArgsGroupStr(pageIdsArray).Add(`
			AND type=?`, core.DomainPageType).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		if err := rows.Scan(&pageId); err != nil {
			return fmt.Errorf("failed to scan for domain page: %v", err)
		}
		domainPages[pageId] = true
		return nil
	})

	return domainPages, err
}

// gets the current set of domains for each of the given pages
func _getOriginalDomains(db *database.DB, pageIds map[string]bool) (map[string]map[string]bool, error) {
	originalDomainsMap := make(map[string]map[string]bool)
	for id := range pageIds {
		originalDomainsMap[id] = make(map[string]bool)
	}

	pageIdsArray := make([]string, 0)
	for id := range pageIds {
		pageIdsArray = append(pageIdsArray, id)
	}

	rows := database.NewQuery(`
		SELECT pageId, domainId
		FROM pageDomainPairs
		WHERE pageId IN`).AddArgsGroupStr(pageIdsArray).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, domainId string
		if err := rows.Scan(&pageId, &domainId); err != nil {
			return fmt.Errorf("failed to scan for pageDomainPair: %v", err)
		}
		// add the domain to this page's set of original domains
		originalDomainsMap[pageId][domainId] = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	return originalDomainsMap, nil
}

// computes the correct domains for each page in toUpdate, assuming that
// the domains for pages in pagesWithValidDomains are correct
func _getComputedDomains(db *database.DB, originalDomainsMap map[string]map[string]bool, toUpdate map[string]bool, pagesWithValidDomains map[string]bool,
	parentMap map[string]map[string]bool, domainPages map[string]bool) map[string]map[string]bool {

	// initialize the map with the page-domain-pairs we believe are already correct
	computedDomainsMap := make(map[string]map[string]bool)
	for id := range pagesWithValidDomains {
		computedDomainsMap[id] = originalDomainsMap[id]
	}

	// re-compute the domains for pages in toUpdate
	for id := range toUpdate {
		_computeDomainsRecursive(db, id, parentMap, domainPages, computedDomainsMap, make(map[string]bool))
	}
	return computedDomainsMap
}

// does a depth-first search up the hierarchy towards parents to find the page's domains
func _computeDomainsRecursive(db *database.DB, pageId string, parentMap map[string]map[string]bool, domainPages map[string]bool,
	computedDomainsMap map[string]map[string]bool, visiting map[string]bool) map[string]bool {

	// if we already know the domains for this page, we're done
	if computedDomains, ok := computedDomainsMap[pageId]; ok {
		return computedDomains
	}

	db.C.Debugf("visiting: %v", pageId)

	// track pages we're working on so we don't infinitely recurse
	// visiting[pageId] = true
	// defer delete(visiting, pageId)

	domains := make(map[string]bool)

	// add this page if it's a domain page
	if _, isDomainPage := domainPages[pageId]; isDomainPage {
		domains[pageId] = true
	}

	// add the domains from all of this page's parents
	if parents, ok := parentMap[pageId]; ok {
		for parentId := range parents {
			// ROGTODO: is this needed?
			// if _, isVisiting := visiting[parentId]; !isVisiting {
			domainsFromParent := _computeDomainsRecursive(db, parentId, parentMap, domainPages, computedDomainsMap, visiting)
			for domainId := range domainsFromParent {
				domains[domainId] = true
			}
			// }
		}
	}

	computedDomainsMap[pageId] = domains
	return domains
}

// diffs the new computed domains with the orignal domains
func _getDomainsToAddRemove(originalDomainsMap map[string]map[string]bool, computedDomainsMap map[string]map[string]bool) (
	map[string]map[string]bool, map[string]map[string]bool) {

	domainsToAddMap := make(map[string]map[string]bool)
	domainsToRemoveMap := make(map[string]map[string]bool)

	for id, computedDomains := range computedDomainsMap {
		originalDomains := originalDomainsMap[id]

		domainsToAdd := make(map[string]bool)
		domainsToRemove := make(map[string]bool)
		for domainId := range computedDomains {
			if _, alreadyApplied := originalDomains[domainId]; !alreadyApplied {
				domainsToAdd[domainId] = true
			}
		}
		for domainId := range originalDomains {
			if _, shouldKeep := computedDomains[domainId]; !shouldKeep {
				domainsToRemove[domainId] = true
			}
		}

		if len(domainsToAdd) > 0 {
			domainsToAddMap[id] = domainsToAdd
		}
		if len(domainsToRemove) > 0 {
			domainsToRemoveMap[id] = domainsToRemove
		}
	}
	return domainsToAddMap, domainsToRemoveMap
}

// adds and removes the specified domains to/from the specified pages
func _updateDomains(db *database.DB, domainsToAddMap map[string]map[string]bool, domainsToRemoveMap map[string]map[string]bool) error {
	addDomainArgs := make([]interface{}, 0)
	removeDomainArgsMap := make(map[string][]interface{}, 0)

	for pageId, domainsToAdd := range domainsToAddMap {
		for domainId := range domainsToAdd {
			addDomainArgs = append(addDomainArgs, domainId, pageId)
		}
	}
	for pageId, domainsToRemove := range domainsToRemoveMap {
		removeDomainArgs := make([]interface{}, 0)
		for domainId := range domainsToRemove {
			removeDomainArgs = append(removeDomainArgs, domainId)
		}
		if len(removeDomainArgs) > 0 {
			removeDomainArgsMap[pageId] = removeDomainArgs
		}
	}

	// Add missing domains
	if len(addDomainArgs) > 0 {
		statement := db.NewStatement(`
			INSERT INTO pageDomainPairs
			(domainId,pageId) VALUES ` + database.ArgsPlaceholder(len(addDomainArgs), 2))
		if _, err := statement.Exec(addDomainArgs...); err != nil {
			return fmt.Errorf("Failed to add to pageDomainPairs: %v", err)
		}
	}

	for pageId, removeDomainArgs := range removeDomainArgsMap {
		// Remove obsolete domains
		if len(removeDomainArgs) > 0 {
			statement := db.NewStatement(`
				DELETE FROM pageDomainPairs
				WHERE pageId=? AND domainId IN ` + database.InArgsPlaceholder(len(removeDomainArgs)))
			args := append([]interface{}{pageId}, removeDomainArgs...)
			if _, err := statement.Exec(args...); err != nil {
				return fmt.Errorf("Failed to remove pageDomainPairs: %v", err)
			}
		}
	}

	return nil
}
