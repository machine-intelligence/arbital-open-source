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

type TransactionCallback func(tx *Tx) (string, error)

// InsertMap is the map passed in to various database helper functions.
type InsertMap map[string]interface{}

// DB is our structure for the database. For convenience it wraps around the
// sessions context.
type DB struct {
	db *sql.DB
	C  sessions.Context
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
	row *sql.Row
	DB  *DB
}

type Rows struct {
	rows *sql.Rows
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
// Examplers: "(?)", "(?,?,?)", "(?,?),(?,?)"
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

// NewTxStatement returns a new SQL statement built from the given SQL string,
// The statement will be executed *atomically* as part of the given transaction.
func (tx *Tx) NewTxStatement(query string) *Stmt {
	statement, err := tx.tx.Prepare(query)
	if err != nil {
		tx.DB.C.Errorf("Error creating TX statement from query:\n%s\n%v", query, err)
	}
	return &Stmt{stmt: statement, QueryStr: query, DB: tx.DB}
}

// newInsertStmtInternal creates an INSERT-like query based on the given
// parameters.
func newInsertStmtInternal(command string, tableName string, hashmap InsertMap, updateArgs ...string) *QueryPart {
	// Extract hash keys and values
	hashKeys := make([]string, 0, len(hashmap))
	hashValues := make([]interface{}, 0, len(hashmap))
	for k, v := range hashmap {
		hashKeys = append(hashKeys, k)
		hashValues = append(hashValues, v)
	}

	variables := strings.Join(hashKeys, ",")

	// Check if we should update some values in case of key collision.
	onDuplicateKeyPart := NewQuery("")
	if len(updateArgs) > 0 {
		onDuplicateKeyPart.Add("ON DUPLICATE KEY UPDATE")
		updateVars := make([]string, 0, len(updateArgs))
		for _, v := range updateArgs {
			updateVars = append(updateVars, v+"=VALUES("+v+")")
		}
		onDuplicateKeyPart.Add(strings.Join(updateVars, ","))
	}

	query := NewQuery(command + " INTO " + tableName + "(" + variables + ") VALUES").AddArgsGroup(
		hashValues).AddPart(onDuplicateKeyPart)
	return query
}

// NewInsertStatement returns an SQL statement for inserting a row into the given table.
// The hashmap describes what values to set for that row.
// If there is a name collision, the variables in updateArgs will be updated.
func (db *DB) NewInsertStatement(tableName string, hashmap InsertMap, updateArgs ...string) *Stmt {
	query := newInsertStmtInternal("INSERT", tableName, hashmap, updateArgs...)
	return query.ToStatement(db)
}

// NewReplaceStatement acts just like NewInsertStatement, but does REPLACE instead of INSERT.
func (db *DB) NewReplaceStatement(tableName string, hashmap InsertMap) *Stmt {
	query := newInsertStmtInternal("REPLACE", tableName, hashmap)
	return query.ToStatement(db)
}

func (tx *Tx) NewInsertTxStatement(tableName string, hashmap InsertMap, updateArgs ...string) *Stmt {
	query := newInsertStmtInternal("INSERT", tableName, hashmap, updateArgs...)
	return query.ToTxStatement(tx)
}

func (tx *Tx) NewReplaceTxStatement(tableName string, hashmap InsertMap) *Stmt {
	query := newInsertStmtInternal("REPLACE", tableName, hashmap)
	return query.ToTxStatement(tx)
}

// Execute executes the given SQL statement.
func (statement *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	if len(statement.args) > 0 {
		if len(args) > 0 {
			return nil, fmt.Errorf("Calling Exec with args, when statement already has args")
		}
		args = statement.args
	}
	result, err := statement.stmt.Exec(args...)
	if err != nil {
		statement.DB.C.Inc("sql_command_fail")
		return nil, fmt.Errorf("Error while executing an sql statement:\n%v\n%v", statement, err)
	}
	statement.DB.C.Debugf("Executed SQL statement: %v\nwith args: %+v", statement, args)
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
	rows, err := statement.stmt.Query(args...)
	if err != nil {
		statement.DB.C.Inc("sql_command_fail")
		statement.DB.C.Errorf("Error while querying:\n%v\n%v", statement, err)
		return nil
	}
	return &Rows{rows: rows, DB: statement.DB}
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
	return &Row{row: statement.stmt.QueryRow(args...), DB: statement.DB}
}

// String return's statement query
func (statement *Stmt) String() string {
	return statement.QueryStr
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
	return err != sql.ErrNoRows, nil
}

// Transaction calls the given callback with the transaction that will be
// commited when the callback returns.
// If an error occurs, the transaction is rolled back.
func (db *DB) Transaction(f TransactionCallback) (string, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return "Couldn't create transaction", err
	}

	message, err := f(&Tx{tx: tx, DB: db})
	if message != "" {
		tx.Rollback()
		return message, err
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "Couldn't commit transaction", err
	}
	return "", nil
}
