// editPagePage.go serves the editPage.tmpl.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

var (
	editPageTmpls   = append(baseTmpls, "tmpl/editPage.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl")
	editPageOptions = newPageOptions{RequireLogin: true}
)

// editPageTmplData stores the data that we pass to the template file to render the page
type editPageTmplData struct {
	Page *page
	User *user.User
	Tags []tag
}

// These pages serve the edit page, but vary slightly in the parameters they take in the url.
var newPagePage = newPageWithOptions("/pages/edit/", editPageRenderer, editPageTmpls, editPageOptions)
var editPagePage = newPageWithOptions("/pages/edit/{id:[0-9]+}", editPageRenderer, editPageTmpls, editPageOptions)
var editPrivatePagePage = newPageWithOptions("/pages/edit/{id:[0-9]+}/{privacyKey:[0-9]+}", editPageRenderer, editPageTmpls, editPageOptions)

// editPageRenderer renders the edit page page.
func editPageRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data editPageTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Check if we are creating a new page or editing an existing one.
	pageIdStr := mux.Vars(r)["id"]
	if len(pageIdStr) > 0 {
		var pageId int64
		pageId, err = strconv.ParseInt(pageIdStr, 10, 64)
		if err != nil {
			c.Inc("edit_page_failed")
			c.Errorf("Invalid id passed: %s", pageIdStr)
			return pages.InternalErrorWith(err)
		}

		// Load the actual page.
		data.Page, err = loadFullPage(c, pageId)
		if err != nil {
			c.Inc("edit_page_failed")
			c.Errorf("Couldn't load existing page: %v", err)
			return pages.InternalErrorWith(err)
		}
		// Check if the privacy key we got is correct.
		if data.Page.PrivacyKey > 0 && fmt.Sprintf("%d", data.Page.PrivacyKey) != mux.Vars(r)["privacyKey"] {
			return pages.UnauthorizedWith(fmt.Errorf("This page is private. Invalid privacy key given."))
		}
	} else {
		data.Page = &page{}
	}

	// Load tags.
	data.Tags = make([]tag, 0)
	query := fmt.Sprintf(`
		SELECT id,parentId,text,fullName
		FROM tags`)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var t tag
		err := rows.Scan(&t.Id, &t.ParentId, &t.Text, &t.FullName)
		if err != nil {
			return fmt.Errorf("failed to scan for tag: %v", err)
		}
		data.Tags = append(data.Tags, t)
		return nil
	})
	if err != nil {
		c.Inc("edit_page_failed")
		c.Errorf("Couldn't load tags: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		// Return the highest karma lock amount a user can create.
		"GetMaxKarmaLock": func() int {
			return getMaxKarmaLock(data.User.Karma)
		},
		"GetEditLevel": func(p *page) int {
			return getEditLevel(p, data.User)
		},
	}
	return pages.StatusOK(data).AddFuncMap(funcMap)
}
