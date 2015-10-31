// parentsSearchJsonHandler.go contains the handler for matching a partial query against
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

// parentsSearchJsonHandler handles the request.
func parentsSearchJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

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

	// Load user grups
	err = core.LoadUserGroupIds(db, u)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load user groups", err)
	}

	groupIds := append(u.GroupIds, "0")
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
						"must_not": [
							{
								"terms": { "type": ["comment", "answer"] }
							}
						],
						"must": [
							{
								"terms": { "seeGroupId": [%[2]s] }
							}
						]
					}
				}
			}
		},
		"_source": []
	}`, escapedTerm, strings.Join(groupIds, ","))
	return searchJsonInternalHandler(params, jsonStr)
}
