// Package sessions provides the necessary structures and functions to work with Contexts and Sessions
package sessions

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"

	"appengine"

	"zanaduu3/src/config"

	"github.com/gorilla/sessions"
)

var (
	Live  = !appengine.IsDevAppServer()
	ctxs  *contexts             // All request contexts.
	store *sessions.CookieStore // Gorilla session store.
)

// GetDomain returns the domain we are working under.
func GetDomain() string {
	address := config.XC.Site.Dev.Address
	if Live {
		return config.XC.Site.Live.Address
	}
	return strings.TrimPrefix(address, "http://")
}

func GetMuxDomain() string {
	address := "localhost"
	if Live {
		address = config.XC.Site.Live.Address
	}
	return strings.TrimPrefix(address, "http://")
}

func GetDomainForTestEmail() string {
	address := ""
	if Live {
		return config.XC.Site.Live.Address
	}
	return strings.TrimPrefix(address, "http://")
}

// GetRawDomain returns the domain without http:// but with port #.
func GetRawDomain() string {
	address := config.XC.Site.Dev.Address
	if Live {
		address = config.XC.Site.Live.Address
	}
	return strings.TrimPrefix(address, "http://")
}

func GetElasticDomain() string {
	if Live {
		return config.XC.Elastic.Live.Address
	}
	return config.XC.Elastic.Dev.Address
}

// GetSession returns the user's session.
func GetSession(r *http.Request) (*sessions.Session, error) {
	c := NewContext(r)
	session, err := store.Get(r, config.XC.Site.Cookie)
	if err != nil {
		if session.IsNew {
			// Apparently errors are always raised when first creating session?
			// C.f. `godoc github.com/gorilla/sessions | grep 'ignoring the error' -C 2`
			c.Debugf("ignoring error retrieving session, since it is new: %v\n", err)
		} else {
			c.Inc("session_creation_failed")
			return nil, fmt.Errorf("error fetching session: %v", err)
		}
	}
	return session, nil
}

func init() {
	gob.Register(&Credentials{})

	ctxs = &contexts{
		m: make(map[requestID]Context),
	}
	store = sessions.NewCookieStore(
		[]byte(config.XC.Site.Session.Auth),
		[]byte(config.XC.Site.Session.Crypt))
}
