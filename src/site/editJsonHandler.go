// editJsonHandler.go contains the handler for returning JSON with pages data.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var editHandler = siteHandler{
	URI:         "/json/edit/",
	HandlerFunc: editJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// editJsonData contains parameters passed in via the request.
type editJsonData struct {
	PageAlias      string
	SpecificEdit   int
	EditLimit      int
	CreatedAtLimit string
}

// editJsonHandler handles the request.
func editJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data editJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	return editJsonInternalHandler(params, &data)
}

func editJsonInternalHandler(params *pages.HandlerParams, data *editJsonData) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	} else if !ok {
		if core.IsIdValid(data.PageAlias) {
			// We tried to load an edit by id, but it wasn't found
			return pages.Fail("No such page found", err)
		} else {
			// We tried to load an edit by alias, it wasn't found, but we can create a
			// new page with that alias.
			return newPageJsonInternalHandler(params, &newPageJsonData{
				Type:  core.WikiPageType,
				Alias: data.PageAlias,
			})
		}
	}

	// Load full edit for one page.
	options := core.LoadEditOptions{
		LoadNonliveEdit:   true,
		LoadSpecificEdit:  data.SpecificEdit,
		LoadEditWithLimit: data.EditLimit,
		CreatedAtLimit:    data.CreatedAtLimit,
	}
	p, err := core.LoadFullEdit(db, pageId, u, &options)
	if err != nil {
		return pages.Fail("Error while loading full edit", err)
	}
	if p == nil {
		return pages.Fail("Exact page not found", err)
	}
	if p.SeeGroupId != params.PrivateGroupId {
		if core.IsIdValid(p.SeeGroupId) {
			return pages.Fail("Trying to edit a private page. Go to the corresponding group", err).Status(http.StatusBadRequest)
		} else {
			return pages.Fail("Trying to edit a public page. Go to arbital.com", err).Status(http.StatusBadRequest)
		}
	}
	if !p.IsAutosave && !p.IsSnapshot {
		p.PrevEdit = p.Edit
	}

	// Load parents, tags, requirement, and lens pages (to display in Relationship tab)
	core.AddPageToMap("3n", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("178", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("1ln", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("17b", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("35z", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("370", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("187", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("185", returnData.PageMap, core.TitlePlusLoadOptions)
	// Load data
	core.AddPageToMap(pageId, returnData.PageMap, core.PrimaryEditLoadOptions)
	core.AddPageIdToMap(p.EditGroupId, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	// We need to copy some data from the loaded live version to the edit
	// NOTE: AAAAARGH! This is such an ugly workaround
	livePage := returnData.PageMap[pageId]
	p.LensIds = livePage.LensIds
	p.ChildIds = livePage.ChildIds
	p.ParentIds = livePage.ParentIds
	p.TaggedAsIds = livePage.TaggedAsIds
	p.RequirementIds = livePage.RequirementIds
	p.SubjectIds = livePage.SubjectIds
	p.ChangeLogs = livePage.ChangeLogs
	p.SearchStrings = livePage.SearchStrings
	returnData.EditMap[pageId] = p

	// Clear change logs from the live page
	livePage.ChangeLogs = []*core.ChangeLog{}

	return pages.Success(returnData)
}
