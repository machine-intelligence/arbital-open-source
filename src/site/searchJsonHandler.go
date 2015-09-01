// searchJsonHandler.go contains the handler for returning JSON with all pages'
// search and titles.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine/search"

	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

type resultPair struct {
	Label tasks.PageIndexDoc `json:"label"`
	Value string             `json:"value"`
}

// searchJsonData contains parameters passed in via the request.
type searchJsonData struct {
	Term string
}

// aliasesJsonHandler handles the request.
func searchJsonHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	c := sessions.NewContext(r)

	// Decode data
	var data searchJsonData
	q := r.URL.Query()
	data.Term = q.Get("term")
	if data.Term != "" {
		err = searchJsonInternalHandler(w, r, &data)
	} else {
		err = fmt.Errorf("No search term specified")
	}

	if err != nil {
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func searchJsonInternalHandler(w http.ResponseWriter, r *http.Request, data *searchJsonData) error {
	c := sessions.NewContext(r)

	// Load user object
	u, err := user.LoadUser(w, r)
	if err != nil {
		return fmt.Errorf("Couldn't load user: %v", err)
	}

	groupMap := make(map[string]bool)
	groupMap["0"] = true
	groupMap[""] = true
	if u.Id > 0 {
		if err = loadUserGroups(c, u); err != nil {
			return fmt.Errorf("Couldn't load user: %v", err)
		}
		for _, groupId := range u.GroupIds {
			groupMap[groupId] = true
		}
	}

	results := make([]resultPair, 0)
	index, err := search.Open("pages")
	if err != nil {
		return fmt.Errorf("Failed to open pages index: %v", err)
	}
	options := &search.SearchOptions{Limit: 20, Fields: []string{"PageId", "Alias", "Title", "GroupId"}}
	for t := index.Search(c, data.Term, options); ; {
		var pair resultPair
		pair.Value, err = t.Next(&pair.Label)
		if err == search.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Search error: %v", err)
		}
		// TODO: instead of filtering out results the user isn't supposed to see,
		// we should modify the query to only return results from accessible groups.
		if _, ok := groupMap[string(pair.Label.GroupId)]; !ok {
			continue
		}
		pair.Label.Text = ""
		results = append(results, pair)
	}

	// Return the pages in JSON format.
	jsonData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("Couldn't write json: %v", err)
	}
	w.Write(jsonData)
	return nil
}
