// database.go provides functionality to access the database
package database

import (
	_ "appengine/cloudsql"

	"database/sql"
	"fmt"
	"strings"
	"time"

	"zanaduu3/src/database/dbcore"
	"zanaduu3/src/sessions"
)

const TimeLayout = "2006-01-02 15:04:05"

var (
	PacificLocation, _ = time.LoadLocation("America/Los_Angeles")
)

// ProcessRowCallback is the type of function that will be called once for each row
// loaded from an sql query.
type ProcessRowCallback func(db *DB, rows *Rows) error

type TransactionCallback func(tx *Tx) sessions.Error

// InsertMap is map: DB column name -> value, which together corresponds to one row entry.
type InsertMap map[string]interface{}
type InsertMaps []InsertMap

// DB is our structure for the database. For convenience it wraps around the
// sessions context.
type DB struct {
	db        *sql.DB
	C         sessions.Context
	PrintInfo bool
}

type Stmt struct {
	stmt *sql.Stmt
	// If we statement was constructed from a map, this list will have the values
	args     []interface{}
	QueryStr string
	DB       *DB
}

type Tx struct {
	tx *sql.Tx
	DB *DB
}

type Row struct {
	row  *sql.Row
	stmt *Stmt
	DB   *DB
}

type Rows struct {
	rows *sql.Rows
	stmt *Stmt
	DB   *DB
}

func ToSqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// Now returns the present date and time in a format suitable for SQL insertion.
func Now() string {
	return time.Now().UTC().Format(TimeLayout)
}

// GetKeys returns all the keys in the InsertMap.
func (m InsertMap) GetKeys() []string {
	keys := make([]string, 0)
	for key, _ := range m {
		keys = append(keys, key)
	}
	return keys
}

// ArgsPlaceholder returns the placeholder string for an sql.
// Examples: "(?)", "(?,?,?)", "(?,?),(?,?)"
// Total number of question marks will be argsLen. They will be grouped in
// parenthesis groups countPerGroup at a time.
func ArgsPlaceholder(argsLen int, countPerGroup int) string {
	if argsLen%countPerGroup != 0 {
		return "ERROR: ArgsPlaceholder's parameters were incorrect."
	}
	placeholder := "(?" + strings.Repeat(",?", countPerGroup-1) + ")"
	return placeholder + strings.Repeat(","+placeholder, argsLen/countPerGroup-1)
}

// InArgsPlaceholder returns the placeholder string for an sql.
// Examples: "(?)", "(?,?,?)"
// Total number of question marks will be argsLen.
func InArgsPlaceholder(argsLen int) string {
	if argsLen == 0 {
		return "()"
	}
	return "(?" + strings.Repeat(",?", argsLen-1) + ")"
}

// GetDB returns a DB object, creating it first if necessary.
// GetDB calls db.Ping() on each call to ensure a valid connection.
func GetDB(c sessions.Context) (*DB, error) {
	var (
		db  DB
		err error
	)
	if sessions.Live {
		db.db, err = dbcore.GetLiveCloud(c)
	} else {
		db.db, err = dbcore.GetLocal(c)
	}
	if err != nil {
		c.Inc("db_acquire_fail")
		return nil, fmt.Errorf("Couldn't open DB: %v", err)
	}
	db.C = c
	db.PrintInfo = sessions.Live
	db.db.SetMaxOpenConns(12)
	return &db, nil
}

// NewStatement returns a new SQL statement built from the given SQL string.
// The statement will be executed *non-atomically*.
func (db *DB) NewStatement(query string) *Stmt {
	statement, err := db.db.Prepare(query)
	if err != nil {
		db.C.Errorf("Error creating statement from query:\n%s\n%v", query, err)
		return nil
	}
	return &Stmt{stmt: statement, QueryStr: query, DB: db}
}

// newMultipleInsertStmtInternal creates an INSERT-like query based on the given
// parameters.
func newMultipleInsertStmtInternal(command string, tableName string, insertMaps InsertMaps, updateArgs ...string) *QueryPart {
	// Compute map keys list, since Go doesn't iterate through hashmaps in the same order each time
	rowNames := insertMaps[0].GetKeys()
	rowNamesStr := strings.Join(rowNames, ",")

	// Start the query
	query := NewQuery(command + " INTO " + tableName + "(" + rowNamesStr + ") VALUES")

	// Add the values
	for n, insertMap := range insertMaps {
		argsGroup := make([]interface{}, 0)
		for _, rowName := range rowNames {
			argsGroup = append(argsGroup, insertMap[rowName])
		}
		if n > 0 {
			query.Add(",")
		}
		query.AddArgsGroup(argsGroup)
	}

	// Check if we should update some values in case of key collision.
	if len(updateArgs) > 0 {
		query.Add("ON DUPLICATE KEY UPDATE")
		updateVars := make([]string, 0, len(updateArgs))
		for _, v := range updateArgs {
			updateVars = append(updateVars, v+"=VALUES("+v+")")
		}
		query.Add(strings.Join(updateVars, ","))
	}

	return query
}

// NewInsertStatement returns an SQL statement for inserting a row into the given table.
// The hashmap describes what values to set for that row.
// If there is a name collision, the variables in updateArgs will be updated.
func (db *DB) NewInsertStatement(tableName string, hashmap InsertMap, updateArgs ...string) *Stmt {
	query := newMultipleInsertStmtInternal("INSERT", tableName, InsertMaps{hashmap}, updateArgs...)
	return query.ToStatement(db)
}

// NewReplaceStatement acts just like NewInsertStatement, but does REPLACE instead of INSERT.
func (db *DB) NewReplaceStatement(tableName string, hashmap InsertMap) *Stmt {
	query := newMultipleInsertStmtInternal("REPLACE", tableName, InsertMaps{hashmap})
	return query.ToStatement(db)
}

// NewInsertStatement returns an SQL statement for inserting multiple rows into the given table.
// The array of hashmaps describes what values to set for each row.
// If there is a name collision, the variables in updateArgs will be updated.
func (db *DB) NewMultipleInsertStatement(tableName string, insertMaps InsertMaps, updateArgs ...string) *Stmt {
	query := newMultipleInsertStmtInternal("INSERT", tableName, insertMaps, updateArgs...)
	return query.ToStatement(db)
}

// WithTx converts the given statement to be part of the given transaction.
func (statement *Stmt) WithTx(tx *Tx) *Stmt {
	statement.stmt = tx.tx.Stmt(statement.stmt)
	return statement
}

// Execute executes the given SQL statement.
func (statement *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	if len(statement.args) > 0 {
		if len(args) > 0 {
			return nil, fmt.Errorf("Calling Exec with args, when statement already has args")
		}
		args = statement.args
	}

	startTime := time.Now()
	result, err := statement.stmt.Exec(args...)
	statement.Close()
	if err != nil {
		statement.DB.C.Inc("sql_command_fail")
		return nil, fmt.Errorf("Error while executing an sql statement:\n%v\n%v", statement, err)
	}

	duration := time.Since(startTime)
	if statement.DB.PrintInfo {
		statement.DB.C.Infof(`Executed SQL statement: %v
		With args: %+v
		Operation took %v`, statement, args, duration)
	}
	return result, nil
}

// Query executes the given SQL statement and returns the
// result rows.
func (statement *Stmt) Query(args ...interface{}) *Rows {
	if len(statement.args) > 0 {
		if len(args) > 0 {
			statement.DB.C.Errorf("Calling Query with args, when statement already has args")
			return nil
		}
		args = statement.args
	}

	startTime := time.Now()
	rows, err := statement.stmt.Query(args...)
	if err != nil {
		statement.DB.C.Inc("sql_command_fail")
		statement.DB.C.Errorf("Error while querying:\n%v\n%v", statement, err)
		return nil
	}

	duration := time.Since(startTime)
	if statement.DB.PrintInfo {
		statement.DB.C.Infof(`Executed SQL statement: %v
		With args: %+v
		Operation took %v`, statement, args, duration)
	}
	return &Rows{rows: rows, stmt: statement, DB: statement.DB}
}

// QueryRow executes the given SQL statement and returns the
// result row.
func (statement *Stmt) QueryRow(args ...interface{}) *Row {
	if len(statement.args) > 0 {
		if len(args) > 0 {
			statement.DB.C.Errorf("Calling QueryRow with args, when statement already has args")
			return nil
		}
		args = statement.args
	}
	return &Row{row: statement.stmt.QueryRow(args...), stmt: statement, DB: statement.DB}
}

// String return's statement query
func (statement *Stmt) String() string {
	return statement.QueryStr
}

// Close closes the statement.
func (statement *Stmt) Close() error {
	return statement.stmt.Close()
}

// ProcessRows calls the given function for every row returned when executing the
// given sql statement.
func (rows *Rows) Process(f ProcessRowCallback) error {
	defer rows.rows.Close()
	for rows.rows.Next() {
		if err := f(rows.DB, rows); err != nil {
			return err
		}
	}
	rows.stmt.Close()
	return nil
}

// Scan processes the row and outputs the results into the given variables.
func (rows *Rows) Scan(dest ...interface{}) error {
	return rows.rows.Scan(dest...)
}

// QueryRowSql executes the given SQL statement that's expected to return only
// one row. The results are put into the given args.
// Function returns whether or not a row was read and any errors that occured,
// not including sql.ErrNoRows.
func (row *Row) Scan(outArgs ...interface{}) (bool, error) {
	err := row.row.Scan(outArgs...)
	if err != nil && err != sql.ErrNoRows {
		row.DB.C.Inc("sql_command_fail")
		return false, fmt.Errorf("Error while querying row: %v", err)
	}
	row.stmt.Close()
	return err != sql.ErrNoRows, nil
}

// Transaction calls the given callback with the transaction that will be
// commited when the callback returns.
// If an error occurs, the transaction is rolled back.
func (db *DB) Transaction(f TransactionCallback) sessions.Error {
	tx, err := db.db.Begin()
	if err != nil {
		return sessions.NewError("Couldn't create transaction", err)
	}

	err2 := f(&Tx{tx: tx, DB: db})
	if err2 != nil {
		tx.Rollback()
		return err2
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return sessions.NewError("Couldn't commit transaction", err)
	}
	return nil
}
