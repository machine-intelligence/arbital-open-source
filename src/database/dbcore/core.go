// Package dbcore provides core DB functionality without being tied to appengine.
package dbcore

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"zanaduu3/src/config"
	"zanaduu3/src/logger"
)

var (
	xc            = config.Load()
	db            *sql.DB // db object - should be accessed through GetDB
	maxDBFailures = 5     // max number of tries before giving up on acquiring DB
	maxOpenDBConn = 30    // max open db connections
	localSQL      = fmt.Sprintf(
		"%s:%s@/%s",
		xc.MySQL.User,
		xc.MySQL.Password,
		xc.MySQL.Database)
	liveSQL = fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?collation=utf8mb4_general_ci",
		xc.MySQL.User,
		xc.MySQL.Password,
		xc.MySQL.Live.Address,
		xc.MySQL.Database)
	liveCloudSQL = fmt.Sprintf(
		"root@cloudsql(%s)/%s?collation=utf8mb4_general_ci",
		xc.MySQL.Live.Instance,
		xc.MySQL.Database)
)

// GetLocal returns a connection to a local MySQL DB.
func GetLocal(pl logger.Logger) (*sql.DB, error) {
	return get(pl, localSQL)
}

// GetLiveCloud returns a connection to the live CloudSQL DB.
//
// Should only be called on live AppEngine.
func GetLiveCloud(pl logger.Logger) (*sql.DB, error) {
	return get(pl, liveCloudSQL)
}

// GetLive returns a connection to the live MySQL DB.
func GetLive(pl logger.Logger) (*sql.DB, error) {
	return get(pl, liveSQL)
}

// get returns a DB object, creating it first if necessary.
//
// get calls db.Ping() on each call to ensure a valid connection.
func get(pl logger.Logger, source string) (*sql.DB, error) {
	// tryOpenDB returns a DB object if possible.
	tryOpenDB := func(failures int) (*sql.DB, error) {
		var (
			err error
			db  *sql.DB
		)
		for failures < maxDBFailures {
			pl.Debugf("Opening the DB..\n")
			db, err = sql.Open("mysql", source)
			if err == nil {
				break
			}
			pl.Warningf("[attempt %d] Failed to Open() db: %v\n", failures, err)
			failures += 1
		}
		if failures >= maxDBFailures {
			return nil, err
		} else {
			return db, nil
		}
	}

	var (
		failures int
		err      error
	)
	for failures < maxDBFailures {
		if db == nil || err != nil {
			// No db object, or one that returned error on last call.
			db, err = tryOpenDB(failures)
			if err != nil {
				return nil, fmt.Errorf("failed to open DB in %d attempts, last error: %v",
					maxDBFailures, err)
			}
			db.SetMaxOpenConns(maxOpenDBConn)
		}
		err = db.Ping()
		if err == nil {
			// Pinged DB successfully.
			return db, nil
		}
		pl.Warningf("[attempt %d] Failed to Ping() db: %v", failures, err)
		failures += 1
	}
	return nil, fmt.Errorf("failed to acquire DB after %d attempts, last error %v", failures, err)
}
