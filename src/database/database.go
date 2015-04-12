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

// ProcessRow is the type of function that will be called once for each row
// loaded from an sql query.
type ProcessRow func(c sessions.Context, rows *sql.Rows) error

// InsertMap is the map passed in to various database helper functions.
type InsertMap map[string]interface{}

func ToSqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// Now returns the present date and time in a format suitable for SQL insertion.
func Now() string {
	return time.Now().UTC().Format(TimeLayout)
}

// GetDB returns a DB object, creating it first if necessary.
//
// GetDB calls db.Ping() on each call to ensure a valid connection.
func GetDB(c sessions.Context) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)
	if sessions.Live {
		db, err = dbcore.GetLiveCloud(c)
	} else {
		db, err = dbcore.GetLocal(c)
	}
	if err != nil {
		c.Inc("db_acquire_fail")
	}
	return db, err
}

// sanitizeSqlValue makes sure that a string doesn't have any characters that
// would mess up the sql query.
func sanitizeSqlValue(value string) string {
	value = strings.Replace(value, "\\", "\\\\", -1)
	value = strings.Replace(value, "\"", "\\\"", -1)
	return "\"" + value + "\""
}

// FormatQuery formats the given query such that we can easily copy&paste it
// from the terminal output into sql console.
func FormatQuery(query string) string {
	query = strings.Replace(query, "\n", " ", -1)
	return strings.Replace(query, "\t", "", -1)
}

// GetInsertSql returns an SQL command for inserting a row into the given table.
// The hashmap describes what values to set for that row.
// If there is a name collision, the variables in updateArgs will be updated.
func GetInsertSql(tableName string, hashmap InsertMap, updateArgs ...string) string {
	hashKeys := make([]string, 0, len(hashmap))
	hashValues := make([]string, 0, len(hashmap))
	for k, v := range hashmap {
		hashKeys = append(hashKeys, k)
		vStr := fmt.Sprintf("%v", v)
		switch v.(type) {
		case string:
			vStr = sanitizeSqlValue(vStr)
		}
		hashValues = append(hashValues, vStr)
	}
	variables := strings.Join(hashKeys, ",")
	values := strings.Join(hashValues, ",")
	onDuplicateKeyOpt := ""
	if len(updateArgs) > 0 {
		updateVars := make([]string, 0, len(updateArgs))
		for _, v := range updateArgs {
			updateVars = append(updateVars, fmt.Sprintf("%s=VALUES(%[1]s)", v))
		}
		updateVarsStr := strings.Join(updateVars, ",")
		onDuplicateKeyOpt = "ON DUPLICATE KEY UPDATE " + updateVarsStr
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) %s",
		tableName, variables, values, onDuplicateKeyOpt)
}

// GetReplaceSql acts just like GetInsertSql, but does REPLACE instead of INSERT.
func GetReplaceSql(tableName string, hashmap InsertMap) string {
	query := GetInsertSql(tableName, hashmap)
	return strings.Replace(query, "INSERT", "REPLACE", 1)
}

// ExecuteSql *non-atomically* executes a series of the given SQL commands.
func ExecuteSql(c sessions.Context, commands ...string) (sql.Result, error) {
	db, err := GetDB(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %v", err)
	}
	var result sql.Result
	for _, command := range commands {
		result, err = db.Exec(command)
		if err != nil {
			c.Inc("sql_command_fail")
			return nil, fmt.Errorf("error while executing an sql command:\n%v\n%v", command, err)
		}
		c.Debugf("Executed SQL command: %v", command)
	}
	return result, nil
}

// QuerySql calls the given function for every row returned when executing the
// given sql command.
func QuerySql(c sessions.Context, command string, f ProcessRow) error {
	db, err := GetDB(c)
	if err != nil {
		return fmt.Errorf("failed to get DB: %v", err)
	}
	rows, err := db.Query(command)
	if err != nil {
		c.Inc("sql_command_fail")
		return fmt.Errorf("error while querying:\n%v\n%v", command, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err = f(c, rows); err != nil {
			return err
		}
	}
	return nil
}

// QueryRowSql executes the given SQL command that's expected to return only
// one row. The results are put into the given args.
// Function returns whether or not a row was read and any errors that occured,
// not including sql.ErrNoRows.
func QueryRowSql(c sessions.Context, command string, args ...interface{}) (bool, error) {
	db, err := GetDB(c)
	if err != nil {
		return false, fmt.Errorf("failed to get DB: %v", err)
	}
	err = db.QueryRow(command).Scan(args...)
	if err != nil && err != sql.ErrNoRows {
		c.Inc("sql_command_fail")
		return false, fmt.Errorf("error while querying:\n%v\n%v", command, err)
	}
	return err != sql.ErrNoRows, nil
}

// NewTransaction returns a new transaction object for the database.
func NewTransaction(c sessions.Context) (*sql.Tx, error) {
	db, err := GetDB(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %v", err)
	}
	return db.Begin()
}
