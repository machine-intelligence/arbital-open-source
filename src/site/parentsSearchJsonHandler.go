// parentsSearchJsonHandler.go contains the handler for matching a partial query against
// pages' ids, aliases, and titles.
package site

import (
	"fmt"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

// parentsSearchJsonHandler handles the request.
func parentsSearchJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data searchJsonData
	q := params.R.URL.Query()
	data.Term = q.Get("term")
	if data.Term == "" {
		return pages.HandlerBadRequestFail("No search term specified", nil)
	}
	result, message, err := parentsSearchJsonInternalHandler(params, &data)

	if result == nil {
		return pages.HandlerErrorFail(message, err)
	}
	return pages.StatusOK(result)
}

func parentsSearchJsonInternalHandler(params *pages.HandlerParams, data *searchJsonData) (map[string]interface{}, string, error) {
	db := params.DB
	u := params.U

	// Load user grups
	err := loadUserGroups(db, u)
	if err != nil {
		return nil, "Couldn't load user groups", err
	}

	// Compute list of group ids we can access
	groupMap := make(map[int64]*core.Group)
	err = loadGroupNames(db, u, groupMap)
	if err != nil {
		return nil, "Couldn't load groupMap", err
	}
	groupIds := make([]string, 0)
	groupIds = append(groupIds, "0")
	for id, _ := range groupMap {
		groupIds = append(groupIds, fmt.Sprintf("%d", id))
	}

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
								"terms": { "groupId": [%[2]s] }
							}
						]
					}
				}
			}
		},
		"_source": []
	}`, escapedTerm, strings.Join(groupIds, ","))

	// Perform search.
	results, err := elastic.SearchPageIndex(params.C, jsonStr)
	if err != nil {
		return nil, "Error with elastic search", err
	}

	// Create page map.
	pageMap := make(map[int64]*core.Page)
	for _, hit := range results.Hits.Hits {
		pageMap[hit.Id] = &core.Page{PageId: hit.Id}
	}

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, &core.LoadPageOptions{})
	if err != nil {
		return nil, "error while loading pages", err
	}

	// Load auxillary data.
	err = loadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return nil, "error while loading aux data", err
	}

	// Return the data in JSON format.
	returnPageData := make(map[string]*core.Page)
	for k, v := range pageMap {
		returnPageData[fmt.Sprintf("%d", k)] = v
	}

	returnData := make(map[string]interface{})
	returnData["searchHits"] = results.Hits
	returnData["pages"] = returnPageData
	return returnData, "", nil
}
