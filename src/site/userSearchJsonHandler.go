// userSearchJsonHandler.go contains the handler for matching a partial query against
// pages' ids, aliases, and titles.
package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

var userSearchHandler = siteHandler{
	URI:         "/json/userSearch/",
	HandlerFunc: userSearchJsonHandler,
}

// userSearchJsonHandler handles the request.
func userSearchJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data searchJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerErrorFail("Error decoding JSON", err)
	}
	if data.Term == "" {
		return pages.HandlerBadRequestFail("No search term specified", nil)
	}

	groupIds := []string{"\"\""}
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
								"terms": { "seeGroupId": [%[2]s] }
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
	}`, escapedTerm, strings.Join(groupIds, ","), core.GroupPageType)
	return searchJsonInternalHandler(params, jsonStr)
}
