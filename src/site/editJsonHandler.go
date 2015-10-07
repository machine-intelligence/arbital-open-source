// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"

	"github.com/gorilla/schema"
)

// editJsonData contains parameters passed in via the request.
type editJsonData struct {
	PageId         int64 `json:",string"`
	SpecificEdit   int
	EditLimit      int
	CreatedAtLimit string
}

// editJsonHandler handles the request.
func editJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data editJsonData
	params.R.ParseForm()
	err := schema.NewDecoder().Decode(&data, params.R.Form)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)

	// Load full edit for one page.
	options := loadEditOptions{
		loadSpecificEdit:  data.SpecificEdit,
		loadEditWithLimit: data.EditLimit,
		createdAtLimit:    data.CreatedAtLimit,
	}
	p, err := loadFullEdit(db, data.PageId, u.Id, &options)
	if err != nil || p == nil {
		return pages.HandlerErrorFail("error while loading full edit", err)
	}
	pageMap[data.PageId] = p

	// Load all the users
	for _, p := range pageMap {
		userMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, userMap)
	if err != nil {
		return pages.HandlerErrorFail("error while loading users", err)
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
	returnData := make(map[string]interface{})
	returnData["pages"] = returnPageData
	returnData["users"] = returnUserData

	return pages.StatusOK(returnData)
}
