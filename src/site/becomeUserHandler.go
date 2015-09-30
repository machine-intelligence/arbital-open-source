// becomeUserHandler.go allows an admin to become any user.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// becomeUserHandler renders the comment page.
func becomeUserHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("become_user_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user object
	u, err := user.LoadUser(w, r, db)
	if err != nil {
		c.Inc("become_user_fail")
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
		fmt.Fprintf(w, "Please replace id=0 with actual id in the url.")
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
	fmt.Fprintf(w, "You are now agent #00%s", q.Get("id"))
}
