// childrenJsonHandler.go contains the handler for returning JSON with children pages.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/schema"
)

// childrenJsonData contains parameters passed in to create a page.
type childrenJsonData struct {
	ParentId int64
}

// childrenJsonHandler handles requests to create a new page.
func childrenJsonHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Decode data
	var data childrenJsonData
	r.ParseForm()
	err := schema.NewDecoder().Decode(&data, r.Form)
	if err != nil {
		c.Errorf("Couldn't decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.ParentId <= 0 {
		c.Errorf("Need a valid parentId", err)
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

	// Load the children.
	pageMap := make(map[int64]*page)
	pageMap[data.ParentId] = &page{PageId: data.ParentId}
	err = loadChildrenIds(c, pageMap, loadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		c.Errorf("Couldn't load children: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Remove parent, since we only want to return children.
	delete(pageMap, data.ParentId)

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

	// Return the page in JSON format.
	strPageMap := make(map[string]*page)
	for k, v := range pageMap {
		strPageMap[fmt.Sprintf("%d", k)] = v
	}
	var jsonData []byte
	jsonData, err = json.Marshal(strPageMap)
	if err != nil {
		fmt.Println("Error marshalling pageMap into json:", err)
	}
	w.Write(jsonData)
}
