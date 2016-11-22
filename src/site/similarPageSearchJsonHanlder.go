// similarPageSearchJsonHandler.go contains the handler for searching for a
// page that's similar to the one the user is creating.

package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

type similarPageSearchJSONData struct {
	Title     string
	Clickbait string
	Text      string
	PageType  string
}

var similarPageSearchHandler = siteHandler{
	URI:         "/json/similarPageSearch/",
	HandlerFunc: similarPageSearchJSONHandler,
}

// similarPageSearchJsonHandler handles the request.
func similarPageSearchJSONHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data similarPageSearchJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Error decoding JSON", err)
	}
	if len(data.Title) < 3 && len(data.Clickbait) < 3 && len(data.Text) < 3 {
		return pages.Success(nil)
	}

	domainIDs := []string{"\"\"", "\"" + params.PrivateDomain.ID + "\""}
	escapedTitle := elastic.EscapeMatchTerm(data.Title)
	escapedClickbait := elastic.EscapeMatchTerm(data.Clickbait)
	escapedText := elastic.EscapeMatchTerm(data.Text)

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
							}
						]
					}
				},
				"filter": {
					"bool": {
						"must": [
							{
								"terms": { "seeDomainId": [%s] }
							}
						]
					}
				}
			}
		},
		"_source": []
	}`, minSearchScore, searchSize, escapedTitle, escapedClickbait, escapedText, strings.Join(domainIDs, ","))
	return searchJSONInternalHandler(params, jsonStr)
}
