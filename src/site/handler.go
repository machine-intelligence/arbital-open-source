// handler.go: Logic for modifying HTTP handlers.
package site

import (
	"fmt"
	"net/http"

	"appengine/taskqueue"

	"zanaduu3/src/config"
	"zanaduu3/src/sessions"
	"zanaduu3/src/twitter"
	"zanaduu3/src/user"

	"github.com/hkjn/pages"
)

// Handler serves HTTP.
type handler http.HandlerFunc

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
	return h.loggedIn().domain().monitor()
}

// sendToAuth redirects to an authentication URL.
func sendToAuth(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	var authUrl string
	var err error
	if !sessions.Live && config.XC.Site.Dev.Auth == "fake" {
		// On dev when requested we use a fake static user instead of real auth.
		authUrl = "/"
		err = user.BecomeFakeUser(w, r)
	} else {
		authUrl, err = twitter.GetAuthUrl(w, r)
	}
	if err != nil {
		showOops(w, r, fmt.Errorf("error fetching auth URL: %v", err))
		return
	}
	c.Infof("redirecting to auth URL %q..\n", authUrl)
	http.Redirect(w, r, authUrl, http.StatusSeeOther)
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

// loggedIn forwards to specified handler iff we have sign in info for the user.
//
// If there is no currently signed in user, loggedIn forwards to the
// authentication URL to connect with Twitter, then calls /verify_credentials
// to retrieve user info.
//
// Note that we don't check that the credentials stay current, since
// Twitter access tokens don't expire.
func (fn handler) loggedIn() handler {
	return func(w http.ResponseWriter, r *http.Request) {
		c := sessions.NewContext(r)
		c.Debugf("loggedIn check for %q", r.URL)
		q := r.URL.Query()
		errorMsg := q.Get("error_msg")
		var u *user.User
		s, err := sessions.GetSession(r)
		if err != nil {
			pages.ShowError(w, r, fmt.Errorf("failed to get session: %v", err))
			return
		}
		creds, err := sessions.LoadCreds(c, s)
		if err != nil {
			pages.ShowError(w, r, fmt.Errorf("error loading credentials: %v", err))
			return
		}
		if creds == nil && errorMsg == "" {
			// NOTE: We check errorMsg here and below to avoid going into
			// redirect loop between Twitter and us in case of programming bug
			// in auth handling within /authorize_callback.
			c.Debugf("no real credentials, sending user to auth URL..\n")
			sendToAuth(w, r)
			return
		} else if creds != nil {
			c.Debugf("we have real credentials, now checking user info..\n")
			var err error
			u, err = user.LoadUser(r)
			if err != nil {
				pages.ShowError(w, r, fmt.Errorf("error loading user: %v", err))
				return
			}
			if u != nil {
				c.Debugf("an existing user, welcome %q (%d), and onward to the handler..\n", u.Twitter.ScreenName, u.Twitter.Id)
				u.TwitterCreds = creds
				fn(w, r)
				return
			}
			c.Debugf("no user in session, fetching new one using credentials\n")
			var twitterUser *twitter.TwitterUser
			twitterUser, err = twitter.NewUser(w, r, creds)
			if err != nil {
				c.Errorf("error fetching user info, sending user to auth URL to start over: %v", err)
				if errorMsg == "" {
					sendToAuth(w, r)
					return
				}
			}
			u := user.User{Twitter: *twitterUser, TwitterCreds: creds}
			err = u.Save(w, r)
			if err != nil {
				pages.ShowError(w, r, fmt.Errorf("error saving user: %v", err))
				return
			} else {
				c.Debugf("created new user, welcome %q (%d), and onward to the handler..\n", u.Twitter.ScreenName, u.Twitter.Id)
				fn(w, r)
			}
		}
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
