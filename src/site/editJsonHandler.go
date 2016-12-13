// editJsonHandler.go contains the handler for returning JSON with pages data.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

var editHandler = siteHandler{
	URI:         "/json/edit/",
	HandlerFunc: editJSONHandler,
	Options:     pages.PageOptions{},
}

// editJsonData contains parameters passed in via the request.
type editJSONData struct {
	PageAlias      string
	SpecificEdit   int
	EditLimit      int
	CreatedAtLimit string

	// Optional pages to load as well (e.g. for making a quick parent)
	AdditionalPageIDs []string
}

// editJsonHandler handles the request.
func editJSONHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data editJSONData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	return editJSONInternalHandler(params, &data)
}

func editJSONInternalHandler(params *pages.HandlerParams, data *editJSONData) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	// Get actual page id
	pageID, ok, err := core.LoadAliasToPageID(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	} else if !ok {
		if core.IsIDValid(data.PageAlias) {
			// We tried to load an edit by id, but it wasn't found
			return pages.Fail("No such page found", err)
		} else {
			// We tried to load an edit by alias, it wasn't found, but we can create a
			// new page with that alias.
			return newPageInternalHandler(params, &newPageData{
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
	p, err := core.LoadFullEdit(db, pageID, u, returnData.DomainMap, &options)
	if err != nil {
		return pages.Fail("Error while loading full edit", err)
	}
	if p == nil {
		return pages.Fail("Exact page not found", err)
	}
	if p.SeeDomainID != params.PrivateDomain.ID {
		if core.IsIntIDValid(p.SeeDomainID) {
			return pages.Fail("Trying to edit a private page. Go to the corresponding group", err).Status(http.StatusBadRequest)
		} else {
			return pages.Fail("Trying to edit a public page. Go to arbital.com", err).Status(http.StatusBadRequest)
		}
	}

	// If it's an autosave or a snapshot, we can't count on all links to be loaded,
	// since they are not stored in the links table. So we manually extract them.
	if p.IsAutosave || p.IsSnapshot {
		linkAliases := core.ExtractPageLinks(p.Text, sessions.GetDomain())
		aliasToIDMap, err := core.LoadAliasToPageIDMap(db, u, linkAliases)
		if err != nil {
			return pages.Fail("Couldn't load links", err)
		}
		for _, pageID := range aliasToIDMap {
			core.AddPageToMap(pageID, returnData.PageMap, &core.PageLoadOptions{VoteSummary: true})
		}
	}

	// Load additional pages (for which we need to display a greenlink)
	if data.AdditionalPageIDs == nil {
		data.AdditionalPageIDs = make([]string, 0)
	}
	data.AdditionalPageIDs = append(data.AdditionalPageIDs, "3n", "178", "1ln",
		"17b", "35z", "370", "187", "185", "3hs", "1rt", "595", "596", "597")
	for _, pageID := range data.AdditionalPageIDs {
		core.AddPageIDToMap(pageID, returnData.PageMap)
	}

	// Load data
	p.LoadOptions = *core.PrimaryEditLoadOptions
	returnData.PageMap[p.PageID] = p
	returnData.EditMap[pageID] = p
	core.AddPageIDToMap(p.EditDomainID, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
