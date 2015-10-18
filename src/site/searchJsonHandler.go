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
	err = core.LoadUserGroups(db, u)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load user groups", err)
	}

	// Compute list of group ids we can access
	groupMap := make(map[int64]*core.Group)
	err = core.LoadGroupNames(db, u, groupMap)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load groupMap", err)
	}
	groupIds := make([]string, 0)
	groupIds = append(groupIds, "0")
	for id, _ := range groupMap {
		groupIds = append(groupIds, fmt.Sprintf("%d", id))
	}

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

	// Perform search.
	results, err := elastic.SearchPageIndex(db.C, jsonStr)
	if err != nil {
		return pages.HandlerErrorFail("Error with elastic search", err)
	}

	// Create page map.
	pageMap := make(map[int64]*core.Page)
	for _, hit := range results.Hits.Hits {
		pageMap[hit.Id] = &core.Page{PageId: hit.Id}
	}

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, &core.LoadPageOptions{})
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Load auxillary data.
	err = core.LoadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return pages.HandlerErrorFail("error while loading aux data", err)
	}

	returnData := createReturnData(pageMap).AddResult(results.Hits)
	return pages.StatusOK(returnData)
}
