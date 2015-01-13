// index.go serves the index page.
package site

import (
	//"fmt"
	"net/http"

	"zanaduu3/src/sessions"
	//"zanaduu3/src/tasks"

	"github.com/hkjn/pages"
)

// indexTmplData stores the data that we pass to the index.tmpl to render the page
type indexTmplData struct {
	ErrorMsg string
}

// indexPage serves the index page.
var indexPage = pages.Add(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/main.tmpl")...)

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request) pages.Result {
	var data indexTmplData
	c := sessions.NewContext(r)
	q := r.URL.Query()
	data.ErrorMsg = q.Get("error_msg")

	/*var err error
	data.User, err = user.LoadUser(r)
	if err != nil {
		return pages.InternalErrorWith(fmt.Errorf("error loading user: %v", err))
	}

	// Update users table to mark this user as active.
	err = tasks.UpdateUsersTable(c, &data.User.Twitter)
	if err != nil {
		c.Inc("db_user_insert_query_fail")
		c.Errorf("error while update users table for userId=%d: %v", data.User.Twitter.Id, err)
	}*/

	// Get more data about the user.
	/*err = queryUpdateUser(c, data.User)
	if err != nil {
		return pages.InternalErrorWith(
			fmt.Errorf("error querying for rewards for %d: %v\n", data.User.Twitter.Id, err))
	}*/

	c.Inc("index_page_served_success")
	return pages.StatusOK(data)
}
