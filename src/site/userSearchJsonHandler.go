// userSearchJsonHandler.go contains the handler for matching a partial query against
// pages' ids, aliases, and titles.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

var userSearchHandler = siteHandler{
	URI:         "/json/userSearch/",
	HandlerFunc: userSearchJSONHandler,
}

// userSearchJsonHandler handles the request.
func userSearchJSONHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data searchJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Error decoding JSON", err)
	}
	if data.Term == "" {
		return pages.Fail("No search term specified", nil).Status(http.StatusBadRequest)
	}

	groupIDs := []string{"\"\""}
	escapedTerm := elastic.EscapeMatchTerm(data.Term)

	// Construct the search JSON
	jsonStr := fmt.Sprintf(`{
		"query": {
			"filtered": {
				"query": {
					"bool": {
						"should": [
							{
								"term": { "pageId": "%[1]s" }
							},
							{
								"match_phrase_prefix": { "title": "%[1]s" }
							},
							{
								"match_phrase_prefix": { "alias": "%[1]s" }
							}
						]
					}
				},
				"filter": {
					"bool": {
						"must": [
							{
								"terms": { "seeDomainId": [%[2]s] }
							},
							{
								"terms": { "type": ["%[3]s"] }
							}
						]
					}
				}
			}
		},
		"_source": []
	}`, escapedTerm, strings.Join(groupIDs, ","), core.GroupPageType)
	return searchJSONInternalHandler(params, jsonStr)
}
