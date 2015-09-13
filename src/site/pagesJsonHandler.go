// pagesJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"database/sql"
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
	PageAliases []string
	// Load entire page text
	IncludeText bool
	// Load auxillary data: likes, votes, subscription
	IncludeAuxData bool
	// If true, at most one page id can be passed. We'll load the most recent version
	// of the page, even if it's a draft.
	AllowDraft     bool
	LoadComments   bool
	LoadVotes      bool
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
	if len(data.PageAliases) <= 0 {
		rand.Seed(time.Now().UnixNano())
		pageId := rand.Int63()
		returnPageData := make(map[string]*core.Page)
		returnPageData[fmt.Sprintf("%d", pageId)] = &core.Page{PageId: pageId}
		returnData["pages"] = returnPageData
		return returnData, nil
	}

	// Convert all aliases to ids
	pageIds := make([]int64, 0)
	strAliases := make([]string, 0)
	for _, alias := range data.PageAliases {
		pageId, err := strconv.ParseInt(alias, 10, 64)
		if err == nil {
			pageIds = append(pageIds, pageId)
		} else {
			strAliases = append(strAliases, fmt.Sprintf(`"%s"`, alias))
		}
	}

	// Convert actual aliases into page ids
	if len(strAliases) > 0 {
		query := fmt.Sprintf(`
			SELECT pageId
			FROM aliases
			WHERE fullName IN (%s)`, strings.Join(strAliases, ","))
		err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan for original createdAt: %v", err)
			}
			pageIds = append(pageIds, pageId)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("couldn't convert aliases to page ids: %v", err)
		}
	}
	if len(pageIds) <= 0 {
		return nil, fmt.Errorf("All of the passed in aliases weren't found.")
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
		for _, pageId := range pageIds {
			pageMap[pageId] = &core.Page{PageId: pageId}
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

		err = core.LoadPages(c, pageMap, u.Id, &core.LoadPageOptions{LoadText: true, LoadSummary: true})
		if err != nil {
			return nil, fmt.Errorf("error while loading pages: %v", err)
		}
	} else {
		pageId := pageIds[0]

		// Load full edit for one page.
		p, err := loadFullEdit(c, pageId, u.Id, nil)
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
	}

	// Load probability votes
	if data.LoadVotes {
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
