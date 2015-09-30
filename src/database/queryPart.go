// queryPart.go contains all the stuff about QueryPart struct
package database

import (
	"strings"
)

// QueryPart is used to construct query piece by piece, while keeping track of
// the arguments.
type QueryPart struct {
	query []string
	args  []interface{}
}

// NewQuery creates a new query part and starts it off with the given content.
func NewQuery(queryStr string, args ...interface{}) *QueryPart {
	var q QueryPart
	q.query = make([]string, 0)
	q.args = make([]interface{}, 0)
	return q.Add(queryStr, args...)
}

// Add adds more to the query and the args.
func (q *QueryPart) Add(queryStr string, args ...interface{}) *QueryPart {
	q.query = append(q.query, " ", queryStr)
	q.args = append(q.args, args...)
	return q
}

// AddArg adds the "?" to the query string, which will be replaced with the given argument.
func (q *QueryPart) AddArg(arg interface{}) *QueryPart {
	q.query = append(q.query, " ?")
	q.args = append(q.args, arg)
	return q
}

// AddArgsGroup adds the arguments that usually preceed "IN" clause.
// Example: q.Add("IN").AddArgsGroup([]interface{}{1,2,3})
// Query gets: "IN (?,?,?)", and the args list gets [1,2,3]
func (q *QueryPart) AddArgsGroup(args []interface{}) *QueryPart {
	q.query = append(q.query, " ", InArgsPlaceholder(len(args)))
	q.args = append(q.args, args...)
	return q
}

// AddPart appends the given query part to this part, mering the query and the args.
func (q *QueryPart) AddPart(part *QueryPart) *QueryPart {
	q.query = append(q.query, " ")
	q.query = append(q.query, part.query...)
	q.args = append(q.args, part.args...)
	return q
}

// Convert the query part into a statement.
func (q *QueryPart) ToStatement(db *DB) *Stmt {
	statement := db.NewStatement(strings.Join(q.query, ""))
	statement.args = q.args
	return statement
}

// Convert the query part into a transaction statement.
func (q *QueryPart) ToTxStatement(tx *Tx) *Stmt {
	statement := tx.NewTxStatement(strings.Join(q.query, ""))
	statement.args = q.args
	return statement
}
