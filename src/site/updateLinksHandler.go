// becomeUserHandler.go allows an admin to become any user.
package site

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// updateLinksHandler renders the comment page.
func updateLinksHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	query := fmt.Sprintf(`
		SELECT pageId,text
		FROM pages
		WHERE isCurrentEdit`)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var text string
		err := rows.Scan(&pageId, &text)
		if err != nil {
			return fmt.Errorf("failed to scan for pages: %v", err)
		}

		// NOTE: these regexps are waaaay too simplistic and don't account for the
		// entire complexity of Markdown, like 4 spaces, backticks, and escaped
		// brackets / parens.
		aliasesAndIds := make([]string, 0, 0)
		extractLinks := func(exp *regexp.Regexp) {
			submatches := exp.FindAllStringSubmatch(text, -1)
			for _, submatch := range submatches {
				aliasesAndIds = append(aliasesAndIds, submatch[1])
			}
		}
		// Find directly encoded urls
		extractLinks(regexp.MustCompile(regexp.QuoteMeta(getConfigAddress()) + "/pages/([0-9]+)"))
		// Find ids and aliases using [[id/alias]] syntax.
		extractLinks(regexp.MustCompile("\\[\\[([A-Za-z0-9_-]+?)\\]\\](?:[^(]|$)"))
		// Find ids and aliases using [[text]]((id/alias)) syntax.
		extractLinks(regexp.MustCompile("\\[\\[.+?\\]\\]\\(\\(([A-Za-z0-9_-]+?)\\)\\)"))
		if len(aliasesAndIds) > 0 {
			// Populate linkTuples
			linkMap := make(map[string]bool) // track which aliases we already added to the list
			linkTuples := make([]string, 0, 0)
			for _, alias := range aliasesAndIds {
				if linkMap[alias] {
					continue
				}
				insertValue := fmt.Sprintf("(%d, '%s')", pageId, alias)
				linkTuples = append(linkTuples, insertValue)
				linkMap[alias] = true
			}

			// Insert all the tuples into the links table.
			linkTuplesStr := strings.Join(linkTuples, ",")
			query = fmt.Sprintf(`
				INSERT INTO links (parentId,childAlias)
				VALUES %s`, linkTuplesStr)
			database.ExecuteSql(c, query)
		}
		return nil
	})

	fmt.Fprintf(w, "DONE!!! err: %+v", err)
}
