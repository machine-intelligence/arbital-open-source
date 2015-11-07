// dynamicPage.go serves a page which then loads more data dynamically.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var (
	dynamicPage = newPageWithOptions("", dynamicPageRenderer, dynamicTmpls, pages.PageOptions{})
)

// dynamicPageRenderer renders the dynamic page.
func dynamicPageRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data commonPageData
	data.User = u
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	data.MasteryMap = make(map[int64]*core.Mastery)

	// Load pages.
	err := core.ExecuteLoadPipeline(db, u, data.PageMap, data.UserMap, data.MasteryMap)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.StatusOK(&data)
}
