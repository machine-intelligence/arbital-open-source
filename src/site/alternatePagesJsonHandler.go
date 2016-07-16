// alternatePagesJsonHandler.go returns a list of pages the user might want to read instead of the given page
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// alternatePagesData is the data received from the request.
type alternatePagesData struct {
	PageId string
}

var alternatePagesHandler = siteHandler{
	URI:         "/json/alternatePages/",
	HandlerFunc: alternatePagesJsonHandler,
	Options:     pages.PageOptions{},
}

func alternatePagesJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	decoder := json.NewDecoder(params.R.Body)
	var data alternatePagesData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}

	subjectsTaughtByThisPage, err := core.GetSubjects(db, data.PageId)
	if err != nil {
		return pages.Fail("Couldn't get subjects taught by the page", err)
	}

	// load title and requisite info for pages that also teach any of the subjects taught by this page
	loadOptions := (&core.PageLoadOptions{Requisites: true}).Add(core.TitlePlusLoadOptions)

	rows := database.NewQuery(`
		SELECT childId
		FROM pagePairs AS pp
		JOIN`).AddPart(core.PageInfosTable(nil)).Add(`AS pi
		ON pp.childId=pi.pageId
		WHERE pp.parentId IN`).AddArgsGroupStr(subjectsTaughtByThisPage).Add(`
			AND pp.type=?`, core.SubjectPagePairType).ToStatement(db).Query()
	alternateTeachers, err := core.LoadPageIds(rows, returnData.PageMap, loadOptions)
	if err != nil {
		return pages.Fail("Error while loading alternate pages", err)
	}

	returnData.ResultMap["alternate_teachers"] = alternateTeachers

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
