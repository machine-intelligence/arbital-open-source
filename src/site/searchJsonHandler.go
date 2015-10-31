// searchJsonHandler.go contains the handler for searching all the pages.
package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

type searchJsonData struct {
	Term string `json:"term"`
	// If this is set, only pages of this type will be returned
	PageType string
}

// searchJsonHandler handles the request.
func searchJsonHandler(params *pages.HandlerParams) *pages.Result {
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

	// Load user groups
	err = core.LoadUserGroupIds(db, u)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load user groups", err)
	}

	groupIds := append(u.GroupIds, "0")
	escapedTerm := elastic.EscapeMatchTerm(data.Term)

	optionalTermFilter := ""
	if data.PageType != "" {
		optionalTermFilter = fmt.Sprintf(`{"term": { "type": "%s" } },`, elastic.EscapeMatchTerm(data.PageType))
	}

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
								"match": { "title": "%[1]s" }
							},
							{
								"match": { "clickbait": "%[1]s" }
							},
							{
								"match": { "text": "%[1]s" }
							},
							{
								"match_phrase_prefix": { "alias": "%[1]s" }
							}
						]
					}
				},
				"filter": {
					"bool": {
						"must": [`+optionalTermFilter+`
							{
								"terms": { "seeGroupId": [%[2]s] }
							}
						]
					}
				}
			}
		},
		"_source": ["pageId", "alias", "title", "clickbait", "seeGroupId"]
	}`, escapedTerm, strings.Join(groupIds, ","))
	return searchJsonInternalHandler(params, jsonStr)
}

func searchJsonInternalHandler(params *pages.HandlerParams, query string) *pages.Result {
	u := params.U
	db := params.DB

	// Perform search.
	results, err := elastic.SearchPageIndex(params.C, query)
	if err != nil {
		return pages.HandlerErrorFail("Error with elastic search", err)
	}

	pageMap := make(map[int64]*core.Page)
	userMap := make(map[int64]*core.User)
	masteryMap := make(map[int64]*core.Mastery)

	// Create page map.
	for _, hit := range results.Hits.Hits {
		core.AddPageToMap(hit.Id, pageMap, core.TitlePlusLoadOptions)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	returnData := createReturnData(pageMap).AddUsers(userMap).AddMasteries(masteryMap).AddResult(results.Hits)
	return pages.StatusOK(returnData)
}
