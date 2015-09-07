// editPagePage.go serves the editPage.tmpl.
package site

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

var (
	editPageTmpls   = append(baseTmpls, "tmpl/editPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl")
	editPageOptions = newPageOptions{RequireLogin: true, LoadUserGroups: true}
)

// editPageTmplData stores the data that we pass to the template file to render the page
type editPageTmplData struct {
	commonPageData
	Page *core.Page
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
		rand.Seed(time.Now().UnixNano())
		return pages.RedirectWith(getEditPageUrl(&core.Page{PageId: rand.Int63()}))
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
			return pages.RedirectWith(getEditPageUrl(&core.Page{PageId: rand.Int63()}) + "?alias=" + pageAlias)
		}
		mux.Vars(r)["alias"] = pageAlias
	}

	data, err := editPageInternalRenderer(w, r, u)
	if err != nil {
		c.Errorf("%s", err)
		c.Inc("edit_page_served_fail")
		return pages.InternalErrorWith(err)
	}
	data.PrimaryPageId = data.Page.PageId
	// If "alias" URL parameter is set, we want to populate the Alias field.
	q := r.URL.Query()
	aliasUrlParam := q.Get("alias")
	if aliasUrlParam != "" {
		data.Page.Alias = aliasUrlParam
	}

	funcMap := template.FuncMap{
		"GetPageEditUrl": func(p *core.Page) string {
			return getEditPageUrl(p)
		},
		"GetEditLevel": func(p *core.Page) string {
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
	var options *loadEditOptions = nil
	data.Page, err = loadFullEditWithOptions(c, pageId, data.User.Id, options)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load existing page: %v", err)
	} else if data.Page == nil {
		// Set IsAutosave to true, so we can check whether or not to show certain settings
		data.Page = &core.Page{PageId: pageId, Alias: fmt.Sprintf("%d", pageId), IsAutosave: true}
	}
	// Check if the privacy key we got is correct.
	if !data.Page.WasPublished && data.Page.CreatorId == data.User.Id {
		// We can skip privacy key check
	} else if data.Page.PrivacyKey > 0 && fmt.Sprintf("%d", data.Page.PrivacyKey) != mux.Vars(r)["privacyKey"] {
		return nil, fmt.Errorf("This page is private. Invalid privacy key given.")
	}

	// Load parents
	data.PageMap = make(map[int64]*core.Page)
	data.GroupMap = make(map[int64]*core.Group)
	pageMap := make(map[int64]*core.Page)
	pageMap[data.Page.PageId] = data.Page
	err = data.Page.ProcessParents(c, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load parents: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}
	data.PageMap[data.Page.PageId] = data.Page

	// Load all the groups.
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Grab the lock to this page, but only if we have the right group permissions
	if data.Page.GroupId <= 0 || u.IsMemberOfGroup(data.Page.GroupId) {
		now := database.Now()
		if data.Page.LockedBy <= 0 || data.Page.LockedUntil < now {
			hashmap := make(map[string]interface{})
			hashmap["pageId"] = data.Page.PageId
			hashmap["createdAt"] = database.Now()
			hashmap["currentEdit"] = -1
			hashmap["lockedBy"] = data.User.Id
			hashmap["lockedUntil"] = time.Now().UTC().Add(core.PageLockDuration * time.Second).Format(database.TimeLayout)
			query := database.GetInsertSql("pageInfos", hashmap, "lockedBy", "lockedUntil")
			if _, err = database.ExecuteSql(c, query); err != nil {
				return nil, fmt.Errorf("Couldn't add a lock: %v", err)
			}
			data.Page.LockedBy = hashmap["lockedBy"].(int64)
			data.Page.LockedUntil = hashmap["lockedUntil"].(string)
		}
	}

	// Load all the users.
	data.UserMap = make(map[int64]*core.User)
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	data.UserMap[data.Page.LockedBy] = &core.User{Id: data.Page.LockedBy}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading users: %v", err)
	}

	return &data, nil
}
