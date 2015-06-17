// editPagePage.go serves the editPage.tmpl.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

var (
	editPageTmpls   = append(baseTmpls, "tmpl/editPage.tmpl", "tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl")
	editPageOptions = newPageOptions{RequireLogin: true, LoadUserGroups: true}
)

// editPageTmplData stores the data that we pass to the template file to render the page
type editPageTmplData struct {
	PageMap map[int64]*page
	Page    *page
	User    *user.User
	UserMap map[int64]*dbUser
	Aliases []*alias
}

// These pages serve the edit page, but vary slightly in the parameters they take in the url.
var newPagePage = newPageWithOptions("/edit/", editPageRenderer, editPageTmpls, editPageOptions)
var editPagePage = newPageWithOptions("/edit/{alias:[A-Za-z0-9_-]+}", editPageRenderer, editPageTmpls, editPageOptions)
var editPrivatePagePage = newPageWithOptions("/edit/{alias:[A-Za-z0-9_-]+}/{privacyKey:[0-9]+}", editPageRenderer, editPageTmpls, editPageOptions)

// editPageRenderer renders the page page.
func editPageRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)

	pageAlias := mux.Vars(r)["alias"]
	// If we are creating a new page, redirect to a new id
	if len(pageAlias) <= 0 {
		// Check if we already created a new page for this user that the user never saved.
		var p page
		query := fmt.Sprintf(`
			SELECT pageId,privacyKey
			FROM pages
			WHERE edit=0 AND isAutosave AND creatorId=%d`, u.Id)
		exists, err := database.QueryRowSql(c, query, &p.PageId, &p.PrivacyKey)
		if err != nil {
			return pages.InternalErrorWith(fmt.Errorf("Couldn't check tags: %v", err))
		} else if !exists {
			rand.Seed(time.Now().UnixNano())
			p.PageId = rand.Int63()
		}
		return pages.RedirectWith(getEditPageUrl(&p))
	}

	// Check if the user is trying to create a new page with an alias.
	_, err := strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		// Okay, it's not an id, but could be an alias.
		query := fmt.Sprintf(`SELECT pageId FROM aliases WHERE fullName="%s"`, pageAlias)
		exists, err := database.QueryRowSql(c, query, &pageAlias)
		if err != nil {
			return pages.InternalErrorWith(fmt.Errorf("Couldn't query aliases: %v", err))
		} else if !exists {
			// User is trying to create a new page with an alias.
			rand.Seed(time.Now().UnixNano())
			return pages.RedirectWith(getEditPageUrl(&page{PageId: rand.Int63()}) + "?alias=" + pageAlias)
		}
		mux.Vars(r)["alias"] = pageAlias
	}

	data, err := editPageInternalRenderer(w, r, u)
	if err != nil {
		c.Errorf("%s", err)
		c.Inc("edit_page_served_fail")
		return pages.InternalErrorWith(err)
	}
	// If "alias" URL parameter is set, we want to populate the Alias field.
	q := r.URL.Query()
	aliasUrlParam := q.Get("alias")
	if aliasUrlParam != "" {
		data.Page.Alias = aliasUrlParam
	}

	funcMap := template.FuncMap{
		"GetPageGroupName": func() string {
			return data.Page.Group.Name
		},
		"GetPageEditUrl": func(p *page) string {
			return getEditPageUrl(p)
		},
		"GetEditLevel": func(p *page) string {
			return getEditLevel(p, data.User)
		},
	}
	c.Inc("edit_page_served_success")
	return pages.StatusOK(data).AddFuncMap(funcMap)
}

// editPageInternalRenderer renders the edit page page.
func editPageInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*editPageTmplData, error) {
	var err error
	var data editPageTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Get page id.
	pageIdStr := mux.Vars(r)["alias"]
	var pageId int64
	pageId, err = strconv.ParseInt(pageIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid id passed: %s", pageIdStr)
	}

	// Load the actual page.
	userIdParam := data.User.Id
	q := r.URL.Query()
	if q.Get("ignoreMySaves") != "" {
		userIdParam = -1
	}
	data.Page, err = loadFullEdit(c, pageId, userIdParam)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load existing page: %v", err)
	} else if data.Page == nil {
		// Set IsAutosave to true, so we can check whether or not to show certain settings
		data.Page = &page{PageId: pageId, Alias: fmt.Sprintf("%d", pageId), IsAutosave: true}
	}
	// Check if the privacy key we got is correct.
	if !data.Page.WasPublished && data.Page.Author.Id == data.User.Id {
		// We can skip privacy key check
	} else if data.Page.PrivacyKey > 0 && fmt.Sprintf("%d", data.Page.PrivacyKey) != mux.Vars(r)["privacyKey"] {
		return nil, fmt.Errorf("This page is private. Invalid privacy key given.")
	}

	// Load parents
	data.PageMap = make(map[int64]*page)
	pageMap := make(map[int64]*page)
	pageMap[data.Page.PageId] = data.Page
	err = data.Page.processParents(c, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load parents: %v", err)
	}

	// Load pages.
	err = loadPages(c, data.PageMap, u.Id, loadPageOptions{})
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}
	data.PageMap[data.Page.PageId] = data.Page

	// Load aliases.
	data.Aliases = make([]*alias, 0)
	query := fmt.Sprintf(`
		SELECT pageId,alias,title
		FROM pages
		WHERE isCurrentEdit AND (groupName="" OR groupName IN (SELECT groupName FROM groupMembers WHERE userId=%d))`,
		data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var a alias
		err := rows.Scan(&a.PageId, &a.FullName, &a.PageTitle)
		if err != nil {
			return fmt.Errorf("failed to scan for aliases: %v", err)
		}
		data.Aliases = append(data.Aliases, &a)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load aliases: %v", err)
	}

	return &data, nil
}
