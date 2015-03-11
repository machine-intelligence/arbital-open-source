// tagHelpers.go contains the tag struct as well as helpful functions.
package site

import (
	"database/sql"
	"fmt"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

type tag struct {
	// DB values.
	Id       int64  `json:",string"`
	ParentId int64  `json:",string"`
	Text     string // e.g. "Vitamin"
	FullName string // e.g. "Health.Nutrition.Vitamin"
}

// loadTagNames loads tag names for each tag in the given map.
func loadTagNames(c sessions.Context, tagMap map[int64]*tag) error {
	if len(tagMap) <= 0 {
		return nil
	}
	tagIds := make([]string, 0, len(tagMap))
	for id, _ := range tagMap {
		tagIds = append(tagIds, fmt.Sprintf("%d", id))
	}
	tagIdsStr := strings.Join(tagIds, ",")
	query := fmt.Sprintf(`
		SELECT id,text
		FROM tags
		WHERE id IN (%s)`, tagIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var t tag
		err := rows.Scan(&t.Id, &t.Text)
		if err != nil {
			return fmt.Errorf("failed to scan for tag: %v", err)
		}
		*tagMap[t.Id] = t
		return nil
	})
	return err
}

// loadTags loads tags corresponding to the given pages.
func loadTags(c sessions.Context, pageIds string, pageMap map[int64]*richPage) error {
	if len(pageIds) <= 0 {
		return nil
	}
	query := fmt.Sprintf(`
		SELECT p.pageId,t.id,t.Text
		FROM pageTagPairs AS p
		LEFT JOIN tags AS t
		ON p.tagId=t.Id
		WHERE p.pageId IN (%s)`, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var t tag
		err := rows.Scan(&pageId, &t.Id, &t.Text)
		if err != nil {
			return fmt.Errorf("failed to scan for pageTagPair: %v", err)
		}
		pageMap[pageId].Tags = append(pageMap[pageId].Tags, &t)
		return nil
	})
	return err
}

// getTagUrl returns URL for looking at recently created pages with the given tag.
func getTagUrl(tagId int64) string {
	return fmt.Sprintf("/pages/filter?tag=%d", tagId)
}
