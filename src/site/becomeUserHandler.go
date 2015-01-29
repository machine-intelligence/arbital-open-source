// becomeUserHandler.go allows an admin to become any user.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// becomeUserHandler renders the comment page.
func becomeUserHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Get user object
	u, err := user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	q := r.URL.Query()
	if q.Get("id") == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := u.BecomeUserWithId(q.Get("id"), c); err != nil {
		c.Errorf("Couldn't become user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := u.Save(w, r); err != nil {
		c.Errorf("Couldn't save user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "You are now agend #00%s", q.Get("id"))
}
