// pagesJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/schema"
)

// pagesJsonData contains parameters passed in via the request.
type pagesJsonData struct {
	PageIds string // comma separated string of page ids
	// Load entire page text
	IncludeText bool
	// Load auxillary data: likes, votes, subscription
	IncludeAuxData bool
	// If true, at most one page id can be passed. We'll load the most recent version
	// of the page, even if it's a draft.
	AllowDraft     bool
	LoadComments   bool
	LoadChildren   bool
	LoadChildDraft bool
}

// pagesJsonHandler handles the request.
func pagesJsonHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Decode data
	var data pagesJsonData
	r.ParseForm()
	err := schema.NewDecoder().Decode(&data, r.Form)
	if err != nil {
		c.Inc("pages_json_handler_fail")
		c.Errorf("Couldn't decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	returnData, err := pagesJsonHandlerInternal(w, r, &data)
	if err != nil {
		c.Inc("pages_json_handler_fail")
		c.Errorf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write JSON response
	err = writeJson(w, returnData)
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("Couldn't write json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// pagesJsonHandler handles the request.
func pagesJsonHandlerInternal(w http.ResponseWriter, r *http.Request, data *pagesJsonData) (map[string]interface{}, error) {
	c := sessions.NewContext(r)
	returnData := make(map[string]interface{})

	// If no page ids, return a new random page id.
	if len(data.PageIds) <= 0 {
		rand.Seed(time.Now().UnixNano())
		pageId := rand.Int63()
		returnPageData := make(map[string]*core.Page)
		returnPageData[fmt.Sprintf("%d", pageId)] = &core.Page{PageId: pageId}
		returnData["pages"] = returnPageData
		return returnData, nil
	}

	// Load user object
	u, err := user.LoadUser(w, r)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load user: %v", err)
	}

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)
	if !data.AllowDraft {
		// Process pageIds
		pageIds := strings.Split(strings.Trim(data.PageIds, ","), ",")
		for _, id := range pageIds {
			pageId, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				c.Errorf("Couldn't parse page id: %v", pageId)
			} else {
				pageMap[pageId] = &core.Page{PageId: pageId}
			}
		}

		// Load comment ids.
		if data.LoadComments {
			err = loadCommentIds(c, pageMap, pageMap)
			if err != nil {
				return nil, fmt.Errorf("Couldn't load comments: %v", err)
			}
		}

		// Load children
		if data.LoadChildren {
			err = loadChildrenIds(c, pageMap, loadChildrenIdsOptions{})
			if err != nil {
				return nil, fmt.Errorf("Couldn't load children: %v", err)
			}
		}

		err = core.LoadPages(c, pageMap, u.Id, core.LoadPageOptions{LoadText: true})
		if err != nil {
			return nil, fmt.Errorf("error while loading pages: %v", err)
		}
	} else {
		pageId, err := strconv.ParseInt(strings.Trim(data.PageIds, ","), 10, 64)
		if err != nil {
			c.Errorf("Couldn't parse page id: %v", data.PageIds)
		}

		// Load full edit for one page.
		p, err := loadFullEdit(c, pageId, u.Id)
		if err != nil || p == nil {
			return nil, fmt.Errorf("error while loading full edit: %v", err)
		}
		pageMap[pageId] = p
	}

	// Load the auxillary data.
	if data.IncludeAuxData {
		err = loadAuxPageData(c, u.Id, pageMap, nil)
		if err != nil {
			return nil, fmt.Errorf("error while loading aux data: %v", err)
		}

		// Load probability votes
		err = loadVotes(c, u.Id, core.PageIdsStringFromMap(pageMap), pageMap, userMap)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load probability votes: %v", err)
		}
	}

	// Load links
	err = loadLinks(c, pageMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load links: %v", err)
	}

	if data.LoadChildDraft {
		// Load child draft
		for _, p := range pageMap {
			if p.Type == core.CommentPageType {
				continue
			}
			err = loadChildDraft(c, u.Id, p, pageMap)
			if err != nil {
				return nil, fmt.Errorf("Couldn't load child draft: %v", err)
			}
			break
		}
	}

	// Load all the users
	for _, p := range pageMap {
		userMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, userMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading users: %v", err)
	}

	// Return the data in JSON format.
	visitedValues := ""
	returnPageData := make(map[string]*core.Page)
	for k, v := range pageMap {
		if !data.IncludeText {
			v.Text = ""
		} else {
			visitedValues += fmt.Sprintf("(%d, %d, '%s'),",
				u.Id, k, database.Now())
		}
		returnPageData[fmt.Sprintf("%d", k)] = v
	}
	returnUserData := make(map[string]*core.User)
	for k, v := range userMap {
		returnUserData[fmt.Sprintf("%d", k)] = v
	}
	returnData["pages"] = returnPageData
	returnData["users"] = returnUserData

	// Add a visit to pages for which we loaded text.
	visitedValues = strings.TrimRight(visitedValues, ",")
	query := fmt.Sprintf(`
		INSERT INTO visits (userId, pageId, createdAt)
		VALUES %s`, visitedValues)
	database.ExecuteSql(c, query)

	return returnData, nil
}
