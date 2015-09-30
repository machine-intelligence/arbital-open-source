// similarPageSearchJsonHandler.go contains the handler for searching for a
// page that's similar to the one the user is creating.
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

type similarPageSearchJsonData struct {
	Title     string
	Clickbait string
	Text      string
}

// similarPageSearchJsonHandler handles the request.
func similarPageSearchJsonHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	c := sessions.NewContext(r)

	// Decode data
	var data similarPageSearchJsonData
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&data)
	if err != nil {
		c.Errorf("Error decoding JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(data.Title) > 2 || len(data.Clickbait) > 2 || len(data.Text) > 2 {
		err = similarPageSearchJsonInternalHandler(w, r, &data)
	}

	if err != nil {
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func similarPageSearchJsonInternalHandler(w http.ResponseWriter, r *http.Request, data *similarPageSearchJsonData) error {
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

	// Load user groups
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
								"terms": { "groupId": [%[4]s] }
							}
						]
					}
				}
			}
		},
		"_source": ["pageId", "alias", "title", "clickbait", "groupId"]
	}`, data.Title, data.Clickbait, data.Text, strings.Join(groupIds, ","))

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
