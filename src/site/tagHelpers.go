// tagHelpers.go contains the tag struct as well as helpful functions.
package site

/*import (
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

	// These values are taken from pageTagPairs table.
	PairCreatedBy int64 `json:",string"`
	PairCreatedAt string
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
func loadTags(c sessions.Context, pageMap map[int64]*richPage) error {
	if len(pageMap) <= 0 {
		return nil
	}
	whereClause := "FALSE"
	for id, p := range pageMap {
		whereClause += fmt.Sprintf(" OR (pageId=%d AND edit=%d)", id, p.Edit)
	}
	query := fmt.Sprintf(`
		SELECT p.pageId,p.createdBy,p.createdAt,t.id,t.parentId,t.text,t.fullName
		FROM (
			SELECT pageId,tagId,createdBy,createdAt
			FROM pageTagPairs
			WHERE %s
		) AS p
		LEFT JOIN tags AS t
		ON p.tagId=t.Id`, whereClause)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var t tag
		err := rows.Scan(&pageId, &t.PairCreatedBy, &t.PairCreatedAt, &t.Id, &t.ParentId, &t.Text, &t.FullName)
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
	return fmt.Sprintf("/filter?tag=%d", tagId)
}*/
