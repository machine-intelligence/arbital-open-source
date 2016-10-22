// redLinkPopoverHandler.go contains the handler for returning data to display a red link popover.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// redLinkPopoverData contains parameters passed in via the request.
type redLinkPopoverData struct {
	PageAlias string
}

var redLinkPopoverHandler = siteHandler{
	URI:         "/json/redLinkPopover/",
	HandlerFunc: redLinkPopoverHandlerFunc,
}

// redLinkPopoverHandler handles the request.
func redLinkPopoverHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data redLinkPopoverData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Load associated likeable id
	redLinkRow := &RedLinkRow{
		Likeable: *core.NewLikeable(core.RedLinkLikeableType),
		Alias:    data.PageAlias,
	}
	row := database.NewQuery(`
		SELECT likeableId
		FROM redLinks
		WHERE alias=?`, data.PageAlias).ToStatement(db).QueryRow()
	_, err = row.Scan(&redLinkRow.LikeableID)
	if err != nil {
		return pages.Fail("Failed to scan likeableID", err)
	}

	// Load likes
	likeablesMap := make(map[int64]*core.Likeable)
	if redLinkRow.LikeableID != 0 {
		likeablesMap[redLinkRow.LikeableID] = &redLinkRow.Likeable
	}
	err = core.LoadLikes(db, u, likeablesMap, likeablesMap, returnData.UserMap)
	if err != nil {
		return pages.Fail("Couldn't load red link like count", err)
	}

	// Load related pages
	aliases := []string{data.PageAlias}
	related, err := loadRelationships(db, aliases, returnData, true)
	if err != nil {
		return pages.Fail("Couldn't load relationships", err)
	}
	redLinkRow.LinkedByPageIDs = related[redLinkRow.Alias]

	returnData.ResultMap["redLinkRow"] = redLinkRow
	return pages.Success(returnData)
}
