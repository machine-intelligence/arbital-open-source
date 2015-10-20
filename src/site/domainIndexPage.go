// domainIndexPage.go serves the index page for a domain.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

const (
	indexPanelLimit = 10
)

// domainIndexTmplData stores the data that we pass to the domainIndex.tmpl to render the page
type domainIndexTmplData struct {
	commonPageData
	MostLikedIds         []string
	MostControversialIds []string
	RecentlyCreatedIds   []string
	RecentlyEditedIds    []string
}

// domainIndexPage serves the domain index page.
var domainIndexPage = newPageWithOptions(
	fmt.Sprintf("/domains/{domain:%s}", core.AliasRegexpStr),
	domainIndexRenderer,
	append(baseTmpls,
		"tmpl/domainIndexPage.tmpl",
		"tmpl/angular.tmpl.js"),
	pages.PageOptions{})

// domainIndexRenderer renders the domain index page.
func domainIndexRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var data domainIndexTmplData
	data.User = u
	data.PageMap = make(map[int64]*core.Page)

	// Load the domain.
	domainAlias := mux.Vars(params.R)["domain"]
	data.User.DomainAlias = domainAlias
	row := db.NewStatement(`
		SELECT pageId
		FROM pages
		WHERE alias=?`).QueryRow(domainAlias)
	foundDomain, err := row.Scan(&data.DomainId)
	if err != nil {
		return pages.Fail("Couldn't retrieve subscription", err)
	} else if !foundDomain {
		return pages.Fail(fmt.Sprintf("Couldn't find the domain: %s", domainAlias), nil)
	}

	// Load recently created page ids.
	rows := db.NewStatement(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		WHERE p.isCurrentEdit AND pd.domainId=?
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(data.DomainId, indexPanelLimit)
	data.RecentlyCreatedIds, err = core.LoadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading recently created page ids", err)
	}

	// Load most liked page ids.
	rows = db.NewStatement(`
		SELECT l2.pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM likes
				ORDER BY id DESC
			) AS l1
			GROUP BY userId,pageId
		) AS l2
		JOIN pageDomainPairs AS pd
		ON (l2.pageId=pd.pageId)
		WHERE pd.domainId=?
		GROUP BY l2.pageId
		ORDER BY SUM(value) DESC
		LIMIT ?`).Query(data.DomainId, indexPanelLimit)
	data.MostLikedIds, err = core.LoadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading most liked page ids", err)
	}

	// Load recently edited page ids.
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,max(createdAt) AS createdAt
			FROM pages
			WHERE NOT isSnapshot AND NOT isAutosave 
			GROUP BY pageId
			HAVING(SUM(1) > 1)
		) AS p
		JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		WHERE pd.domainId=?
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.DomainId, indexPanelLimit)
	data.RecentlyEditedIds, err = core.LoadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading recently edited page ids", err)
	}

	// Load most controversial page ids.
	// TODO: make sure the page still has voting turned on
	rows = db.NewStatement(`
		SELECT pd.pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM votes
				ORDER BY id DESC
			) AS v1
			GROUP BY userId,pageId
		) AS v2
		JOIN pageDomainPairs AS pd
		ON (v2.pageId=pd.pageId)
		WHERE pd.domainId=?
		GROUP BY pd.pageId
		ORDER BY VAR_POP(v2.value) DESC
		LIMIT ?`).Query(data.DomainId, indexPanelLimit)
	data.MostControversialIds, err = core.LoadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading most controversial page ids", err)
	}

	// Load pages.
	data.PageMap[data.DomainId] = &core.Page{PageId: data.DomainId}
	core.AddUserGroupIdsToPageMap(data.User, data.PageMap)
	err = core.LoadPages(db, data.PageMap, u.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	// Load auxillary data.
	err = core.LoadAuxPageData(db, u.Id, data.PageMap, nil)
	if err != nil {
		return pages.Fail("Couldn't load aux data", err)
	}

	return pages.StatusOK(&data)
}
