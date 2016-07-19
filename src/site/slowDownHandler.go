// slowDownHandler.go returns a list of pages the user might want if they want an easier version
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// slowDownData is the data received from the request.
type slowDownData struct {
	PageID string
}

var slowDownHandler = siteHandler{
	URI:         "/json/slowDown/",
	HandlerFunc: slowDownHandlerFunc,
	Options:     pages.PageOptions{},
}

func slowDownHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data slowDownData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Find other pages that teach the same subjects, but at easier levels
	pagePairs := make([]*core.PagePair, 0)
	queryPart := database.NewQuery(`
			/* pp2 selects pages that this page teaches */
			/* pp selects pages that also teach the same subjects but at lower levels */
			JOIN pagePairs AS pp2
			ON (pp.parentId=pp2.parentId AND pp.level<pp2.level)
			JOIN`).AddPart(core.PageInfosTableAll(u)).Add(`AS pi
			ON (pi.pageId=pp.childId)
			WHERE pp2.childId=?`, data.PageID).Add(`
				AND pp2.type=?`, core.SubjectPagePairType).Add(`
				AND pp.type=?`, core.SubjectPagePairType)
	err = core.LoadPagePairs(db, queryPart, func(db *database.DB, pagePair *core.PagePair) error {
		pagePairs = append(pagePairs, pagePair)
		core.AddPageIDToMap(pagePair.ChildID, returnData.PageMap)
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading page pairs", err)
	}

	// Load pages.
	p := core.AddPageIDToMap(data.PageID, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	p.SlowDownMap = make(map[string][]*core.PagePair)
	for _, pagePair := range pagePairs {
		p.SlowDownMap[pagePair.ParentID] = append(p.SlowDownMap[pagePair.ParentID], pagePair)
	}
	return pages.Success(returnData)
}
