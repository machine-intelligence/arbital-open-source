// parentsJsonHandler.go contains the handler for returning JSON with parents pages.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("page_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load the parents.
	pageMap := make(map[int64]*page)
	pageMap[data.ChildId] = &page{PageId: data.ChildId}
	err = loadParentsIds(c, pageMap, loadParentsIdsOptions{LoadHasParents: true})
	if err != nil {
		c.Errorf("Couldn't load parent ids: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Remove child, since we only want to return parents.
	delete(pageMap, data.ChildId)

	// Load pages.
	err = loadPages(c, pageMap, u.Id, loadPageOptions{})
	if err != nil {
		c.Errorf("error while loading pages: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load likes.
	err = loadLikes(c, u.Id, pageMap)
	if err != nil {
		c.Errorf("Couldn't retrieve page likes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load probability votes
	/*err = loadVotes(c, u.Id, pageIdStr, pageMap)
	if err != nil {
		c.Errorf("Couldn't load probability votes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}*/

	// Return the pages in JSON format.
	strPageMap := make(map[string]*page)
	for k, v := range pageMap {
		strPageMap[fmt.Sprintf("%d", k)] = v
	}
	var jsonData []byte
	jsonData, err = json.Marshal(strPageMap)
	if err != nil {
		fmt.Println("Error marshalling pageMap into json:", err)
	}
	// Write some stuff for "JSON Vulnerability Protection"
	w.Write([]byte(")]}',\n"))
	w.Write(jsonData)
}
