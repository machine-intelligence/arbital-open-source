// editPagePage.go serves the editPage.tmpl.
package site

import (
	"fmt"
	"math/rand"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/mux"
)

var (
	editPageTmpls   = append(baseTmpls, "tmpl/editPage.tmpl", "tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl")
	editPageOptions = pages.PageOptions{RequireLogin: true}
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
func editPageRenderer(params *pages.HandlerParams) *pages.Result {
	r := params.R
	c := params.C
	db := params.DB
	u := params.U

	pageAlias := mux.Vars(r)["alias"]
	// If we are creating a new page, redirect to a new id
	if len(pageAlias) <= 0 || pageAlias == "0" {
		return pages.RedirectWith(getEditPageUrl(&core.Page{PageId: rand.Int63()}))
	}

	// Check if the user is trying to create a new page with an alias.
	_, err := strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		// Okay, it's not an id, but could be an alias.
		row := db.NewStatement(`SELECT pageId FROM aliases WHERE fullName=?`).QueryRow(pageAlias)
		exists, err := row.Scan(&pageAlias)
		if err != nil {
			return pages.Fail("Couldn't query aliases", err)
		} else if !exists {
			// User is trying to create a new page with an alias.
			return pages.RedirectWith(getEditPageUrl(&core.Page{PageId: rand.Int63()}) + "?alias=" + pageAlias)
		}
		mux.Vars(r)["alias"] = pageAlias
	}

	// Get page id.
	pageIdStr := mux.Vars(r)["alias"]
	pageId, err := strconv.ParseInt(pageIdStr, 10, 64)
	if err != nil {
		return pages.Fail(fmt.Sprintf("Invalid id passed: %s", pageIdStr), nil)
	}

	// Potentially get edit the person wants to load.
	var options loadEditOptions
	q := r.URL.Query()
	loadEdit, err := strconv.ParseInt(q.Get("edit"), 10, 64)
	if err == nil {
		options.loadSpecificEdit = int(loadEdit)
	} else if q.Get("edit") != "live" {
		options.loadNonliveEdit = true
	}

	var data editPageTmplData
	data.User = u

	// Load the actual page.
	data.Page, err = loadFullEdit(db, pageId, data.User.Id, &options)
	if err != nil {
		return pages.Fail("Couldn't load existing page: %v", err)
	} else if data.Page == nil {
		// Set IsAutosave to true, so we can check whether or not to show certain settings
		data.Page = &core.Page{PageId: pageId, Alias: fmt.Sprintf("%d", pageId), IsAutosave: true}
	}
	// Check if the privacy key we got is correct.
	if !data.Page.WasPublished && data.Page.CreatorId == data.User.Id {
		// We can skip privacy key check
	} else if data.Page.PrivacyKey > 0 && fmt.Sprintf("%d", data.Page.PrivacyKey) != mux.Vars(r)["privacyKey"] {
		return pages.Fail("This page is private. Invalid privacy key given.", nil)
	}

	// Load edit history.
	err = core.LoadEditHistory(db, data.Page, data.User.Id)
	if err != nil {
		return pages.Fail("Couldn't load editHistory: %v", err)
	}

	// Load parents
	data.PageMap = make(map[int64]*core.Page)
	data.GroupMap = make(map[int64]*core.Group)
	pageMap := make(map[int64]*core.Page)
	pageMap[data.Page.PageId] = data.Page
	err = data.Page.ProcessParents(c, data.PageMap)
	if err != nil {
		return pages.Fail("Couldn't load parents: %v", err)
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, u.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages: %v", err)
	}
	data.PageMap[data.Page.PageId] = data.Page

	// Load all the groups.
	err = loadGroupNames(db, u, data.GroupMap)
	if err != nil {
		return pages.Fail("Couldn't load group names: %v", err)
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
			hashmap["lockedUntil"] = core.GetPageLockedUntilTime()
			statement := db.NewInsertStatement("pageInfos", hashmap, "lockedBy", "lockedUntil")
			if _, err = statement.Exec(); err != nil {
				return pages.Fail("Couldn't add a lock: %v", err)
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
	err = core.LoadUsers(db, data.UserMap)
	if err != nil {
		return pages.Fail("error while loading users: %v", err)
	}
	data.PrimaryPageId = data.Page.PageId

	// If "alias" URL parameter is set, we want to populate the Alias field.
	aliasUrlParam := q.Get("alias")
	if aliasUrlParam != "" {
		data.Page.Alias = aliasUrlParam
	}
	return pages.StatusOK(data)
}
