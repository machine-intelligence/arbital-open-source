// changeSpeedHandler.go returns a list of pages the user might want if they want an different version
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// changeSpeedData is the data received from the request.
type changeSpeedData struct {
	PageID string
}

var changeSpeedHandler = siteHandler{
	URI:         "/json/changeSpeed/",
	HandlerFunc: changeSpeedHandlerFunc,
	Options:     pages.PageOptions{},
}

// ROGTODO: why does this have its own handler -- don't we load this information for every primary page?
func changeSpeedHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data changeSpeedData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Find other pages that teach the same subjects, but at different levels
	slowerPagePairs, err := _loadChangeSpeedPagePairs(db, true, data.PageID, returnData)
	if err != nil {
		return pages.Fail("Error while loading slower page pairs", err)
	}
	fasterPagePairs, err := _loadChangeSpeedPagePairs(db, false, data.PageID, returnData)
	if err != nil {
		return pages.Fail("Error while loading faster page pairs", err)
	}

	// Load pages.
	p := core.AddPageIDToMap(data.PageID, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	p.SlowDownMap = make(map[string][]*core.PagePair)
	for _, pagePair := range slowerPagePairs {
		p.SlowDownMap[pagePair.ParentID] = append(p.SlowDownMap[pagePair.ParentID], pagePair)
	}
	p.SpeedUpMap = make(map[string][]*core.PagePair)
	for _, pagePair := range fasterPagePairs {
		p.SpeedUpMap[pagePair.ParentID] = append(p.SpeedUpMap[pagePair.ParentID], pagePair)
	}
	return pages.Success(returnData)
}

func _loadChangeSpeedPagePairs(db *database.DB, slower bool, pageID string, returnData *core.CommonHandlerData) ([]*core.PagePair, error) {
	pagePairs := make([]*core.PagePair, 0)
	comparison := "<"
	if !slower {
		comparison = ">"
	}
	queryPart := database.NewQuery(`
			/* pp2 selects pages that this page teaches */
			/* pp selects pages that also teach the same subjects but at lower levels */
			JOIN pagePairs AS pp2
			ON (pp.parentId=pp2.parentId AND pp.level`+comparison+`pp2.level)
			JOIN`).AddPart(core.PageInfosTableAll(returnData.User)).Add(`AS pi
			ON (pi.pageId=pp.childId)
			WHERE pp.isStrong AND pp2.isStrong
				AND pp2.childId=?`, pageID).Add(`
				AND pp2.type=?`, core.SubjectPagePairType).Add(`
				AND pp.type=?`, core.SubjectPagePairType)
	err := core.LoadPagePairs(db, queryPart, func(db *database.DB, pagePair *core.PagePair) error {
		pagePairs = append(pagePairs, pagePair)
		core.AddPageIDToMap(pagePair.ChildID, returnData.PageMap)
		return nil
	})
	return pagePairs, err
}
