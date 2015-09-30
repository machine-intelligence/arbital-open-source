// childrenJsonHandler.go contains the handler for returning JSON with children pages.
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

	err = childrenJsonInternalHandler(w, r, &data)
	if err != nil {
		c.Errorf("%v", err)
		c.Inc("children_json_handler_fail")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// childrenJsonInternalHandler handles requests to create a new page.
func childrenJsonInternalHandler(w http.ResponseWriter, r *http.Request, data *childrenJsonData) error {
	c := sessions.NewContext(r)

	db, err := database.GetDB(c)
	if err != nil {
		return err
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		return fmt.Errorf("Couldn't load user: %v", err)
	}

	// Load the children.
	pageMap := make(map[int64]*core.Page)
	pageMap[data.ParentId] = &core.Page{PageId: data.ParentId}
	err = loadChildrenIds(db, pageMap, loadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		return fmt.Errorf("Couldn't load children: %v", err)
	}
	// Remove parent, since we only want to return children.
	delete(pageMap, data.ParentId)

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, nil)
	if err != nil {
		return fmt.Errorf("error while loading pages: %v", err)
	}

	// Load likes.
	err = loadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return fmt.Errorf("Couldn't load aux data: %v", err)
	}

	// Load probability votes
	/*err = loadVotes(c, u.Id, pageIdStr, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load probability votes: %v", err)
	}*/

	// Return the page in JSON format.
	strPageMap := make(map[string]*core.Page)
	for k, v := range pageMap {
		strPageMap[fmt.Sprintf("%d", k)] = v
	}
	err = writeJson(w, strPageMap)
	if err != nil {
		fmt.Println("Error writing data to json:", err)
	}
	return nil
}
