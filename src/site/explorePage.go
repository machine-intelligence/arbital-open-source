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
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js"))

// explorePage serves the Explore page.
var explorePage = newPage(
	fmt.Sprintf("/explore/{domain:%s}", core.AliasRegexpStr),
	exploreRenderer,
	append(baseTmpls,
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js"))

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
		data.User.DomainAlias = domainAlias
		row := db.NewStatement(`
			SELECT pageId
			FROM pages
			WHERE alias=?`).QueryRow(domainAlias)
		foundDomain, err := row.Scan(&data.DomainId)
		if err != nil {
			return pages.Fail("Couldn't retrieve domain", err)
		} else if !foundDomain {
			return pages.Fail(fmt.Sprintf("Couldn't find the domain: %s", domainAlias), nil)
		}
		data.RootPageIds = append(data.RootPageIds, fmt.Sprintf("%d", data.DomainId))
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
			core.AddPageIdToMap(pageId, data.PageMap)
			data.RootPageIds = append(data.RootPageIds, fmt.Sprintf("%d", pageId))
			return nil
		})
		if err != nil {
			return pages.Fail("error while loading page pairs", err)
		}
	} else {
		core.AddPageIdToMap(data.DomainId, data.PageMap)
	}

	// Load the children
	err := core.LoadChildrenIds(db, data.PageMap, &core.LoadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		return pages.Fail("error while loading children", err)
	}

	// Load pages.
	core.AddUserGroupIdsToPageMap(data.User, data.PageMap)
	err = core.LoadPages(db, data.PageMap, u, nil)
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
	err = core.LoadAuxPageData(db, data.User.Id, data.PageMap, nil)
	if err != nil {
		return pages.Fail("Couldn't load aux data", err)
	}

	// Load number of red links.
	err = core.LoadRedLinkCount(db, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading red link count", err)
	}

	return pages.StatusOK(&data)
}
