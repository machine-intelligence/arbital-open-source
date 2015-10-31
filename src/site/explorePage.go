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
		// Get actual domain id
		aliasToIdMap, err := core.LoadAliasToPageIdMap(db, []string{domainAlias})
		if err != nil {
			return pages.Fail("Couldn't convert alias", err)
		}

		var ok bool
		data.DomainId, ok = aliasToIdMap[domainAlias]
		if !ok {
			return pages.Fail("Couldn't find alias", nil)
		}
		data.RootPageIds = append(data.RootPageIds, fmt.Sprintf("%d", data.DomainId))
	}

	// Options for loading the root pages
	loadOptions := (&core.PageLoadOptions{
		Children:                true,
		HasGrandChildren:        true,
		RedLinkCountForChildren: true,
		RedLinkCount:            true,
	}).Add(core.TitlePlusLoadOptions)

	// Load the root page(s)
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	if domainAlias == "" {
		rows := db.NewStatement(`
			SELECT parentPair.parentId
			FROM pagePairs AS parentPair
			LEFT JOIN pagePairs AS grandParentPair
			ON (parentPair.parentId=grandParentPair.childId)
			WHERE grandParentPair.parentId IS NULL
			GROUP BY 1
			LIMIT 10`).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan a page id", err)
			}
			core.AddPageToMap(pageId, data.PageMap, loadOptions)
			data.RootPageIds = append(data.RootPageIds, fmt.Sprintf("%d", pageId))
			return nil
		})
		if err != nil {
			return pages.Fail("error while loading page pairs", err)
		}
	} else {
		core.AddPageToMap(data.DomainId, data.PageMap, loadOptions)
	}

	// Load pages.
	err := core.ExecuteLoadPipeline(db, u, data.PageMap, data.UserMap, data.MasteryMap)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	return pages.StatusOK(&data)
}
