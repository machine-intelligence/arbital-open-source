// pagePage.go serves the page page.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

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
	pageOptions = newPageOptions{LoadUserGroups: true}
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

// loadLikes loads likes corresponding to the given pages and updates the pages.
func loadLikes(c sessions.Context, currentUserId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsStr := core.PageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT userId,pageId,value
		FROM (
			SELECT *
			FROM likes
			WHERE pageId IN (%s)
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var userId int64
		var pageId int64
		var value int
		err := rows.Scan(&userId, &pageId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a like: %v", err)
		}
		page := pageMap[pageId]
		if value > 0 {
			if page.LikeCount >= page.DislikeCount {
				page.LikeScore++
			} else {
				page.LikeScore += 2
			}
			page.LikeCount++
		} else if value < 0 {
			if page.DislikeCount >= page.LikeCount {
				page.LikeScore--
			}
			page.DislikeCount++
		}
		if userId == currentUserId {
			page.MyLikeValue = value
		}
		return nil
	})
	return err
}

// loadVotes loads probability votes corresponding to the given pages and updates the pages.
func loadVotes(c sessions.Context, currentUserId int64, pageIds string, pageMap map[int64]*core.Page, usersMap map[int64]*core.User) error {
	query := fmt.Sprintf(`
		SELECT userId,pageId,value,createdAt
		FROM (
			SELECT *
			FROM votes
			WHERE pageId IN (%s)
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var v core.Vote
		var pageId int64
		err := rows.Scan(&v.UserId, &pageId, &v.Value, &v.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a vote: %v", err)
		}
		if v.Value == 0 {
			return nil
		}
		page := pageMap[pageId]
		if page.Votes == nil {
			page.Votes = make([]*core.Vote, 0, 0)
		}
		page.Votes = append(page.Votes, &v)
		if _, ok := usersMap[v.UserId]; !ok {
			usersMap[v.UserId] = &core.User{Id: v.UserId}
		}
		return nil
	})
	return err
}

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

	// Figure out main page's id
	var pageId int64
	pageAlias := mux.Vars(r)["alias"]
	pageId, err = strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		// Okay, it's not an id, but could be an alias.
		query := fmt.Sprintf(`SELECT pageId FROM aliases WHERE fullName="%s"`, pageAlias)
		exists, err := database.QueryRowSql(c, query, &pageId)
		if err != nil {
			return nil, fmt.Errorf("Couldn't query aliases: %v", err)
		} else if !exists {
			return nil, fmt.Errorf("Page with alias '%s' doesn't exists", pageAlias)
		}
	}

	// Load the main page
	data.Page, err = loadFullEdit(c, pageId, data.User.Id, &loadEditOptions{ignoreParents: true})
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
	err = loadChildrenIds(c, data.PageMap, loadChildrenIdsOptions{ForPages: mainPageMap, LoadHasChildren: true})
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
	err = loadCommentIds(c, data.PageMap, embeddedPageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load comments: %v", err)
	}
	// Add comments to the embedded pages map.
	for id, p := range data.PageMap {
		if p.Type == core.CommentPageType {
			embeddedPageMap[id] = p
		}
	}

	// Load parents
	err = loadParentsIds(c, data.PageMap, loadParentsIdsOptions{ForPages: mainPageMap, LoadHasParents: true})
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
	parentIds := make([]string, len(data.Page.Parents))
	for i, parent := range data.Page.Parents {
		parentIds[i] = fmt.Sprintf("%d", parent.ParentId)
	}
	if len(parentIds) > 0 {
		parentIdsStr := strings.Join(parentIds, ",")
		query := fmt.Sprintf(`
			SELECT childId
			FROM pagePairs AS pp
			WHERE parentId IN (%s) AND childId!=%d
			GROUP BY childId
			HAVING SUM(1)>=%d`, parentIdsStr, data.Page.PageId, len(data.Page.Parents))
		data.RelatedIds, err = loadPageIds(c, query, data.PageMap)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load related ids: %v", err)
		}
	}

	// Load the domains for the primary page
	err = loadDomains(c, u, data.Page, data.PageMap, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load domains: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, &core.LoadPageOptions{LoadText: true})
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
	err = loadAuxPageData(c, data.User.Id, data.PageMap, &options)
	if err != nil {
		return nil, fmt.Errorf("error while loading aux data: %v", err)
	}

	// Load all the votes
	err = loadVotes(c, data.User.Id, fmt.Sprintf("%d", pageId), mainPageMap, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while fetching votes: %v", err)
	}

	// Load links
	err = loadLinks(c, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load links: %v", err)
	}

	// Load child draft
	err = loadChildDraft(c, u.Id, data.Page, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load child draft: %v", err)
	}

	// Load all the groups.
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Load all the users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading users: %v", err)
	}

	// From here on we can render the page successfully. Further queries are nice,
	// but not mandatory, so we are not going to return an error if they fail.

	// Add a visit to embedded pages.
	values := ""
	for _, pg := range embeddedPageMap {
		values += fmt.Sprintf("(%d, %d, '%s'),",
			data.User.Id, pg.PageId, database.Now())
	}
	values = strings.TrimRight(values, ",")
	query := fmt.Sprintf(`
		INSERT INTO visits (userId, pageId, createdAt)
		VALUES %s`, values)
	database.ExecuteSql(c, query)

	return &data, nil
}
