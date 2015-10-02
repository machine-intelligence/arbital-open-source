// pagePage.go serves the page page.
package site

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

type alias struct {
	FullName  string `json:"fullName"`
	PageId    int64  `json:"pageId,string"`
	PageTitle string `json:"pageTitle"`
}

// pageTmplData stores the data that we pass to the index.tmpl to render the page
type pageTmplData struct {
	commonPageData
	Page        *core.Page
	LinkedPages []*core.Page
	RelatedIds  []string
}

var (
	pageOptions = newPageOptions{}
)

// pagePage serves the page page.
var pagePage = newPageWithOptions(
	"/pages/{alias:[A-Za-z0-9_-]+}",
	pageRenderer,
	append(baseTmpls,
		"tmpl/pagePage.tmpl", "tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl", "tmpl/footer.tmpl"), pageOptions)

var privatePagePage = newPageWithOptions(
	"/pages/{alias:[A-Za-z0-9_-]+}/{privacyKey:[0-9]+}",
	pageRenderer,
	append(baseTmpls,
		"tmpl/pagePage.tmpl", "tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl", "tmpl/footer.tmpl"), pageOptions)

// pageRenderer renders the page page.
func pageRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)

	data, err := pageInternalRenderer(w, r, u)
	if err != nil {
		c.Errorf("%s", err)
		c.Inc("page_page_served_fail")
		return showError(w, r, fmt.Errorf("%s", err))
	}

	if data.Page.Type == core.LensPageType {
		// Redirect lens pages to the parent page.
		parentId, _ := strconv.ParseInt(data.Page.ParentsStr, core.PageIdEncodeBase, 64)
		pageUrl := getPageUrl(&core.Page{Alias: fmt.Sprintf("%d", parentId)})
		return pages.RedirectWith(fmt.Sprintf("%s?lens=%d", pageUrl, data.Page.PageId))
	} else if data.Page.Type == core.CommentPageType {
		// Redirect comment pages to the primary page.
		// Note: we are actually redirecting blindly to a parent, which for replies
		// could be the parent comment. For now that's okay, since we just do anther
		// redirect then.
		for _, p := range data.Page.Parents {
			parent := data.PageMap[p.ParentId]
			if parent.Type != core.CommentPageType {
				pageUrl := getPageUrl(&core.Page{Alias: fmt.Sprintf("%d", parent.PageId)})
				return pages.RedirectWith(fmt.Sprintf("%s#comment-%d", pageUrl, data.Page.PageId))
			}
		}
	}

	data.PrimaryPageId = data.Page.PageId

	funcMap := template.FuncMap{
		"GetEditLevel": func(p *core.Page) string {
			return getEditLevel(p, data.User)
		},
		"GetDeleteLevel": func(p *core.Page) string {
			return getDeleteLevel(p, data.User)
		},
		"GetPageEditUrl": func(p *core.Page) string {
			return getEditPageUrl(p)
		},
	}
	c.Inc("page_page_served_success")
	return pages.StatusOK(data).AddFuncMap(funcMap)
}

// pageInternalRenderer renders the page page.
func pageInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*pageTmplData, error) {
	var err error
	var data pageTmplData
	data.User = u
	c := sessions.NewContext(r)

	db, err := database.GetDB(c)
	if err != nil {
		return nil, err
	}

	// Figure out main page's id
	var pageId int64
	pageAlias := mux.Vars(r)["alias"]
	pageId, err = strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		// Okay, it's not an id, but could be an alias.
		row := db.NewStatement(`SELECT pageId FROM aliases WHERE fullName=?`).QueryRow(pageAlias)
		exists, err := row.Scan(&pageId)
		if err != nil {
			return nil, fmt.Errorf("Couldn't query aliases: %v", err)
		} else if !exists {
			return nil, fmt.Errorf("Page with alias '%s' doesn't exists", pageAlias)
		}
	}

	// Load the main page
	data.Page, err = loadFullEdit(db, pageId, data.User.Id, &loadEditOptions{ignoreParents: true})
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if data.Page == nil {
		return nil, fmt.Errorf("Couldn't find a page with id: %d", pageId)
	}

	// Redirect lens pages to the parent page.
	if data.Page.Type == core.LensPageType {
		return &data, nil
	}

	// Check privacy setting
	if data.Page.PrivacyKey > 0 {
		privacyKey := mux.Vars(r)["privacyKey"]
		if privacyKey != fmt.Sprintf("%d", data.Page.PrivacyKey) {
			return nil, fmt.Errorf("Unauthorized access. You don't have the correct privacy key.")
		}
	}

	// Create maps.
	mainPageMap := make(map[int64]*core.Page)
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	data.GroupMap = make(map[int64]*core.Group)
	mainPageMap[data.Page.PageId] = data.Page

	// Load children
	err = loadChildrenIds(db, data.PageMap, loadChildrenIdsOptions{ForPages: mainPageMap, LoadHasChildren: true})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load children: %v", err)
	}

	// Create embedded pages map, which will have pages that are displayed more
	// fully and need additional info loaded.
	embeddedPageMap := make(map[int64]*core.Page)
	embeddedPageMap[data.Page.PageId] = data.Page
	if data.Page.Type == core.QuestionPageType {
		for id, p := range data.PageMap {
			if p.Type == core.AnswerPageType {
				embeddedPageMap[id] = p
			}
		}
	}

	// Load comment ids.
	err = loadCommentIds(db, data.PageMap, embeddedPageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load comments: %v", err)
	}

	// Load question ids.
	err = loadQuestionIds(db, data.PageMap, embeddedPageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load questions: %v", err)
	}

	// Add comments to the embedded pages map.
	for id, p := range data.PageMap {
		if p.Type == core.CommentPageType {
			embeddedPageMap[id] = p
		}
	}

	// Load parents
	err = loadParentsIds(db, data.PageMap, loadParentsIdsOptions{ForPages: mainPageMap, LoadHasParents: true})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load parents: %v", err)
	}

	// Load where page is linked from.
	// TODO: also account for old aliases
	/*query := fmt.Sprintf(`
		SELECT p.pageId
		FROM links as l
		JOIN pages as p
		ON l.parentId=p.pageId
		WHERE (l.childAlias=%d || l.childAlias="%s") AND p.isCurrentEdit
		GROUP BY p.pageId`, pageId, data.Page.Alias)
	data.Page.LinkedFrom, err = loadPageIds(c, query, mainPageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load contexts: %v", err)
	}*/

	// Load page ids of related pages (pages that have at least all the same parents).
	parentIds := make([]interface{}, len(data.Page.Parents))
	for i, parent := range data.Page.Parents {
		parentIds[i] = parent.ParentId
	}
	if len(parentIds) > 0 {
		rows := database.NewQuery(`
			SELECT childId
			FROM pagePairs AS pp
			WHERE parentId IN`).AddArgsGroup(parentIds).Add(` AND childId != ?`, data.Page.PageId).Add(`
			GROUP BY childId
			HAVING SUM(1)>=?`, len(data.Page.Parents)).ToStatement(db).Query()
		data.RelatedIds, err = loadPageIds(rows, data.PageMap)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load related ids: %v", err)
		}
	}

	// Load the domains for the primary page
	err = loadDomains(db, u, data.Page, data.PageMap, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load domains: %v", err)
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, u.Id, &core.LoadPageOptions{LoadText: true})
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Erase Text from pages that don't need it.
	for _, p := range data.PageMap {
		if (data.Page.Type != core.QuestionPageType || p.Type != core.AnswerPageType) && p.Type != core.CommentPageType {
			p.Text = ""
		}
	}

	// From here on, we also load info for the main page as well.
	data.PageMap[data.Page.PageId] = data.Page

	// Load auxillary data.
	q := r.URL.Query()
	options := loadAuxPageDataOptions{ForcedLastVisit: q.Get("lastVisit")}
	err = loadAuxPageData(db, data.User.Id, data.PageMap, &options)
	if err != nil {
		return nil, fmt.Errorf("error while loading aux data: %v", err)
	}

	// Load all the votes
	err = loadVotes(db, data.User.Id, mainPageMap, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while fetching votes: %v", err)
	}

	// Load links
	err = loadLinks(db, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load links: %v", err)
	}

	// Load child draft
	err = loadChildDraft(db, u.Id, data.Page, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load child draft: %v", err)
	}

	// Load all the groups.
	err = loadGroupNames(db, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Load all the users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading users: %v", err)
	}

	// From here on we can render the page successfully. Further queries are nice,
	// but not mandatory, so we are not going to return an error if they fail.

	// Add a visit to embedded pages.
	args := make([]interface{}, 0, 3*len(embeddedPageMap))
	for _, pg := range embeddedPageMap {
		args = append(args, data.User.Id, pg.PageId, database.Now())
	}
	statement := db.NewStatement(`
		INSERT INTO visits (userId, pageId, createdAt)
		VALUES` + database.ArgsPlaceholder(len(args), 3))
	_, err = statement.Exec(args...)
	if err != nil {
		c.Errorf("Error updating visits: %v", err)
	}

	return &data, nil
}
