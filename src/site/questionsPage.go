// questionsPage.go serves the question page.
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

// questionsTmplData stores the data that we pass to the questions.tmpl to render the page
type questionsTmplData struct {
	User      *user.User
	Questions []*question
}

// questionsPage serves the questions page.
var questionsPage = pages.Add(
	"/questions/all/",
	questionsRenderer,
	append(baseTmpls,
		"tmpl/questions.tmpl", "tmpl/navbar.tmpl")...)

// questionsRenderer renders the question page.
func questionsRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data questionsTmplData
	c := sessions.NewContext(r)

	// Load the questions
	data.Questions = make([]*question, 0, 50)
	query := fmt.Sprintf(`
		SELECT id,text
		FROM questions
		WHERE privacyKey IS NULL
		ORDER BY id DESC
		LIMIT 50`)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var q question
		err := rows.Scan(
			&q.Id,
			&q.Text)
		if err != nil {
			return fmt.Errorf("failed to scan a question: %v", err)
		}

		// Load tags.
		err = loadTags(c, &q)
		if err != nil {
			return fmt.Errorf("Couldn't retrieve question tags: %v", err)
		}

		data.Questions = append(data.Questions, &q)
		return nil
	})
	if err != nil {
		c.Errorf("error while loading questions: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load user, if possible
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
	}
	c.Inc("questions_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
