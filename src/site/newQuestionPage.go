// newQuestion.go serves the new question page.
package site

import (
	"html/template"
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newQuestionTmplData stores the data that we pass to the template file to render the page
type newQuestionTmplData struct {
	User *user.User
}

// newQuestionPage serves the question page.
var newQuestionPage = pages.Add(
	"/questions/new",
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

	funcMap := template.FuncMap{
		"UserId":  func() int64 { return data.User.Id },
		"IsAdmin": func() bool { return data.User.IsAdmin },
	}
	c.Inc("new_question_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
