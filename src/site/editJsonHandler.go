// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/schema"
)

// editJsonData contains parameters passed in via the request.
type editJsonData struct {
	PageId         int64 `json:",string"`
	EditLimit      int
	CreatedAtLimit string
}

// editJsonHandler handles the request.
func editJsonHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Decode data
	var data editJsonData
	r.ParseForm()
	err := schema.NewDecoder().Decode(&data, r.Form)
	if err != nil {
		c.Inc("edit_json_handler_fail")
		c.Errorf("Couldn't decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	returnData, err := editJsonInternalHandler(w, r, &data)
	if err != nil {
		c.Inc("edit_json_handler_fail")
		c.Errorf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write JSON response
	err = writeJson(w, returnData)
	if err != nil {
		c.Inc("edit_handler_fail")
		c.Errorf("Couldn't write json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// editJsonInternalHandler handles the request.
func editJsonInternalHandler(w http.ResponseWriter, r *http.Request, data *editJsonData) (map[string]interface{}, error) {
	c := sessions.NewContext(r)
	returnData := make(map[string]interface{})

	// Load user object
	u, err := user.LoadUser(w, r)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load user: %v", err)
	}

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)

	// Load full edit for one page.
	options := loadEditOptions{loadEditWithLimit: data.EditLimit, createdAtLimit: data.CreatedAtLimit}
	p, err := loadFullEdit(c, data.PageId, u.Id, &options)
	if err != nil || p == nil {
		return nil, fmt.Errorf("error while loading full edit: %v", err)
	}
	pageMap[data.PageId] = p

	// Load all the users
	for _, p := range pageMap {
		userMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, userMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading users: %v", err)
	}

	// Return the data in JSON format.
	returnPageData := make(map[string]*core.Page)
	for k, v := range pageMap {
		returnPageData[fmt.Sprintf("%d", k)] = v
	}
	returnUserData := make(map[string]*core.User)
	for k, v := range userMap {
		returnUserData[fmt.Sprintf("%d", k)] = v
	}
	returnData["pages"] = returnPageData
	returnData["users"] = returnUserData

	return returnData, nil
}
