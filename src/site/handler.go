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

	"github.com/gorilla/mux"
)

// siteHandler is the wrapper object for handler functions
type siteHandler struct {
	URI         string
	HandlerFunc func(*pages.HandlerParams) *pages.Result
	Options     pages.PageOptions
}

// handlerWrapper wraps our siteHandler to provide standard http handler interface.
func handlerWrapper(h siteHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If live, check that this is an HTTPS request
		if sessions.Live && r.URL.Scheme != "https" {
			safeUrl := strings.Replace(r.URL.String(), "http", "https", 1)
			http.Redirect(w, r, safeUrl, http.StatusSeeOther)
		}

		c := sessions.NewContext(r)
		rand.Seed(time.Now().UnixNano())

		fail := func(responseCode int, message string, err error) {
			c.Inc(fmt.Sprintf("%s-fail", r.URL.Path))
			c.Errorf("handlerWrapper: %s: %v", message, err)
			w.WriteHeader(responseCode)
			fmt.Fprintf(w, "%s", message)
		}

		// Recover from panic
		defer func() {
			if sessions.Live {
				if r := recover(); r != nil {
					c.Errorf("%v", r)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", "Super serious error has occurred. Super. Serious. Error.")
				}
			}
		}()

		// Open DB connection
		db, err := database.GetDB(c)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't open DB", err)
			return
		}

		// Create params
		params := &pages.HandlerParams{W: w, R: r, C: c, DB: db}
		u, err := core.LoadCurrentUser(w, r, db)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load user", err)
			return
		}
		params.U = u
		params.PrivateGroupId, err = loadSubdomain(r, db, u)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load subdomain", err)
			return
		}

		// Load the user's trust
		err = core.LoadUserTrust(db, u)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't retrieve user trust", err)
			return
		}

		// Check permissions
		if h.Options.RequireLogin && !core.IsIdValid(u.Id) {
			fail(http.StatusInternalServerError, "Have to be logged in", nil)
			return
		}
		if h.Options.RequireTrusted && !u.IsTrusted {
			fail(http.StatusInternalServerError, "Have to be a trusted user", nil)
			return
		}
		if h.Options.AdminOnly && !u.IsAdmin {
			fail(http.StatusInternalServerError, "Have to be an admin", nil)
			return
		}

		// Check if we have access to the private group
		if core.IsIdValid(params.PrivateGroupId) {
			if !h.Options.AllowAnyone && !u.IsMemberOfGroup(params.PrivateGroupId) {
				fail(http.StatusForbidden, "Don't have access to this group", nil)
				return
			}
			// We don't allow personal private groups for now
			if params.PrivateGroupId == u.Id {
				fail(http.StatusForbidden, "Arbital no longer supports personal private groups", nil)
			}
		}

		result := h.HandlerFunc(params)
		if result.Err != nil {
			fail(result.ResponseCode, result.Err.Message, result.Err.Err)
			return
		}

		if result.Data != nil {
			if core.IsIdValid(u.Id) && result.Data.(*core.CommonHandlerData).ResetEverything {
				// Load updates count. (Loading it afterwards since it could be affected by the page)
				u.UpdateCount, err = core.LoadUpdateCount(db, u.Id)
				u.NewAchievementCount, err = core.LoadNewAchievementCount(db, u)
				u.MaintenanceUpdateCount, err = core.LoadMaintenanceUpdateCount(db, u.Id, false)
				if err != nil {
					fail(http.StatusInternalServerError, "Couldn't retrieve updates count", err)
					return
				}
			}

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
		c.Inc(fmt.Sprintf("%s-success", r.URL.Path))
	}
}

// loadSubdomain loads the id for the private group corresponding to the private group id.
func loadSubdomain(r *http.Request, db *database.DB, u *core.CurrentUser) (string, error) {
	subdomain := strings.ToLower(mux.Vars(r)["subdomain"])
	if subdomain == "" {
		return "", nil
	}
	// Get actual page id for the group
	privateGroupId, ok, err := core.LoadAliasToPageId(db, u, subdomain)
	if err != nil {
		return "", fmt.Errorf("Couldn't convert subdomain to id: %v", err)
	}
	if !ok {
		return "", fmt.Errorf("Couldn't find private group %s", subdomain)
	}
	return privateGroupId, nil
}
