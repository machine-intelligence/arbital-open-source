// contests.go loads and manages all the contests.
package tasks

import (
	"database/sql"
	"fmt"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
)

var (
	contests map[int64]*Contest
)

type Contest struct {
	Id, OwnerId, PayoutId int64
	Hashtag, Text         sql.NullString
}

// processContestRow is called for each row when loading contests from the database.
func processContestRow(rows *sql.Rows) error {
	var c Contest
	err := rows.Scan(
		&c.Id,
		&c.OwnerId,
		&c.PayoutId,
		&c.Hashtag,
		&c.Text)
	if err != nil {
		return fmt.Errorf("failed to scan for user reward: %v", err)
	}
	if c.Text.Valid {
		c.Text = database.ToSqlNullString(simplifyContestText(c.Text.String))
	}
	contests[c.Id] = &c
	return nil
}

// LoadContests loads all the available contests from the contests table.
func LoadContests(c sessions.Context) error {
	contests = make(map[int64]*Contest)
	sql := "SELECT contestId, ownerId, payoutId, hashtag, text FROM contests WHERE isActive"
	if err := database.QuerySql(c, sql, processContestRow); err != nil {
		return fmt.Errorf("Failed to execute sql command to add a contest: %v", err)
	}
	return nil
}
