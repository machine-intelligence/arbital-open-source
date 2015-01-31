// newQuestion.go serves the new question page.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newQuestionTmplData stores the data that we pass to the template file to render the page
type newQuestionTmplData struct {
	User *user.User
	Tags []tag
}

// newQuestionPage serves the question page.
var newQuestionPage = pages.Add(
	"/questions/new/",
	newQuestionRenderer,
	append(baseTmpls,
		"tmpl/newQuestion.tmpl", "tmpl/navbar.tmpl")...)

// newQuestionRenderer renders the new question page.
func newQuestionRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data newQuestionTmplData
	c := sessions.NewContext(r)

	// Load user, if possible
	var err error
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}
	if !data.User.IsLoggedIn {
		return pages.UnauthorizedWith(fmt.Errorf("Not logged in"))
	}

	// Load tags.
	data.Tags = make([]tag, 0)
	query := fmt.Sprintf(`
		SELECT id,text
		FROM tags`)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var t tag
		err := rows.Scan(&t.Id, &t.Text)
		if err != nil {
			return fmt.Errorf("failed to scan for tag: %v", err)
		}
		data.Tags = append(data.Tags, t)
		return nil
	})

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
	}
	c.Inc("new_question_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
