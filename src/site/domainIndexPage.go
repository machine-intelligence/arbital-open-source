// domainIndexPage.go serves the index page for a domain.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

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
	"/domains/{domain:[A-Za-z0-9_-]+}",
	domainIndexRenderer,
	append(baseTmpls,
		"tmpl/domainIndexPage.tmpl",
		"tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl",
		"tmpl/footer.tmpl"),
	newPageOptions{})

// domainIndexRenderer renders the domain index page.
func domainIndexRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)

	data, err := domainIndexInternalRenderer(w, r, u)
	if err != nil {
		c.Inc("domain_index_page_served_fail")
		c.Errorf("%s", err)
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("domain_index_page_served_success")
	return pages.StatusOK(data)
}

// domainIndexInternalRenderer renders the domain index page.
func domainIndexInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*domainIndexTmplData, error) {
	var err error
	var data domainIndexTmplData
	data.User = u
	c := sessions.NewContext(r)
	data.PageMap = make(map[int64]*core.Page)

	// Load the domain.
	data.Domain = &core.Group{Alias: mux.Vars(r)["domain"]}
	data.User.DomainAlias = data.Domain.Alias
	query := fmt.Sprintf(`
		SELECT id,name,rootPageId
		FROM groups
		WHERE alias='%s'`, data.Domain.Alias)
	foundDomain, err := database.QueryRowSql(c, query, &data.Domain.Id, &data.Domain.Name, &data.Domain.RootPageId)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve subscription: %v", err)
	} else if !foundDomain {
		return nil, fmt.Errorf("Couldn't find the domain: %s", data.Domain.Alias)
	}

	// Load recently created page ids.
	query = fmt.Sprintf(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		WHERE p.isCurrentEdit AND pd.domainId=%d
		ORDER BY pi.createdAt DESC
		LIMIT %d`, data.Domain.Id, indexPanelLimit)
	c.Debugf("%s", query)
	data.RecentlyCreatedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently created page ids: %v", err)
	}

	// Load most liked page ids.
	query = fmt.Sprintf(`
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
		WHERE pd.domainId=%d
		GROUP BY l2.pageId
		ORDER BY SUM(value) DESC
		LIMIT %d`, data.Domain.Id, indexPanelLimit)
	data.MostLikedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading most liked page ids: %v", err)
	}

	// Load recently edited page ids.
	query = fmt.Sprintf(`
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
		WHERE pd.domainId=%d
		ORDER BY p.createdAt DESC
		LIMIT %d`, data.Domain.Id, indexPanelLimit)
	data.RecentlyEditedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently edited page ids: %v", err)
	}

	// Load most controversial page ids.
	// TODO: make sure the page still has voting turned on
	query = fmt.Sprintf(`
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
		WHERE pd.domainId=%d
		GROUP BY pd.pageId
		ORDER BY VAR_POP(v2.value) DESC
		LIMIT %d`, data.Domain.Id, indexPanelLimit)
	data.MostControversialIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading most controversial page ids: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, &core.LoadPageOptions{AllowUnpublished: true})
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(c, u.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load aux data: %v", err)
	}

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	return &data, nil
}
