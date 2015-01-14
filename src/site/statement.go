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

type fact struct {
	Id        uint64
	CreatedAt string
	Text      string
	IsSupport bool
}

type statement struct {
	Id              uint64
	Text            string
	SupportFacts    []fact
	OppositionFacts []fact
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
		"tmpl/statement.tmpl", "tmpl/fact.tmpl")...)

// loadStatement loads and returns the statement with the correeponding id from the db.
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

// loadFacts loads and returns the facts with the correeponding statement id from the db.
func loadFacts(c sessions.Context, idStr string) ([]fact, error) {
	facts := make([]fact, 0)
	db, err := database.GetDB(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %v", err)
	}

	c.Infof("querying DB for facts with statementId = %s\n", idStr)
	rows, err := db.Query(`
		SELECT id,createdAt,text,isSupport
		FROM facts
		WHERE statementId=?`, idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query for facts: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var f fact
		err := rows.Scan(
			&f.Id,
			&f.CreatedAt,
			&f.Text,
			&f.IsSupport)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for facts: %v", err)
		}
		facts = append(facts, f)
	}
	return facts, nil
}

// statementRenderer renders the statement page.
func statementRenderer(w http.ResponseWriter, r *http.Request) pages.Result {
	var data statementTmplData
	c := sessions.NewContext(r)

	// Load the statement
	idStr := mux.Vars(r)["id"]
	statement, err := loadStatement(c, idStr)
	if err != nil {
		c.Inc("statement_fetch_fail")
		c.Errorf("error while fetching statement id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	data.Statement = *statement

	// Get all the facts
	var facts []fact
	facts, err = loadFacts(c, idStr)
	if err != nil {
		c.Inc("facts_fetch_fail")
		c.Errorf("error while fetching facts for statement id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	data.Statement.SupportFacts = make([]fact, 0, len(facts))
	data.Statement.OppositionFacts = make([]fact, 0, len(facts))
	for _, f := range facts {
		if f.IsSupport {
			data.Statement.SupportFacts = append(data.Statement.SupportFacts, f)
		} else {
			data.Statement.OppositionFacts = append(data.Statement.OppositionFacts, f)
		}
	}

	c.Inc("statement_page_served_success")
	return pages.StatusOK(data)
}
