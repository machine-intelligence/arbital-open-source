// pagesJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/schema"
)

// pagesJsonData contains parameters passed in via the request.
type pagesJsonData struct {
	PageIds string // comma separated string of page ids
}

// pagesJsonHandler handles the request.
func pagesJsonHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	returnData := make(map[string]interface{})

	// Decode data
	var data pagesJsonData
	r.ParseForm()
	err := schema.NewDecoder().Decode(&data, r.Form)
	if err != nil {
		c.Errorf("Couldn't decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If no page ids, return a new random page id.
	if len(data.PageIds) <= 0 {
		rand.Seed(time.Now().UnixNano())
		pageId := rand.Int63()
		returnData[fmt.Sprintf("%d", pageId)] = &page{PageId: pageId}

		err = writeJson(w, returnData)
		if err != nil {
			c.Inc("pages_handler_fail")
			c.Errorf("Couldn't write json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Process pageIds
	pageMap := make(map[int64]*page)
	pageIds := strings.Split(data.PageIds, ",")
	for _, id := range pageIds {
		pageId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			pageMap[pageId] = &page{PageId: pageId}
		}
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load pages.
	err = loadPages(c, pageMap, u.Id, loadPageOptions{})
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("error while loading pages: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the pages in JSON format.
	for k, v := range pageMap {
		returnData[fmt.Sprintf("%d", k)] = v
	}
	err = writeJson(w, returnData)
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("Couldn't write json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
