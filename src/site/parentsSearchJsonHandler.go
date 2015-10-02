// parentsSearchJsonHandler.go contains the handler for matching a partial query against
// pages' ids, aliases, and titles.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
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

	db, err := database.GetDB(c)
	if err != nil {
		return err
	}

	// Load user object
	u, err := user.LoadUser(w, r, db)
	if err != nil {
		return fmt.Errorf("Couldn't load user: %v", err)
	}

	// Load user grups
	err = loadUserGroups(db, u)
	if err != nil {
		return fmt.Errorf("Couldn't load user groups: %v", err)
	}

	// Compute list of group ids we can access
	groupMap := make(map[int64]*core.Group)
	err = loadGroupNames(db, u, groupMap)
	if err != nil {
		return fmt.Errorf("Couldn't load groupMap: %v", err)
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
	results, err := elastic.SearchPageIndex(c, jsonStr)
	if err != nil {
		return fmt.Errorf("Error with elastic search: %v", err)
	}

	// Create page map.
	pageMap := make(map[int64]*core.Page)
	for _, hit := range results.Hits.Hits {
		pageMap[hit.Id] = &core.Page{PageId: hit.Id}
	}

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, &core.LoadPageOptions{})
	if err != nil {
		return fmt.Errorf("error while loading pages: %v", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return fmt.Errorf("error while loading aux data: %v", err)
	}

	// Return the data in JSON format.
	returnPageData := make(map[string]*core.Page)
	for k, v := range pageMap {
		returnPageData[fmt.Sprintf("%d", k)] = v
	}

	returnData := make(map[string]interface{})
	returnData["searchHits"] = results.Hits
	returnData["pages"] = returnPageData

	// Return the pages in JSON format.
	jsonData, err := json.Marshal(returnData)
	if err != nil {
		return fmt.Errorf("Couldn't write json: %v", err)
	}
	w.Write(jsonData)
	return nil
}
