// helpers.go has various functions we use in many places
package site

import (
	"database/sql"
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

//loadTags loads tags for the given claim from the DB.
func loadTags(c sessions.Context, q *claim) error {
	q.Tags = make([]*tag, 0)
	query := fmt.Sprintf(`
		SELECT t.id,t.text
		FROM claimTagPairs AS qt
		LEFT JOIN tags AS t
		ON (qt.tagId=t.id)
		WHERE qt.claimId=%d`, q.Id)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var t tag
		err := rows.Scan(&t.Id, &t.Text)
		if err != nil {
			return fmt.Errorf("failed to scan for tag: %v", err)
		}
		q.Tags = append(q.Tags, &t)
		return nil
	})
	return err
}
