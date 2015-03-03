// handler.go: Logic for modifying HTTP handlers.
package site

import (
	"fmt"
	"html/template"
	"net/http"

	"appengine/taskqueue"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// Handler serves HTTP.
type handler http.HandlerFunc

// newPageOptions specify options when we create a new html page.
// NOTE: make sure that default values are okay for all pages.
type newPageOptions struct {
	SkipLoadingUser bool
	RequireLogin    bool
}

// newHandler returns a standard handler from given handler function.
//
// The standard handlers requires are monitored, require the proper
// domain for live requests, and require the user to be logged in.
//
// Note that the order of the chaining is relevant - the right-most
// call is applied first, and if a check fails (e.g. the live domain
// is incorrect) that may cause further checks not to run (e.g. we
// wouldn't even check if the user was logged in).
func stdHandler(h handler) handler {
	return h.domain()
}

// newPageWithOptions returns a new page which will wrap the given renderer so
// that it gets the user object.
func newPageWithOptions(uri string, renderer pages.Renderer, tmpls []string, options newPageOptions) pages.Page {
	return pages.Add(uri, loadUserHandler(renderer, options), tmpls...)
}

// newPage returns a new page using default options.
func newPage(uri string, renderer pages.Renderer, tmpls []string) pages.Page {
	return newPageWithOptions(uri, renderer, tmpls, newPageOptions{})
}

// loadUserHandler is a wrapper around a randerer, which allows us to load the
// user object and add user related template functions.
func loadUserHandler(h pages.Renderer, options newPageOptions) pages.Renderer {
	return func(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
		var err error
		c := sessions.NewContext(r)
		if u != nil {
			c.Errorf("User is already set when calling loadUserHandler.")
		}
		if !options.SkipLoadingUser {
			u, err = user.LoadUser(w, r)
			if err != nil {
				c.Errorf("Couldn't load user: %v", err)
				return pages.InternalErrorWith(err)
			}
			if u.Id > 0 && len(u.FirstName) <= 0 && r.URL.Path != "/signup/" {
				// User has created an account but hasn't gone through signup page
				return pages.RedirectWith("/signup/")
			}
			if options.RequireLogin && u.Id <= 0 {
				return pages.UnauthorizedWith(fmt.Errorf("Not logged in"))
			}
		}
		result := h(w, r, u)
		funcMap := template.FuncMap{
			"UserId":     func() int64 { return u.Id },
			"IsAdmin":    func() bool { return u.IsAdmin },
			"IsLoggedIn": func() bool { return u.IsLoggedIn },
		}
		return result.AddFuncMap(funcMap)
	}
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
