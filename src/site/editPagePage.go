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
	editPageTmpls   = append(baseTmpls, "tmpl/editPage.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl")
	editPageOptions = newPageOptions{RequireLogin: true}
)

// editPageTmplData stores the data that we pass to the template file to render the page
type editPageTmplData struct {
	Page    *page
	User    *user.User
	Parents []*page
	Aliases []*alias
}

// These pages serve the edit page, but vary slightly in the parameters they take in the url.
var newPagePage = newPageWithOptions("/edit/", editPageRenderer, editPageTmpls, editPageOptions)
var editPagePage = newPageWithOptions("/edit/{id:[0-9]+}", editPageRenderer, editPageTmpls, editPageOptions)
var editPrivatePagePage = newPageWithOptions("/edit/{id:[0-9]+}/{privacyKey:[0-9]+}", editPageRenderer, editPageTmpls, editPageOptions)

// editPageRenderer renders the edit page page.
func editPageRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data editPageTmplData
	data.User = u
	c := sessions.NewContext(r)

	pageIdStr := mux.Vars(r)["id"]
	// If we are creating a new page, redirect to a new id
	if len(pageIdStr) <= 0 {
		// Check if we already created a new page for this user that the user never saved.
		var p page
		query := fmt.Sprintf(`
			SELECT pageId,privacyKey
			FROM pages
			WHERE edit=0 AND isAutosave AND creatorId=%d`, data.User.Id)
		exists, err := database.QueryRowSql(c, query, &p.PageId, &p.PrivacyKey)
		if err != nil {
			return pages.InternalErrorWith(fmt.Errorf("Couldn't check tags: %v", err))
		} else if !exists {
			rand.Seed(time.Now().UnixNano())
			p.PageId = rand.Int63()
		}
		return pages.RedirectWith(getEditPageUrl(&p))
	}

	// Get page id.
	var pageId int64
	pageId, err = strconv.ParseInt(pageIdStr, 10, 64)
	if err != nil {
		c.Inc("edit_page_failed")
		c.Errorf("Invalid id passed: %s", pageIdStr)
		return pages.InternalErrorWith(err)
	}

	// Load the actual page.
	userIdParam := data.User.Id
	q := r.URL.Query()
	if q.Get("ignoreMySaves") != "" {
		userIdParam = -1
	}
	data.Page, err = loadFullEdit(c, pageId, userIdParam)
	if err != nil {
		c.Inc("edit_page_failed")
		c.Errorf("Couldn't load existing page: %v", err)
		return pages.InternalErrorWith(err)
	} else if data.Page == nil {
		// Set IsAutosave to true, so we can check whether or not to show certain settings
		data.Page = &page{PageId: pageId, Alias: fmt.Sprintf("%d", pageId), IsAutosave: true}
	}
	// Check if the privacy key we got is correct.
	if !data.Page.WasPublished && data.Page.Author.Id == data.User.Id {
		// We can skip privacy key check
	} else if data.Page.PrivacyKey > 0 && fmt.Sprintf("%d", data.Page.PrivacyKey) != mux.Vars(r)["privacyKey"] {
		return pages.UnauthorizedWith(fmt.Errorf("This page is private. Invalid privacy key given."))
	}

	// See how many edits have been commited since

	// Load aliases.
	data.Aliases = make([]*alias, 0)
	query := fmt.Sprintf(`
		SELECT pageId,alias,title
		FROM pages
		WHERE isCurrentEdit`)
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
		c.Inc("edit_page_failed")
		c.Errorf("Couldn't load aliases: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		"GetPageEditUrl": func(p *page) string {
			return getEditPageUrl(p)
		},
		"GetMaxKarmaLock": func() int {
			return getMaxKarmaLock(data.User.Karma)
		},
		"GetEditLevel": func(p *page) string {
			return getEditLevel(p, data.User)
		},
	}
	return pages.StatusOK(data).AddFuncMap(funcMap)
}
