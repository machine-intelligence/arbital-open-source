// domains.go contains functions for dealing with domains (e.g. propagating domain changes)
package core

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// recalculate and update the domains for the given pages
func PropagateDomains(db *database.DB, pagesToUpdate []string) error {
	serr := db.Transaction(func(tx *database.Tx) sessions.Error {
		return sessions.PassThrough(PropagateDomainsWithTx(tx, pagesToUpdate))
	})
	if serr != nil {
		return fmt.Errorf("Failed to propagate domains: %v", serr)
	}
	return nil
}

// recalculate and update the domains for the given pages
func PropagateDomainsWithTx(tx *database.Tx, pagesToUpdate []string) error {
	// map from each page-to-be-updated to its parents
	parentMap, err := _getParentMap(tx, pagesToUpdate)
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}

	// the set of all the parents
	allParentsSet := make(map[string]bool)
	for _, parents := range parentMap {
		for parentId := range parents {
			allParentsSet[parentId] = true
		}
	}

	// both the pages-to-update and their parents
	allPagesSet := make(map[string]bool)
	for _, id := range pagesToUpdate {
		allPagesSet[id] = true
	}
	for id := range allParentsSet {
		allPagesSet[id] = true
	}
	allPagesArray := make([]string, 0)
	for id := range allPagesSet {
		allPagesArray = append(allPagesArray, id)
	}

	// set of pages in our subgraph that are themselves domain pages
	domainPagesSet, err := _getDomainPages(tx, allPagesArray)
	if err != nil {
		return fmt.Errorf("Faled to load domain pages: %v", err)
	}

	// set of pages we're assuming already have the right domains
	// (that is, the parents of the to-be-updated pages that are not
	// also to-be-updated pages themselves)
	pagesWithValidDomainsSet := make(map[string]bool)
	pagesToUpdateSet := make(map[string]bool)
	for _, id := range pagesToUpdate {
		pagesToUpdateSet[id] = true
	}
	for parentId := range allParentsSet {
		if _, toBeUpdated := pagesToUpdateSet[parentId]; !toBeUpdated {
			pagesWithValidDomainsSet[parentId] = true
		}
	}

	// map from pages to their current domains
	// (used to get the domains of the pages-with-valid-domains, and also
	// so that we can diff with our computed domains for the to-be-updated pages)
	originalDomainsMap, err := _getOriginalDomains(tx, allPagesArray)
	if err != nil {
		return fmt.Errorf("Faled to load original domains: %v", err)
	}

	// figure out what the new domains should be
	computedDomainsMap := _getComputedDomains(originalDomainsMap, pagesToUpdate, pagesWithValidDomainsSet, parentMap, domainPagesSet)

	// diff with the original domains, so we can do minimal db writes
	domainsToAddMap, domainsToRemoveMap := _getDomainsToAddRemove(originalDomainsMap, computedDomainsMap)

	// update the domains in the db
	err = _updateDomains(tx, domainsToAddMap, domainsToRemoveMap)
	if err != nil {
		return fmt.Errorf("Faled to update domains: %v", err)
	}

	return nil
}

// gets a map from a set of pages to sets of their parents
func _getParentMap(tx *database.Tx, pageIds []string) (map[string]map[string]bool, error) {
	parentMap := make(map[string]map[string]bool)

	rows := database.NewQuery(`
		SELECT childId, parentId
		FROM pagePairs AS pp
		JOIN`).AddPart(PageInfosTable(nil)).Add(`AS pi
		ON pp.parentId=pi.pageId
		WHERE pp.type=?`, ParentPagePairType).Add(`
			AND childId IN`).AddArgsGroupStr(pageIds).ToTxStatement(tx).Query()
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
func _getDomainPages(tx *database.Tx, pageIds []string) (map[string]bool, error) {
	domainPagesSet := make(map[string]bool)
	rows := database.NewQuery(`
		SELECT pageId
		FROM`).AddPart(PageInfosTable(nil)).Add(`AS pi
		WHERE pageId IN`).AddArgsGroupStr(pageIds).Add(`
			AND type=?`, DomainPageType).ToTxStatement(tx).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		if err := rows.Scan(&pageId); err != nil {
			return fmt.Errorf("failed to scan for domain page: %v", err)
		}
		domainPagesSet[pageId] = true
		return nil
	})

	return domainPagesSet, err
}

// gets the current set of domains for each of the given pages
func _getOriginalDomains(tx *database.Tx, pageIds []string) (map[string]map[string]bool, error) {
	originalDomainsMap := make(map[string]map[string]bool)
	for _, id := range pageIds {
		originalDomainsMap[id] = make(map[string]bool)
	}

	rows := database.NewQuery(`
		SELECT pageId, domainId
		FROM pageDomainPairs
		WHERE pageId IN`).AddArgsGroupStr(pageIds).ToTxStatement(tx).Query()
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
// the domains for pages in pagesWithValidDomainsSet are correct
func _getComputedDomains(originalDomainsMap map[string]map[string]bool, toUpdate []string, pagesWithValidDomainsSet map[string]bool,
	parentMap map[string]map[string]bool, domainPagesSet map[string]bool) map[string]map[string]bool {

	// initialize the map with the page-domain-pairs we believe are already correct
	computedDomainsMap := make(map[string]map[string]bool)
	for id := range pagesWithValidDomainsSet {
		computedDomainsMap[id] = originalDomainsMap[id]
	}

	// re-compute the domains for pages in toUpdate
	for _, id := range toUpdate {
		_computeDomainsRecursive(id, parentMap, domainPagesSet, computedDomainsMap)
	}
	return computedDomainsMap
}

// does a depth-first search up the hierarchy towards parents to find the page's domains
func _computeDomainsRecursive(pageId string, parentMap map[string]map[string]bool, domainPagesSet map[string]bool,
	computedDomainsMap map[string]map[string]bool) map[string]bool {

	// if we already know the domains for this page, we're done
	if computedDomains, ok := computedDomainsMap[pageId]; ok {
		return computedDomains
	}

	domainsSet := make(map[string]bool)

	// add this page if it's a domain page
	if _, isDomainPage := domainPagesSet[pageId]; isDomainPage {
		domainsSet[pageId] = true
	}

	// add the domains from all of this page's parents
	if parents, ok := parentMap[pageId]; ok {
		for parentId := range parents {
			domainsFromParent := _computeDomainsRecursive(parentId, parentMap, domainPagesSet, computedDomainsMap)
			for domainId := range domainsFromParent {
				domainsSet[domainId] = true
			}
		}
	}

	computedDomainsMap[pageId] = domainsSet
	return domainsSet
}

// diffs the new computed domains with the orignal domains
func _getDomainsToAddRemove(originalDomainsMap map[string]map[string]bool, computedDomainsMap map[string]map[string]bool) (
	map[string]map[string]bool, map[string]map[string]bool) {

	domainsToAddMap := make(map[string]map[string]bool)
	domainsToRemoveMap := make(map[string]map[string]bool)

	for id, computedDomains := range computedDomainsMap {
		originalDomains := originalDomainsMap[id]

		domainsToAddSet := make(map[string]bool)
		domainsToRemoveSet := make(map[string]bool)
		for domainId := range computedDomains {
			if _, alreadyApplied := originalDomains[domainId]; !alreadyApplied {
				domainsToAddSet[domainId] = true
			}
		}
		for domainId := range originalDomains {
			if _, shouldKeep := computedDomains[domainId]; !shouldKeep {
				domainsToRemoveSet[domainId] = true
			}
		}

		if len(domainsToAddSet) > 0 {
			domainsToAddMap[id] = domainsToAddSet
		}
		if len(domainsToRemoveSet) > 0 {
			domainsToRemoveMap[id] = domainsToRemoveSet
		}
	}
	return domainsToAddMap, domainsToRemoveMap
}

// adds and removes the specified domains to/from the specified pages
func _updateDomains(tx *database.Tx, domainsToAddMap map[string]map[string]bool, domainsToRemoveMap map[string]map[string]bool) error {
	addDomainArgs := make([]interface{}, 0)
	removeDomainArgsMap := make(map[string][]interface{}, 0)

	for pageId, domainsToAddSet := range domainsToAddMap {
		for domainId := range domainsToAddSet {
			addDomainArgs = append(addDomainArgs, domainId, pageId)
		}
	}
	for pageId, domainsToRemoveSet := range domainsToRemoveMap {
		removeDomainArgs := make([]interface{}, 0)
		for domainId := range domainsToRemoveSet {
			removeDomainArgs = append(removeDomainArgs, domainId)
		}
		if len(removeDomainArgs) > 0 {
			removeDomainArgsMap[pageId] = removeDomainArgs
		}
	}

	// Add missing domains
	if len(addDomainArgs) > 0 {
		statement := tx.DB.NewStatement(`
			INSERT INTO pageDomainPairs
			(domainId,pageId) VALUES ` + database.ArgsPlaceholder(len(addDomainArgs), 2)).WithTx(tx)
		if _, err := statement.Exec(addDomainArgs...); err != nil {
			return fmt.Errorf("Failed to add to pageDomainPairs: %v", err)
		}
	}

	for pageId, removeDomainArgs := range removeDomainArgsMap {
		// Remove obsolete domains
		if len(removeDomainArgs) > 0 {
			statement := tx.DB.NewStatement(`
				DELETE FROM pageDomainPairs
				WHERE pageId=? AND domainId IN ` + database.InArgsPlaceholder(len(removeDomainArgs))).WithTx(tx)
			args := append([]interface{}{pageId}, removeDomainArgs...)
			if _, err := statement.Exec(args...); err != nil {
				return fmt.Errorf("Failed to remove pageDomainPairs: %v", err)
			}
		}
	}

	return nil
}
