// moreRelationshipsJsonHandler.go serves JSON data to display on the additional relationships tab in the page editor.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type moreRelationshipsJSONData struct {
	PageAlias            string
	RestrictToMathDomain bool
}

var moreRelationshipsHandler = siteHandler{
	URI:         "/json/moreRelationships/",
	HandlerFunc: moreRelationshipsJSONHandler,
	Options:     pages.PageOptions{},
}

// moreRelationshipsJsonHandler handles the request.
func moreRelationshipsJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data moreRelationshipsJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsAliasValid(data.PageAlias) {
		return pages.Fail("Invalid page id or alias", nil).Status(http.StatusBadRequest)
	}

	m, err := loadRelationships(db, []string{data.PageAlias}, returnData, data.RestrictToMathDomain)
	if err != nil {
		return pages.Fail("error while loading links", err)
	}
	returnData.ResultMap["moreRelationshipIds"] = m[data.PageAlias]

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func loadRelationships(db *database.DB, aliases []string, returnData *core.CommonHandlerData, restrictToMathDomain bool) (map[string][]string, error) {
	query := database.NewQuery(`
		SELECT l.parentId, l.childAlias
		FROM links AS l
		JOIN `).AddPart(core.PageInfosTable(returnData.User)).Add(` AS pi
		ON (l.parentId=pi.pageId OR l.parentId=pi.alias)`)
	if restrictToMathDomain {
		query.Add(`
			JOIN domains AS d
			ON d.pageId=pi.editDomainId AND d.id=?`, core.MathDomainID)
	}
	query.Add(`WHERE l.childAlias IN`).AddIdsGroupStr(aliases)

	rows := query.ToStatement(db).Query()
	loadOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)
	relationships := make(map[string][]string) // alias -> ids of pages that link to it

	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentID, alias string
		if err := rows.Scan(&parentID, &alias); err != nil {
			return err
		}
		core.AddPageToMap(parentID, returnData.PageMap, loadOptions)
		relationships[alias] = append(relationships[alias], parentID)
		return nil
	})

	return relationships, err
}
