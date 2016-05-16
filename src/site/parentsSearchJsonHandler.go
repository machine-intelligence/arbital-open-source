// parentsSearchJsonHandler.go contains the handler for matching a partial query against
// pages' ids, aliases, and titles.
package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

var parentsSearchHandler = siteHandler{
	URI:         "/json/parentsSearch/",
	HandlerFunc: parentsSearchJsonHandler,
}

// parentsSearchJsonHandler handles the request.
func parentsSearchJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data searchJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Error decoding JSON", err)
	}
	if data.Term == "" {
		return pages.HandlerBadRequestFail("No search term specified", nil)
	}

	groupIds := []string{params.PrivateGroupId}
	escapedTerm := elastic.EscapeMatchTerm(data.Term)

	// Construct the search JSON
	jsonStr := fmt.Sprintf(`{
		"min_score": %[1]v,
		"size": %[2]d,
		"query": {
			"filtered": {
				"query": {
					"bool": {
						"should": [
							{
								"term": { "pageId": "%[3]s" }
							},
							{
								"match_phrase_prefix": { "title": "%[3]s" }
							},
							{
								"match_phrase_prefix": { "alias": "%[3]s" }
							}
						]
					}
				},
				"filter": {
					"bool": {
						"must_not": [
							{
								"terms": { "type": ["comment", "answer"] }
							}
						],
						"must": [
							{
								"terms": { "seeGroupId": [%[4]s] }
							}
						]
					}
				}
			}
		},
		"_source": []
	}`, minSearchScore, searchSize, escapedTerm, strings.Join(groupIds, ","))
	return searchJsonInternalHandler(params, jsonStr)
}
