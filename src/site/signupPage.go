// signupPage.go serves the signup page.
package site

import (
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// signupData stores the data that we pass to the signup.tmpl to render the page
type signupData struct {
	User        *user.User
	ContinueUrl string
}

// signupPage serves the signup page.
var signupPage = newPageWithOptions(
	"/signup/",
	signupRenderer,
	append(baseTmpls,
		"tmpl/signupPage.tmpl", "tmpl/navbar.tmpl"),
	newPageOptions{})

// signupRenderer renders the signup page.
func signupRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data signupData
	data.User = u
	c := sessions.NewContext(r)

	if data.User.Id <= 0 {
		return pages.RedirectWith(data.User.LoginLink)
	}

	// Check if there are parameters. In this case, this is a form submit
	// request. We can process it and, if successful, redirect the user
	// to the page they came from / were trying to get to.
	q := r.URL.Query()
	if q.Get("firstName") != "" || q.Get("lastName") != "" {
		// This is a form submission.
		inviteCode := strings.ToUpper(q.Get("inviteCode"))
		karma := 0
		if inviteCode == "BAYES" || inviteCode == "LESSWRONG" {
			karma = 10
		} else if inviteCode == "MATRIX" {
			karma = 200
		}
		if data.User.Karma > karma {
			karma = data.User.Karma
		}
		if len(q.Get("firstName")) <= 0 || len(q.Get("lastName")) <= 0 {
			return pages.InternalErrorWith(fmt.Errorf("Must specify both first and last names"))
		}
		hashmap := make(map[string]interface{})
		hashmap["id"] = data.User.Id
		hashmap["firstName"] = q.Get("firstName")
		hashmap["lastName"] = q.Get("lastName")
		hashmap["inviteCode"] = inviteCode
		hashmap["karma"] = karma
		hashmap["createdAt"] = database.Now()
		// NOTE: that we'll be *always* rewriting an existing row here, since a row
		// is created with empty info as soon as the user authenticates our app.
		sql := database.GetInsertSql("users", hashmap, "firstName", "lastName", "inviteCode", "karma")
		if _, err = database.ExecuteSql(c, sql); err != nil {
			c.Errorf("Couldn't update user's record: %v", err)
			return pages.InternalErrorWith(fmt.Errorf("Couldn't update user's record"))
		}
		data.User.FirstName = q.Get("firstName")
		data.User.LastName = q.Get("lastName")
		data.User.Karma = karma
		data.User.IsLoggedIn = true
		err = data.User.Save(w, r)
		if err != nil {
			c.Errorf("Couldn't re-save the user after adding the name: %v", err)
		}
		continueUrl := q.Get("continueUrl")
		if continueUrl == "" {
			continueUrl = "/"
		}
		return pages.RedirectWith(continueUrl)
	}
	data.ContinueUrl = q.Get("continueUrl")
	c.Inc("signup_page_served_success")
	return pages.StatusOK(data)
}
