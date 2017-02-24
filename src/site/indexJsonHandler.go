// indexJsonHandler.go serves the index page data.

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var indexHandler = siteHandler{
	URI:         "/json/index/",
	HandlerFunc: indexJSONHandler,
	Options:     pages.PageOptions{},
}

func indexJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	if u.ID != "" {
		// Load what paths the user is on
		queryPart := database.NewQuery(`WHERE pathi.userId=?`, u.ID)
		err := core.LoadPathInstances(db, queryPart, u, func(db *database.DB, pathInstance *core.PathInstance) error {
			if pathInstance.GuideID == "3wj" {
				// Log guide
				if u.ContinueLogPath == nil || u.ContinueLogPath.Progress < pathInstance.Progress {
					u.ContinueLogPath = pathInstance
				}
			} else if pathInstance.GuideID == "61b" ||
				pathInstance.GuideID == "62c" ||
				pathInstance.GuideID == "62d" ||
				pathInstance.GuideID == "62f" {
				// Bayes guide
				if u.ContinueBayesPath == nil || u.ContinueBayesPath.Progress < pathInstance.Progress {
					u.ContinueBayesPath = pathInstance
				}
			}
			return nil
		})
		if err != nil {
			return pages.Fail("Couldn't load path instances", err)
		}
	}

	// Load pages.
	core.AddPageIDToMap("3hs", returnData.PageMap)
	core.AddPageIDToMap("4ym", returnData.PageMap)
	core.AddPageToMap("4c7", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("600", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("5xx", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("1ln", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("5zs", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("3rb", returnData.PageMap, core.TitlePlusLoadOptions)

	// ROGTODO: this is for the debate page: Is Eric's music better than Steph's?
	debatePageLoadOptions := (&core.PageLoadOptions{
		Children: true,
		// Tags:             true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap("7qq", returnData.PageMap, debatePageLoadOptions)

	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
