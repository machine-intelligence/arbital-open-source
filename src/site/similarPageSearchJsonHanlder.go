// similarPageSearchJsonHandler.go contains the handler for searching for a
// page that's similar to the one the user is creating.
package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

type similarPageSearchJsonData struct {
	Title     string
	Clickbait string
	Text      string
}

// similarPageSearchJsonHandler handles the request.
func similarPageSearchJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data similarPageSearchJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerErrorFail("Error decoding JSON", err)
	}
	if len(data.Title) < 3 && len(data.Clickbait) < 3 && len(data.Text) < 3 {
		return pages.StatusOK(nil)
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

	escapedTitle := elastic.EscapeMatchTerm(data.Title)
	escapedClickbait := elastic.EscapeMatchTerm(data.Clickbait)
	escapedText := elastic.EscapeMatchTerm(data.Text)

	// Construct the search JSON
	jsonStr := fmt.Sprintf(`{
		"query": {
			"filtered": {
				"query": {
					"bool": {
						"should": [
							{
								"match": { "title": "%[1]s" }
							},
							{
								"match": { "clickbait": "%[2]s" }
							},
							{
								"match": { "text": "%[3]s" }
							}
						]
					}
				},
				"filter": {
					"bool": {
						"must": [
							{
								"term": { "type": "question" }
							}, {
								"terms": { "seeGroupId": [%[4]s] }
							}
						]
					}
				}
			}
		},
		"_source": ["pageId", "alias", "title", "clickbait", "seeGroupId"]
	}`, escapedTitle, escapedClickbait, escapedText, strings.Join(groupIds, ","))

	// Perform search.
	results, err := elastic.SearchPageIndex(params.C, jsonStr)
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

	returnData := createReturnData(pageMap).AddSearchHits(results.Hits)
	return pages.StatusOK(returnData)
}
