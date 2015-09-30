// parentsJsonHandler.go contains the handler for returning JSON with parents pages.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/schema"
)

// parentsJsonData contains parameters passed in via the request.
type parentsJsonData struct {
	ChildId int64
}

// parentsJsonHandler handles the request.
func parentsJsonHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Decode data
	var data parentsJsonData
	r.ParseForm()
	err := schema.NewDecoder().Decode(&data, r.Form)
	if err != nil {
		c.Errorf("Couldn't decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.ChildId <= 0 {
		c.Errorf("Need a valid childId", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("parents_json_handler_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		c.Inc("parents_json_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load the parents.
	pageMap := make(map[int64]*core.Page)
	pageMap[data.ChildId] = &core.Page{PageId: data.ChildId}
	err = loadParentsIds(db, pageMap, loadParentsIdsOptions{LoadHasParents: true})
	if err != nil {
		c.Errorf("Couldn't load parent ids: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Remove child, since we only want to return parents.
	delete(pageMap, data.ChildId)

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, nil)
	if err != nil {
		c.Errorf("error while loading pages: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load auxillary data.
	err = loadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		c.Errorf("Couldn't retrieve page likes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load probability votes
	/*err = loadVotes(db, u.Id, pageIdStr, pageMap)
	if err != nil {
		c.Errorf("Couldn't load probability votes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}*/

	// Return the pages in JSON format.
	strPageMap := make(map[string]*core.Page)
	for k, v := range pageMap {
		strPageMap[fmt.Sprintf("%d", k)] = v
	}
	err = writeJson(w, strPageMap)
	if err != nil {
		fmt.Println("Error writing data to json:", err)
	}
}
