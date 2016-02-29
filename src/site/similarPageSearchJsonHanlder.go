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
	PageType  string
}

var similarPageSearchHandler = siteHandler{
	URI:         "/json/similarPageSearch/",
	HandlerFunc: similarPageSearchJsonHandler,
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
	err = core.LoadUserGroupIds(db, u)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load user groups", err)
	}

	groupIds := []string{"\"\"", "\"" + params.PrivateGroupId + "\""}
	escapedTitle := elastic.EscapeMatchTerm(data.Title)
	escapedClickbait := elastic.EscapeMatchTerm(data.Clickbait)
	escapedText := elastic.EscapeMatchTerm(data.Text)
	escapedPageType := elastic.EscapeMatchTerm(strings.ToLower(data.PageType))

	// Construct the search JSON
	jsonStr := fmt.Sprintf(`{
		"min_score": %v,
		"size": %d,
		"query": {
			"filtered": {
				"query": {
					"bool": {
						"should": [
							{
								"match": { "title": "%s" }
							},
							{
								"match": { "clickbait": "%s" }
							},
							{
								"match": { "text": "%s" }
							},
							{
								"match": { "type": "%s" }
							}
						]
					}
				},
				"filter": {
					"bool": {
						"must": [
							{
								"terms": { "seeGroupId": [%s] }
							}
						]
					}
				}
			}
		},
		"_source": []
	}`, minSearchScore, searchSize, escapedTitle, escapedClickbait, escapedText, escapedPageType, strings.Join(groupIds, ","))
	return searchJsonInternalHandler(params, jsonStr)
}
