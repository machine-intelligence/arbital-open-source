// explorePage.go serves the explore template.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

// exploreTmplData stores the data that we pass to the template to render the page
type exploreTmplData struct {
	commonPageData
	RootPageIds []string
}

// explorePage serves the Explore page.
var exploreAllPage = newPage(
	"/explore/",
	exploreRenderer,
	append(baseTmpls,
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// explorePage serves the Explore page.
var explorePage = newPage(
	fmt.Sprintf("/explore/{domain:%s}", core.AliasRegexpStr),
	exploreRenderer,
	append(baseTmpls,
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// exploreRenderer renders the explore page.
func exploreRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var data exploreTmplData
	data.User = u
	data.RootPageIds = make([]string, 0)

	// Load the domain.
	domainAlias := mux.Vars(params.R)["domain"]
	if domainAlias != "" {
		data.Domain = &core.Group{Alias: domainAlias}
		data.User.DomainAlias = data.Domain.Alias
		row := db.NewStatement(`
			SELECT id,name,rootPageId
			FROM groups
			WHERE alias=?`).QueryRow(data.Domain.Alias)
		foundDomain, err := row.Scan(&data.Domain.Id, &data.Domain.Name, &data.Domain.RootPageId)
		if err != nil {
			return pages.Fail("Couldn't retrieve subscription", err)
		} else if !foundDomain {
			return pages.Fail(fmt.Sprintf("Couldn't find the domain: %s", data.Domain.Alias), nil)
		}
		data.RootPageIds = append(data.RootPageIds, fmt.Sprintf("%d", data.Domain.RootPageId))
	}

	// Load the root page(s)
	data.PageMap = make(map[int64]*core.Page)
	if domainAlias == "" {
		rows := db.NewStatement(`
			SELECT parentPair.parentId
			FROM pagePairs AS parentPair
			LEFT JOIN pagePairs AS grandParentPair
			ON (parentPair.parentId=grandParentPair.childId)
			WHERE grandParentPair.parentId IS NULL
			GROUP BY 1
			LIMIT 50`).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan a page id", err)
			}
			p := &core.Page{PageId: pageId}
			data.PageMap[pageId] = p
			data.RootPageIds = append(data.RootPageIds, fmt.Sprintf("%d", pageId))
			return nil
		})
		if err != nil {
			return pages.Fail("error while loading page pairs", err)
		}
	} else {
		data.PageMap[data.Domain.RootPageId] = &core.Page{PageId: data.Domain.RootPageId}
	}

	// Load the children
	err := loadChildrenIds(db, data.PageMap, loadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		return pages.Fail("error while loading children", err)
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, u.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	// Filter unpublished pages.
	for id, p := range data.PageMap {
		if !p.IsCurrentEdit {
			delete(data.PageMap, id)
		}
	}

	// Load auxillary data.
	err = loadAuxPageData(db, data.User.Id, data.PageMap, nil)
	if err != nil {
		return pages.Fail("Couldn't load aux data", err)
	}

	// Load number of red links.
	err = loadRedLinkCount(db, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading red link count", err)
	}

	return pages.StatusOK(&data)
}
