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

		params := &pages.HandlerParams{W: w, R: r, C: c, DB: db}
		params.PrivateGroupId, err = loadSubdomain(r, db)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load subdomain", err)
			return
		}

		// Get user object
		u, err := user.LoadUser(w, r, db)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load user", err)
			return
		}
		params.U = u

		// Check permissions
		if h.Options.RequireLogin && !core.IsIdValid(u.Id) {
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

		if core.IsIdValid(u.Id) {
			// Load the groups the user belongs to.
			if err = core.LoadUserGroupIds(db, u); err != nil {
				fail(http.StatusInternalServerError, "Couldn't load user groups", err)
				return
			}
		}
		// Check if we have access to the private group
		if !h.Options.AllowAnyone && core.IsIdValid(params.PrivateGroupId) && !u.IsMemberOfGroup(params.PrivateGroupId) {
			fail(http.StatusForbidden, "Don't have access to this group", nil)
			return
		}

		result := h.HandlerFunc(params)
		if result.ResponseCode != http.StatusOK && result.ResponseCode != http.StatusSeeOther {
			fail(result.ResponseCode, result.Message, result.Err)
			return
		}

		if core.IsIdValid(u.Id) && h.Options.LoadUpdateCount {
			// Load updates count. (Loading it afterwards since it could be affected by the page)
			u.UpdateCount, err = core.LoadUpdateCount(db, u.Id)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't retrieve updates count", err)
				return
			}
		}

		if core.IsIdValid(u.Id) && h.Options.LoadUserTrust {
			// Load the user's trust
			u.TrustMap, err = user.LoadUserTrust(db, u.Id)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't retrieve user trust", err)
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
		c.Inc(fmt.Sprintf("%s-success", r.URL.Path))
	}
}

// loadSubdomain loads the id for the private group corresponding to the private group id.
func loadSubdomain(r *http.Request, db *database.DB) (string, error) {
	subdomain := strings.ToLower(mux.Vars(r)["subdomain"])
	if subdomain == "" {
		return "", nil
	}
	// Get actual page id for the group
	privateGroupId, ok, err := core.LoadAliasToPageId(db, subdomain)
	if err != nil {
		return "", fmt.Errorf("Couldn't convert subdomain to id: %v", err)
	}
	if !ok {
		return "", fmt.Errorf("Couldn't find private group %s", subdomain)
	}
	return privateGroupId, nil
}
