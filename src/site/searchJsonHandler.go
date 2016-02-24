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
	Term string
	// If this is set, only pages of this type will be returned
	PageType string
}

var searchHandler = siteHandler{
	URI:         "/json/search/",
	HandlerFunc: searchJsonHandler,
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

	groupIds := make([]string, 0)
	for _, id := range u.GroupIds {
		groupIds = append(groupIds, "\""+id+"\"")
	}
	groupIds = append(groupIds, "\"\"")

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
						"must_not": [
							{
								"terms": { "type": ["comment"] }
							}
						],
						"must": [`+optionalTermFilter+`
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

func searchJsonInternalHandler(params *pages.HandlerParams, query string) *pages.Result {
	db := params.DB

	// Perform search.
	results, err := elastic.SearchPageIndex(params.C, query)
	if err != nil {
		return pages.HandlerErrorFail("Error with elastic search", err)
	}

	returnData := core.NewHandlerData(params.U, false)

	// Create page map.
	for _, hit := range results.Hits.Hits {
		core.AddPageToMap(hit.Id, returnData.PageMap, core.TitlePlusLoadOptions)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	returnData.ResultMap["search"] = results.Hits
	return pages.StatusOK(returnData.ToJson())
}
