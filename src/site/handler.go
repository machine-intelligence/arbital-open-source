// handler.go: Logic for modifying HTTP handlers.
package site

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"appengine/taskqueue"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// Handler serves HTTP.
type handler http.HandlerFunc

type returnJsonData map[string]interface{}

// siteHandler is our type for HTTP request handlers
type siteHandler func(*pages.HandlerParams) *pages.Result

// commonPageData contains data that is common between all pages.
type commonPageData struct {
	// Id of the page that's most prominantly displayed. Usually the id is also in the URL
	PrimaryPageId int64 `json:",string"`
	// Map of page id -> currently live version of the page
	PageMap map[int64]*core.Page
	// Map of page id -> some edit of the page
	EditMap map[int64]*core.Page
	// Logged in user
	User *user.User
	// Map of user ids to corresponding user objects
	UserMap    map[int64]*core.User
	GroupMap   map[int64]*core.Group
	MasteryMap map[int64]*core.Mastery
	// Primary domain
	Domain *core.Group
}

// handlerWrapper wraps our siteHandler to provide standard http handler interface.
func handlerWrapper(h siteHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rand.Seed(time.Now().UnixNano())

		c := sessions.NewContext(r)
		fail := func(responseCode int, message string, err error) {
			c.Inc(fmt.Sprintf("%s-fail", r.URL.Path))
			c.Errorf("handlerWrapper: %s: %v", message, err)
			w.WriteHeader(responseCode)
			fmt.Fprintf(w, "%s", message)
		}

		// Recover from panic.
		defer func() {
			if r := recover(); r != nil {
				c.Errorf("%v", r)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", "Super serious error has occured. Super. Serious. Error.")
			}
		}()
		// Open DB connection
		db, err := database.GetDB(c)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't open DB", err)
			return
		}

		// Get user object
		var u *user.User
		u, err = user.LoadUser(w, r, db)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't load user", err)
			return
		}
		result := h(&pages.HandlerParams{W: w, R: r, C: c, DB: db, U: u})
		if result.ResponseCode != http.StatusOK {
			fail(result.ResponseCode, result.Message, result.Err)
			return
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

// pageHandlerWrapper wraps one of our page handlers.
func pageHandlerWrapper(p *pages.Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rand.Seed(time.Now().UnixNano())

		c := sessions.NewContext(r)
		params := pages.HandlerParams{W: w, R: r, C: c}

		addFuncMap := func(result *pages.Result, u *user.User) {
			// Set up Go TMPL functions.
			result.AddFuncMap(template.FuncMap{
				"UserId":     func() int64 { return u.Id },
				"IsAdmin":    func() bool { return u.IsAdmin },
				"IsLoggedIn": func() bool { return u.IsLoggedIn },
				"GetUserUrl": func(userId int64) string {
					return core.GetUserUrl(userId)
				},
				"GetCurrentUserJson": func() template.JS {
					jsonData, _ := json.Marshal(u)
					return template.JS(string(jsonData))
				},
				"GetUserJson": func(u *core.User) template.JS {
					jsonData, _ := json.Marshal(u)
					return template.JS(string(jsonData))
				},
				"GetPageJson": func(p *core.Page) template.JS {
					jsonData, _ := json.Marshal(p)
					return template.JS(string(jsonData))
				},
				"GetGroupJson": func(g *core.Group) template.JS {
					jsonData, _ := json.Marshal(g)
					return template.JS(string(jsonData))
				},
				"GetMasteryJson": func(m *core.Mastery) template.JS {
					jsonData, _ := json.Marshal(m)
					return template.JS(string(jsonData))
				},
				"GetPageUrl": func(p *core.Page) string {
					return core.GetPageUrl(p.PageId)
				},
				"IsUpdatedPage": func(p *core.Page) bool {
					return p.CreatorId != u.Id && p.LastVisit != "" && p.CreatedAt >= p.LastVisit
				},
				"CanComment":            func() bool { return u.Karma >= core.CommentKarmaReq },
				"CanLike":               func() bool { return u.Karma >= core.LikeKarmaReq },
				"CanCreatePrivatePage":  func() bool { return u.Karma >= core.PrivatePageKarmaReq },
				"CanVote":               func() bool { return u.Karma >= core.VoteKarmaReq },
				"CanKarmaLock":          func() bool { return u.Karma >= core.KarmaLockKarmaReq },
				"CanCreateAlias":        func() bool { return u.Karma >= core.CreateAliasKarmaReq },
				"CanChangeAlias":        func() bool { return u.Karma >= core.ChangeAliasKarmaReq },
				"CanChangeSortChildren": func() bool { return u.Karma >= core.ChangeSortChildrenKarmaReq },
				"CanAddParent":          func() bool { return u.Karma >= core.AddParentKarmaReq },
				"CanDeleteParent":       func() bool { return u.Karma >= core.DeleteParentKarmaReq },
				"CanDashlessAlias":      func() bool { return u.Karma >= core.DashlessAliasKarmaReq },
			})
		}

		// Helper func to when an error occurs and we should render error page.
		fail := func(responseCode int, message string, err error) {
			c.Inc(fmt.Sprintf("%s-fail", r.URL.Path))
			c.Errorf("%s: %v", message, err)
			result := renderErrorPage(&params, message)
			addFuncMap(result, params.U)
			w.WriteHeader(responseCode)
			errorPage.ServeHTTP(w, r, result)
		}

		// Recover from panic.
		defer func() {
			if r := recover(); r != nil {
				c.Errorf("%v", r)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", "Super serious error has occured. Super. Serious. Error.")
			}
		}()

		// Open DB connection
		db, err := database.GetDB(c)
		if err != nil {
			fail(http.StatusInternalServerError, "Couldn't open DB", err)
			return
		}
		params.DB = db

		// Get user object
		var u *user.User
		if !p.Options.SkipLoadingUser {
			u, err = user.LoadUser(w, r, db)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't load user", err)
				return
			}
			params.U = u

			// Check user state
			if u.Id > 0 && len(u.FirstName) <= 0 && r.URL.Path != "/signup/" {
				// User has created an account but hasn't gone through signup page
				http.Redirect(w, r, fmt.Sprintf("/signup/?continueUrl=%s", r.URL), http.StatusSeeOther)
				return
			}
			if p.Options.RequireLogin && !u.IsLoggedIn {
				fail(http.StatusBadRequest, "Have to be logged in", nil)
			}
			if p.Options.AdminOnly && !u.IsAdmin {
				fail(http.StatusBadRequest, "Have to be an admin", nil)
			}
			if u.Id > 0 {
				statement := db.NewStatement(`
						UPDATE users
						SET lastWebsiteVisit=?
						WHERE id=?`)
				if _, err := statement.Exec(database.Now(), u.Id); err != nil {
					fail(http.StatusInternalServerError, "Couldn't update users", err)
				}
				// Load the groups the user belongs to.
				if err = core.LoadUserGroups(db, u); err != nil {
					fail(http.StatusInternalServerError, "Couldn't load user groups", err)
				}
			}
		}

		// Call the page's renderer
		result := p.Render(&params)
		if result.Data == nil {
			c.Debugf("Primary renderer failed")
			fail(result.ResponseCode, result.Message, result.Err)
			return
		}

		// Load more user stuff if required.
		if !p.Options.SkipLoadingUser && u.Id > 0 {
			// Load updates count. (Loading it afterwards since it could be affected by the page)
			u.UpdateCount, err = core.LoadUpdateCount(db, u.Id)
			if err != nil {
				fail(http.StatusInternalServerError, "Couldn't retrieve updates count", err)
			}
		}

		addFuncMap(result, u)
		w.Header().Add("Cache-Control", "max-age=0, no-cache, no-store")
		p.ServeHTTP(w, r, result)
		c.Inc(fmt.Sprintf("%s-success", r.URL.Path))
	}
}

// newPage returns a new page using default options.
func newPage(uri string, renderer pages.Renderer, tmpls []string) pages.Page {
	return newPageWithOptions(uri, renderer, tmpls, pages.PageOptions{})
}

// newPageWithOptions returns a new page with the given options.
func newPageWithOptions(uri string, renderer pages.Renderer, tmpls []string, options pages.PageOptions) pages.Page {
	return pages.Add(uri, renderer, &options, tmpls...)
}

// createReturnData puts together various maps into one "json" object, so we
// can send it to the front-end.
func createReturnData(pages map[int64]*core.Page) returnJsonData {
	returnPageData := make(map[string]*core.Page)
	for k, v := range pages {
		returnPageData[fmt.Sprintf("%d", k)] = v
	}
	returnData := make(returnJsonData)
	returnData["pages"] = returnPageData
	return returnData
}

func (d returnJsonData) AddResult(result interface{}) returnJsonData {
	d["result"] = result
	return d
}

func (d returnJsonData) AddEditMap(pages map[int64]*core.Page) returnJsonData {
	returnEditData := make(map[string]*core.Page)
	for k, v := range pages {
		returnEditData[fmt.Sprintf("%d", k)] = v
	}
	d["edits"] = returnEditData
	return d
}

func (d returnJsonData) AddUsers(users map[int64]*core.User) returnJsonData {
	returnUserData := make(map[string]*core.User)
	for k, v := range users {
		returnUserData[fmt.Sprintf("%d", k)] = v
	}
	d["users"] = returnUserData
	return d
}

func (d returnJsonData) AddMasteries(masteries map[int64]*core.Mastery) returnJsonData {
	returnMasteryData := make(map[string]*core.Mastery)
	for k, v := range masteries {
		returnMasteryData[fmt.Sprintf("%d", k)] = v
	}
	d["masteries"] = returnMasteryData
	return d
}

func (d returnJsonData) AddGroups(groups map[int64]*core.Group) returnJsonData {
	returnGroupData := make(map[string]*core.Group)
	for k, v := range groups {
		returnGroupData[fmt.Sprintf("%d", k)] = v
	}
	d["groups"] = returnGroupData
	return d
}

// domain redirects to proper HTML domain if user arrives elsewhere.
//
// The need for this is e.g. for http://foo.rewards.xelaie.com/, which
// would set a cookie for that domain but redirect to
// http://rewards.xelaie.com after sign-in.
func (fn handler) domain() handler {
	return func(w http.ResponseWriter, r *http.Request) {
		c := sessions.NewContext(r)
		d := fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Host)
		c.Debugf("domain check for %s", d)
		if sessions.Live && d != sessions.GetDomain() {
			// TODO: if we cared enough here, we could preserve
			// r.URL.{Path,RawQuery} in the redirect.
			c.Warningf("request arrived for %s, but live domain is %s. redirecting..\n", d, sessions.GetDomain())
			c.Inc("bad_domain_visited")
			http.Redirect(w, r, sessions.GetDomain(), http.StatusSeeOther)
		} else {
			fn(w, r)
		}
		return
	}
}

// monitor sends counters added within the handler off to monitoring.
func (fn handler) monitor() handler {
	return func(w http.ResponseWriter, r *http.Request) {
		c := sessions.NewContext(r)
		fn(w, r)
		// At end of each request, add task to reporting queue.
		t, err := c.Report()
		if err != nil {
			c.Errorf("failed to create monitoring task: %v\n", err)
			return
		}
		if t == nil {
			// no monitoring task, nothing to do.
			return
		}
		_, err = taskqueue.Add(c, t, "report-monitoring")
		if err != nil {
			c.Errorf("failed to add monitoring POST task to queue: %v\n", err)
			return
		}
	}
}
