// parentsSearchJsonHandler.go contains the handler for matching a partial query against
// pages' ids, aliases, and titles.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/elastic"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// parentsSearchJsonHandler handles the request.
func parentsSearchJsonHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	c := sessions.NewContext(r)

	// Decode data
	var data searchJsonData
	q := r.URL.Query()
	data.Term = q.Get("term")
	if data.Term != "" {
		err = parentsSearchJsonInternalHandler(w, r, &data)
	} else {
		err = fmt.Errorf("No search term specified")
	}

	if err != nil {
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func parentsSearchJsonInternalHandler(w http.ResponseWriter, r *http.Request, data *searchJsonData) error {
	c := sessions.NewContext(r)

	// Load user object
	u, err := user.LoadUser(w, r)
	if err != nil {
		return fmt.Errorf("Couldn't load user: %v", err)
	}

	// Load user grups
	err = loadUserGroups(c, u)
	if err != nil {
		return fmt.Errorf("Couldn't load user groups: %v", err)
	}

	// Compute list of group ids we can access
	groupMap := make(map[int64]*core.Group)
	err = loadGroupNames(c, u, groupMap)
	if err != nil {
		return fmt.Errorf("Couldn't load groupMap: %v", err)
	}
	groupIds := make([]string, 0)
	groupIds = append(groupIds, "0")
	for id, _ := range groupMap {
		groupIds = append(groupIds, fmt.Sprintf("%d", id))
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
		"_source": ["pageId", "alias", "title"]
	}`, data.Term, strings.Join(groupIds, ","))

	// Perform search.
	results, err := elastic.SearchPageIndex(c, jsonStr)
	if err != nil {
		return fmt.Errorf("Error with elastic search: %v", err)
	}

	// Return the pages in JSON format.
	jsonData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("Couldn't write json: %v", err)
	}
	w.Write(jsonData)
	return nil
}
