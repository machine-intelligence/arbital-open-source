// changeSpeedHandler.go returns a list of pages the user might want if they want an different version
package site

import (
	"encoding/json"
	"fmt"
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

	// Find arcs that teach this page
	p := core.AddPageIDToMap(data.PageID, returnData.PageMap)
	p.ArcPageIDs, err = _loadArcs(db, data.PageID, returnData)
	if err != nil {
		return pages.Fail("Couldn't load arcs", err)
	}

	// Load pages.
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

func _loadArcs(db *database.DB, pageID string, returnData *core.CommonHandlerData) ([]string, error) {
	arcPageIDs := make([]string, 0)
	rows := database.NewQuery(`
		SELECT guideId
		FROM pathPages AS pathPages
		JOIN`).AddPart(core.PageInfosTableAll(returnData.User)).Add(`AS pi
		ON (pi.pageId=pathPages.guideId)
		WHERE pathPageId=?`, pageID).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var guideID string
		err := rows.Scan(&guideID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		arcPageIDs = append(arcPageIDs, guideID)
		core.AddPageIDToMap(guideID, returnData.PageMap)
		return nil
	})
	return arcPageIDs, err
}

func _loadChangeSpeedPagePairs(db *database.DB, slower bool, pageID string, returnData *core.CommonHandlerData) ([]*core.PagePair, error) {
	pagePairs := make([]*core.PagePair, 0)
	comparison := "<"
	if !slower {
		comparison = ">"
	}

	pageSpeeds := database.NewQuery(
		`(SELECT childId AS pageId, IF(SUM(parentId='6b4'), -1, IF(SUM(parentId='6b5'), 1, 0)) AS speed FROM pagePairs WHERE type='tag' GROUP BY childId)`)

	queryPart := database.NewQuery(`
		/* find pages (pp.childId) that teach one of the same subjects as pageId teaches, but at a lower or higher level */
		JOIN (
			SELECT parentId as subjectId, level
			FROM pagePairs
			WHERE childId=?`, pageID).Add(`AND type=?`, core.SubjectPagePairType).Add(`AND isStrong
		) AS subjects
		ON pp.parentId=subjectId AND pp.type=?`, core.SubjectPagePairType).Add(`AND pp.isStrong
			AND (
				/* either the pp.childId page teaches at a lower/higher level than the pageID page ...*/
				pp.level `+comparison+` subjects.level

				OR
				/* ... or it teaches at the same level but at a slower/faster speed */
				(
					pp.level=subjects.level AND
					COALESCE((SELECT speed FROM`).AddPart(pageSpeeds).Add(`AS s1 WHERE s1.pageId=pp.childId), 0)`+comparison+`
					COALESCE((SELECT speed FROM`).AddPart(pageSpeeds).Add(`AS s2 WHERE s2.pageId=?), 0)`, pageID).Add(`
				)
			)

		/* filter for pages the user has access to */
		JOIN`).AddPart(core.PageInfosTableAll(returnData.User)).Add(`AS pi
		ON pi.pageId=pp.childId
	`)
	err := core.LoadPagePairs(db, queryPart, func(db *database.DB, pagePair *core.PagePair) error {
		pagePairs = append(pagePairs, pagePair)
		// Load tags so that we can determine the speed of the page on the front end.
		core.AddPageToMap(pagePair.ChildID, returnData.PageMap, &core.PageLoadOptions{Tags: true})
		return nil
	})
	return pagePairs, err
}
