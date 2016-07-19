// domains.go contains functions for dealing with domains (e.g. propagating domain changes)
package core

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// Recalculate and update the domains for the given pages
func PropagateDomains(db *database.DB, pagesToUpdate []string) error {

	serr := db.Transaction(func(tx *database.Tx) sessions.Error {
		_, _, err := PropagateDomainsWithTx(tx, pagesToUpdate)
		return sessions.PassThrough(err)
	})
	if serr != nil {
		return fmt.Errorf("Failed to propagate domains: %v", serr)
	}
	return nil
}

// Recalculate and update the domains for the given pages
func PropagateDomainsWithTx(tx *database.Tx, pagesToUpdate []string) (map[string]map[string]bool,
	map[string]map[string]bool, error) {

	// Map from each page-to-be-updated to its parents
	parentMap, err := _getParentMap(tx, pagesToUpdate)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to load parents: %v", err)
	}

	// The set of all the parents
	allParentsSet := make(map[string]bool)
	for _, parents := range parentMap {
		for parentID := range parents {
			allParentsSet[parentID] = true
		}
	}

	// Both the pages-to-update and their parents
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

	// Set of pages in our subgraph that are themselves domain pages
	domainPagesSet, err := _getDomainPages(tx, allPagesArray)
	if err != nil {
		return nil, nil, fmt.Errorf("Faled to load domain pages: %v", err)
	}

	// Set of pages we're assuming already have the right domains
	// (that is, the parents of the to-be-updated pages that are not
	// also to-be-updated pages themselves)
	pagesWithValidDomainsSet := make(map[string]bool)
	pagesToUpdateSet := make(map[string]bool)
	for _, id := range pagesToUpdate {
		pagesToUpdateSet[id] = true
	}
	for parentID := range allParentsSet {
		if _, toBeUpdated := pagesToUpdateSet[parentID]; !toBeUpdated {
			pagesWithValidDomainsSet[parentID] = true
		}
	}

	// Map from pages to their current domains
	// (used to get the domains of the pages-with-valid-domains, and also
	// so that we can diff with our computed domains for the to-be-updated pages)
	originalDomainsMap, err := _getOriginalDomains(tx, allPagesArray)
	if err != nil {
		return nil, nil, fmt.Errorf("Faled to load original domains: %v", err)
	}

	// Figure out what the new domains should be
	computedDomainsMap := _getComputedDomains(originalDomainsMap, pagesToUpdate, pagesWithValidDomainsSet, parentMap, domainPagesSet)

	// Diff with the original domains, so we can do minimal db writes
	domainsToAddMap, domainsToRemoveMap := _getDomainsToAddRemove(originalDomainsMap, computedDomainsMap)

	// Update the domains in the db
	err = _updateDomains(tx, domainsToAddMap, domainsToRemoveMap)
	if err != nil {
		return nil, nil, fmt.Errorf("Faled to update domains: %v", err)
	}

	return domainsToAddMap, domainsToRemoveMap, nil
}

// Gets a map from a set of pages to sets of their parents
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
		var childID, parentID string
		if err := rows.Scan(&childID, &parentID); err != nil {
			return fmt.Errorf("failed to scan for pagePair: %v", err)
		}
		if _, ok := parentMap[childID]; !ok {
			parentMap[childID] = make(map[string]bool)
		}
		// Add the parent to this page's set of parents
		parentMap[childID][parentID] = true
		return nil
	})

	return parentMap, err
}

// Gets the set of pages, from among those given, that are domain pages
func _getDomainPages(tx *database.Tx, pageIds []string) (map[string]bool, error) {
	domainPagesSet := make(map[string]bool)
	rows := database.NewQuery(`
		SELECT pageId
		FROM`).AddPart(PageInfosTable(nil)).Add(`AS pi
		WHERE pageId IN`).AddArgsGroupStr(pageIds).Add(`
			AND type=?`, DomainPageType).ToTxStatement(tx).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		if err := rows.Scan(&pageID); err != nil {
			return fmt.Errorf("failed to scan for domain page: %v", err)
		}
		domainPagesSet[pageID] = true
		return nil
	})

	return domainPagesSet, err
}

// Gets the current set of domains for each of the given pages.
// Returns a map from page ids to sets of domain ids.
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
		var pageID, domainID string
		if err := rows.Scan(&pageID, &domainID); err != nil {
			return fmt.Errorf("failed to scan for pageDomainPair: %v", err)
		}
		// Add the domain to this page's set of original domains
		originalDomainsMap[pageID][domainID] = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	return originalDomainsMap, nil
}

// Computes the correct domains for each page in toUpdate, assuming that
// the domains for pages in pagesWithValidDomainsSet are correct
func _getComputedDomains(originalDomainsMap map[string]map[string]bool, toUpdate []string, pagesWithValidDomainsSet map[string]bool,
	parentMap map[string]map[string]bool, domainPagesSet map[string]bool) map[string]map[string]bool {

	// Initialize the map with the page-domain-pairs we believe are already correct
	computedDomainsMap := make(map[string]map[string]bool)
	for id := range pagesWithValidDomainsSet {
		computedDomainsMap[id] = originalDomainsMap[id]
	}

	// Re-compute the domains for pages in toUpdate
	for _, id := range toUpdate {
		_computeDomainsRecursive(id, parentMap, domainPagesSet, computedDomainsMap)
	}
	return computedDomainsMap
}

// Does a depth-first search up the hierarchy towards parents to find the page's domains
func _computeDomainsRecursive(pageID string, parentMap map[string]map[string]bool, domainPagesSet map[string]bool,
	computedDomainsMap map[string]map[string]bool) map[string]bool {

	// If we already know the domains for this page, we're done
	if computedDomains, ok := computedDomainsMap[pageID]; ok {
		return computedDomains
	}

	domainsSet := make(map[string]bool)

	// Add this page if it's a domain page
	if _, isDomainPage := domainPagesSet[pageID]; isDomainPage {
		domainsSet[pageID] = true
	}

	// Add the domains from all of this page's parents
	if parents, ok := parentMap[pageID]; ok {
		for parentID := range parents {
			domainsFromParent := _computeDomainsRecursive(parentID, parentMap, domainPagesSet, computedDomainsMap)
			for domainID := range domainsFromParent {
				domainsSet[domainID] = true
			}
		}
	}

	computedDomainsMap[pageID] = domainsSet
	return domainsSet
}

// Diffs the new computed domains with the orignal domains
func _getDomainsToAddRemove(originalDomainsMap map[string]map[string]bool, computedDomainsMap map[string]map[string]bool) (
	map[string]map[string]bool, map[string]map[string]bool) {

	domainsToAddMap := make(map[string]map[string]bool)
	domainsToRemoveMap := make(map[string]map[string]bool)

	for id, computedDomains := range computedDomainsMap {
		originalDomains := originalDomainsMap[id]

		domainsToAddSet := make(map[string]bool)
		domainsToRemoveSet := make(map[string]bool)
		for domainID := range computedDomains {
			if _, alreadyApplied := originalDomains[domainID]; !alreadyApplied {
				domainsToAddSet[domainID] = true
			}
		}
		for domainID := range originalDomains {
			if _, shouldKeep := computedDomains[domainID]; !shouldKeep {
				domainsToRemoveSet[domainID] = true
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

// Adds and removes the specified domains to/from the specified pages
func _updateDomains(tx *database.Tx, domainsToAddMap map[string]map[string]bool, domainsToRemoveMap map[string]map[string]bool) error {
	addDomainArgs := make([]interface{}, 0)
	removeDomainArgsMap := make(map[string][]interface{}, 0)

	for pageID, domainsToAddSet := range domainsToAddMap {
		for domainID := range domainsToAddSet {
			addDomainArgs = append(addDomainArgs, domainID, pageID)
		}
	}
	for pageID, domainsToRemoveSet := range domainsToRemoveMap {
		removeDomainArgs := make([]interface{}, 0)
		for domainID := range domainsToRemoveSet {
			removeDomainArgs = append(removeDomainArgs, domainID)
		}
		if len(removeDomainArgs) > 0 {
			removeDomainArgsMap[pageID] = removeDomainArgs
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

	for pageID, removeDomainArgs := range removeDomainArgsMap {
		// Remove obsolete domains
		if len(removeDomainArgs) > 0 {
			statement := tx.DB.NewStatement(`
				DELETE FROM pageDomainPairs
				WHERE pageId=? AND domainId IN ` + database.InArgsPlaceholder(len(removeDomainArgs))).WithTx(tx)
			args := append([]interface{}{pageID}, removeDomainArgs...)
			if _, err := statement.Exec(args...); err != nil {
				return fmt.Errorf("Failed to remove pageDomainPairs: %v", err)
			}
		}
	}

	return nil
}
