// handler.go: Logic for modifying HTTP handlers.
package site

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

// siteHandler is the wrapper object for handler functions
type siteHandler struct {
	URI         string
	HandlerFunc func(*pages.HandlerParams) *pages.Result
	Options     pages.PageOptions
}

// commonHandlerData is what handlers fill out and return
type commonHandlerData struct {
	// If set, then this packet should reset everything on the FE
	ResetEverything bool
	// Optional user object with the current user's data
	User *user.User
	// Map of page id -> currently live version of the page
	PageMap map[int64]*core.Page
	// Map of page id -> some edit of the page
	EditMap    map[int64]*core.Page
	UserMap    map[int64]*core.User
	MasteryMap map[int64]*core.Mastery
	// ResultMap contains various data the specific handler returns
	ResultMap map[string]interface{}
}

// handlerWrapper wraps our siteHandler to provide standard http handler interface.
func handlerWrapper(h siteHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := sessions.NewContext(r)
		startTime := time.Now()
		rand.Seed(time.Now().UnixNano())

		fail := func(responseCode int, message string, err error) {
			c.Inc(fmt.Sprintf("%s-fail", r.URL.Path))
			c.Errorf("handlerWrapper: %s: %v", message, err)
			w.WriteHeader(responseCode)
			fmt.Fprintf(w, "%s", message)
		}

		// Recover from panic.
		defer func() {
			if sessions.Live {
				if r := recover(); r != nil {
					c.Errorf("%v", r)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", "Super serious error has occured. Super. Serious. Error.")
				}
			}
		}()

		// Open DB connection
		db, err := database.GetDB(c)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't open DB", err)
			return
		}

		params := &pages.HandlerParams{W: w, R: r, C: c, DB: db}
		params.PrivateGroupId, err = loadSubdomain(r, db)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load subdomain", err)
			return
		}

		// Get user object
		var u *user.User
		if !h.Options.SkipLoadingUser {
			u, err = user.LoadUser(w, r, db)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't load user", err)
				return
			}
			params.U = u

			// Check permissions
			if h.Options.RequireLogin && !u.IsLoggedIn {
				fail(http.StatusInternalServerError, "Have to be logged in", nil)
				return
			}
			if h.Options.AdminOnly && !u.IsAdmin {
				fail(http.StatusInternalServerError, "Have to be an admin", nil)
				return
			}
			if h.Options.MinKarma != 0 && u.Karma < h.Options.MinKarma {
				fail(http.StatusInternalServerError, "Not enough karma", nil)
				return
			}

			if u.Id > 0 {
				// Load the groups the user belongs to.
				if err = core.LoadUserGroupIds(db, u); err != nil {
					fail(http.StatusInternalServerError, "Couldn't load user groups", err)
					return
				}
			}
			// Check if we have access to the private group
			if params.PrivateGroupId > 0 && !u.IsMemberOfGroup(params.PrivateGroupId) {
				fail(http.StatusForbidden, "Don't have access to this group", nil)
				return
			}
		}

		result := h.HandlerFunc(params)
		if result.ResponseCode != http.StatusOK && result.ResponseCode != http.StatusSeeOther {
			fail(result.ResponseCode, result.Message, result.Err)
			return
		}

		if u.Id > 0 && h.Options.LoadUpdateCount {
			// Load updates count. (Loading it afterwards since it could be affected by the page)
			u.UpdateCount, err = core.LoadUpdateCount(db, u.Id)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't retrieve updates count", err)
				return
			}
		}

		if result.Data != nil {
			w.Header().Set("Content-type", "application/json")
			// Return the pages in JSON format.
			jsonData, err := json.Marshal(result.Data)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't marshal json", err)
				return
			}
			_, err = w.Write(jsonData)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't write json", err)
				return
			}
		}
		c.Debugf("Time spent: %s", time.Since(startTime).String())
		c.Inc(fmt.Sprintf("%s-success", r.URL.Path))
	}
}

// newHandlerData creates and initializes a new commonHandlerData object.
func newHandlerData(resetEverything bool) *commonHandlerData {
	var data commonHandlerData
	data.ResetEverything = resetEverything
	data.PageMap = make(map[int64]*core.Page)
	data.EditMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	data.MasteryMap = make(map[int64]*core.Mastery)
	data.ResultMap = make(map[string]interface{})
	return &data
}

// toJson puts together the data into one "json" object, so we
// can send it to the front-end.
func (data *commonHandlerData) toJson() map[string]interface{} {
	jsonData := make(map[string]interface{})

	jsonData["resetEverything"] = data.ResetEverything

	if data.User != nil {
		jsonData["user"] = data.User
	}

	returnPageData := make(map[string]*core.Page)
	for k, v := range data.PageMap {
		returnPageData[fmt.Sprintf("%d", k)] = v
	}
	jsonData["pages"] = returnPageData

	returnEditData := make(map[string]*core.Page)
	for k, v := range data.EditMap {
		returnEditData[fmt.Sprintf("%d", k)] = v
	}
	jsonData["edits"] = returnEditData

	returnUserData := make(map[string]*core.User)
	for k, v := range data.UserMap {
		returnUserData[fmt.Sprintf("%d", k)] = v
	}
	jsonData["users"] = returnUserData

	returnMasteryData := make(map[string]*core.Mastery)
	for k, v := range data.MasteryMap {
		returnMasteryData[fmt.Sprintf("%d", k)] = v
	}
	jsonData["masteries"] = returnMasteryData

	jsonData["result"] = data.ResultMap
	return jsonData
}

// loadSubdomain loads the id for the private group corresponding to the private group id.
func loadSubdomain(r *http.Request, db *database.DB) (int64, error) {
	subdomain := strings.ToLower(mux.Vars(r)["subdomain"])
	if subdomain == "" {
		return 0, nil
	}
	// Get actual page id for the group
	privateGroupId, ok, err := core.LoadAliasToPageId(db, subdomain)
	if err != nil {
		return 0, fmt.Errorf("Couldn't convert subdomain to id: %v", err)
	}
	if !ok {
		return 0, fmt.Errorf("Couldn't find private group %s", subdomain)
	}
	return privateGroupId, nil
}
