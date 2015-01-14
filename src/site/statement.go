// statement.go serves the statement page.
package site

import (
	"fmt"
	"net/http"
	"strconv"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	//"zanaduu3/src/tasks"

	"github.com/gorilla/mux"
	"github.com/hkjn/pages"
)

type statement struct {
	Id   uint64
	Text string
}

// statementTmplData stores the data that we pass to the index.tmpl to render the page
type statementTmplData struct {
	Statement statement
	Error     string
}

// statementPage serves the statement page.
var statementPage = pages.Add(
	"/{id:[0-9]+}",
	statementRenderer,
	append(baseTmpls,
		"tmpl/statement.tmpl")...)

func loadStatement(c sessions.Context, idStr string) (*statement, error) {
	var statement statement
	var err error
	statement.Id, err = strconv.ParseUint(idStr, 10, 63)
	if err != nil {
		return nil, fmt.Errorf("Incorrect id: %s", idStr)
	}

	c.Infof("querying DB for statement with id = %s\n", idStr)
	sql := fmt.Sprintf("SELECT text FROM statements WHERE id=%s", idStr)
	exists, err := database.QueryRowSql(c, sql, &statement.Text)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a statement: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Unknown statement id: %s", idStr)
	}
	return &statement, nil
}

// statementRenderer renders the statement page.
func statementRenderer(w http.ResponseWriter, r *http.Request) pages.Result {
	var data statementTmplData
	c := sessions.NewContext(r)
	//q := r.URL.Query()
	//data.ErrorMsg = q.Get("error_msg")

	id := mux.Vars(r)["id"]
	statement, err := loadStatement(c, id)
	if err != nil {
		c.Inc("statement_fetch_fail")
		c.Errorf("error while fetching statement id: %s\n%v", id, err)
		data.Error = err.Error()
	}
	data.Statement = *statement

	// Update users table to mark this user as active.
	/*err = tasks.UpdateUsersTable(c, &data.User.Twitter)
	if err != nil {
		c.Inc("db_user_insert_query_fail")
		c.Errorf("error while update users table for userId=%d: %v", data.User.Twitter.Id, err)
	}*/

	c.Inc("statement_page_served_success")
	return pages.StatusOK(data)
}
