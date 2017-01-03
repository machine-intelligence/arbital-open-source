// pageHandler.go has the functions and wrappers for handling pages

package site

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"

	"zanaduu3/vendor/github.com/gorilla/mux"
	"zanaduu3/vendor/github.com/imdario/mergo"
	"zanaduu3/vendor/google.golang.org/appengine"
)

type dynamicPageTmplData struct {
	Title             string
	URL               string
	Description       string
	VersionID         string
	IsLive            bool
	MaybeServerPrefix string
}

// pageHandlerWrapper wraps one of our page handlers.
func pageHandlerWrapper(p *pages.Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If live, check that this is an HTTPS request
		if sessions.Live && r.URL.Scheme != "https" {
			safeURL := strings.Replace(r.URL.String(), "http", "https", 1)
			http.Redirect(w, r, safeURL, http.StatusSeeOther)
			return
		}

		c := sessions.NewContext(r)
		params := pages.HandlerParams{W: w, R: r, C: c}

		// Redirect www.
		if mux.Vars(r)["www"] != "" {
			if sessions.Live {
				newURL := strings.Replace(r.URL.String(), "www.", "", -1)
				c.Debugf("Redirecting '%s' to '%s' because of 'www'", r.URL.String(), newURL)
				http.Redirect(w, r, newURL, http.StatusSeeOther)
			} else {
				subdomainStr := ""
				if mux.Vars(r)["subdomain"] != "" {
					subdomainStr = mux.Vars(r)["subdomain"] + "."
				}
				url := fmt.Sprintf("http://%s%s%s", subdomainStr, sessions.GetRawDomain(), r.URL.String())
				http.Redirect(w, r, url, http.StatusSeeOther)
			}
			return
		}

		// Helper func to when an error occurs and we should render error page.
		fail := func(responseCode int, message string, err error) {
			c.Inc(fmt.Sprintf("%s-fail", r.URL.Path))
			c.Errorf("%s: %v", message, err)
			w.WriteHeader(responseCode)
			fmt.Fprintf(w, "Error rendering the page: %s", message)
		}

		// Recover from panic.
		if sessions.Live {
			defer func() {
				if r := recover(); r != nil {
					c.Errorf("%v", r)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", "Super serious error has occurred. Super. Serious. Error.")
				}
			}()
		}

		// Open DB connection
		db, err := database.GetDB(c)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't open DB", err)
			return
		}
		params.DB = db

		// Get user object
		u, err := core.LoadCurrentUser(w, r, db)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load user", err)
			return
		}
		params.U = u

		// Get subdomain info
		params.PrivateDomain, err = loadSubdomain(r, db, u)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load subdomain", err)
			return
		}

		// When in a subdomain, we always have to be logged in
		if core.IsIntIDValid(params.PrivateDomain.ID) && !core.IsIDValid(u.ID) {
			if r.URL.Path != "/login/" {
				http.Redirect(w, r, fmt.Sprintf("/login/?continueUrl=%s", url.QueryEscape(r.URL.String())), http.StatusSeeOther)
				return
			}
		}
		if userID := u.GetSomeID(); userID != "" {
			statement := db.NewStatement(`
						UPDATE users
						SET lastWebsiteVisit=?
						WHERE id=?`)
			if _, err := statement.Exec(database.Now(), userID); err != nil {
				fail(http.StatusInternalServerError, "Couldn't update users", err)
				return
			}
		}
		// Check if we have access to the private domain
		if core.IsIntIDValid(params.PrivateDomain.ID) {
			if !core.CanUserSeeDomain(u, params.PrivateDomain.ID) && r.URL.Path != "/login/" {
				fail(http.StatusForbidden, "Don't have access to this domain", nil)
				return
			}
		}

		// Call the page's renderer
		result := p.Render(&params)
		if result.Err != nil {
			c.Errorf("Primary renderer failed")
			fail(result.ResponseCode, result.Err.Message, result.Err.Err)
			return
		}

		if d, ok := result.Data.(dynamicPageTmplData); ok || result.Data == nil {
			isLive := !appengine.IsDevAppServer()
			devPrefix := "http://localhost:8014" // Keep in sync with webpack.config.js
			if isLive {
				devPrefix = ""
			}
			mergo.Merge(&d, dynamicPageTmplData{
				Title:             "Arbital",
				URL:               "https://" + r.Host + r.RequestURI,
				VersionID:         appengine.VersionID(c),
				IsLive:            isLive,
				MaybeServerPrefix: devPrefix,
			})
			result.Data = d
		}

		p.ServeHTTP(w, r, result)
		c.Inc(fmt.Sprintf("%s-success", r.URL.Path))
	}
}

// newPage returns a new page using default options.
func newPage(renderer pages.Renderer, tmpls []string) pages.Page {
	return pages.Add("", renderer, pages.PageOptions{}, tmpls...)
}
