// updatePageObject.go handles request to update the value of a page object
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updatePageObject contains the data we get in the request.
type updatePageObject struct {
	PageId string
	Edit   int
	Object string
	Value  string
}

var updatePageObjectHandler = siteHandler{
	URI:         "/updatePageObject/",
	HandlerFunc: updatePageObjectHandlerFunc,
	Options:     pages.PageOptions{},
}

func updatePageObjectHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data updatePageObject
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	return updatePageObjectInternalHandlerFunc(params, &data)
}

func updatePageObjectInternalHandlerFunc(params *pages.HandlerParams, data *updatePageObject) *pages.Result {
	db := params.DB
	u := params.U

	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	}
	if data.Object == "" {
		return pages.Fail("Object alias isn't set", nil).Status(http.StatusBadRequest)
	}
	userId := u.GetSomeId()
	if userId == "" {
		return pages.Fail("No user id or session id", nil).Status(http.StatusBadRequest)
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = userId
	hashmap["pageId"] = data.PageId
	hashmap["edit"] = data.Edit
	hashmap["object"] = data.Object
	hashmap["value"] = data.Value
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	statement := db.NewInsertStatement("userPageObjectPairs", hashmap, "edit", "value", "updatedAt")
	if _, err := statement.Exec(); err != nil {
		return pages.Fail("Couldn't update a page object", err)
	}

	return pages.Success(nil)
}
