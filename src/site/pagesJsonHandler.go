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
	// If true, we expect only one pageId, but will load the last edit, just like
	// for the edit page.
	LoadFullEdit bool
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

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load data
	pageMap := make(map[int64]*page)
	if !data.LoadFullEdit {
		// Process pageIds
		pageIds := strings.Split(strings.Trim(data.PageIds, ","), ",")
		for _, id := range pageIds {
			pageId, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				c.Errorf("Couldn't parse page id: %v", pageId)
			} else {
				pageMap[pageId] = &page{PageId: pageId}
			}
		}

		err = loadPages(c, pageMap, u.Id, loadPageOptions{loadText: true})
		if err != nil {
			c.Inc("pages_handler_fail")
			c.Errorf("error while loading pages: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		pageId, err := strconv.ParseInt(strings.Trim(data.PageIds, ","), 10, 64)
		if err != nil {
			c.Errorf("Couldn't parse page id: %v", data.PageIds)
		}

		// Load full edit for one page.
		p, err := loadFullEdit(c, pageId, u.Id)
		if err != nil || p == nil {
			c.Inc("pages_handler_fail")
			c.Errorf("error while loading full edit: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pageMap[pageId] = p
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
