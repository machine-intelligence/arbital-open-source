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

const (
	// How many results to get from Elastic
	searchSize = 20
	// How many results to return to the FE
	returnSearchSize = 10
	// Minimum allowed score for results
	minSearchScore = 0.1
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
								"match": { "title": "%[3]s" }
							},
							{
								"match": { "clickbait": "%[3]s" }
							},
							{
								"match": { "text": "%[3]s" }
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
								"terms": { "type": ["comment"] }
							}
						],
						"must": [`+optionalTermFilter+`
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

func searchJsonInternalHandler(params *pages.HandlerParams, query string) *pages.Result {
	db := params.DB

	// Perform search.
	results, err := elastic.SearchPageIndex(params.C, query)
	if err != nil {
		return pages.HandlerErrorFail("Error with elastic search", err)
	}

	returnData := core.NewHandlerData(params.U, false)

	loadOptions := (&core.PageLoadOptions{
		Tags:     true,
		Creators: true,
	}).Add(core.TitlePlusLoadOptions)

	// Create page map.
	for _, hit := range results.Hits.Hits {
		core.AddPageToMap(hit.Id, returnData.PageMap, loadOptions)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Adjust hit values and then extract top N
	// 22t - just a requisite
	// 15r - out of date
	// 4v - work in progress
	// 72 - stub
	penaltyMap := map[string]float32{
		"22t": 0.9,
	}
	for _, hit := range results.Hits.Hits {
		if page, ok := returnData.PageMap[hit.Source.PageId]; ok {
			for _, tagId := range page.TaggedAsIds {
				if penalty, ok := penaltyMap[tagId]; ok {
					hit.Score *= penalty
				}
			}
		} else {
			hit.Score = 0
		}
	}

	results.Hits.Hits = results.Hits.Hits[0:returnSearchSize]
	returnData.ResultMap["search"] = results.Hits
	return pages.StatusOK(returnData.ToJson())
}
