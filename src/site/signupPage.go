// signupPage.go serves the signup page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// signupPage serves the signup page.
var signupPage = newPageWithOptions("", signupRenderer, dynamicTmpls, pages.PageOptions{})

// signupRenderer renders the signup page.
func signupRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	if u.Id <= 0 {
		return pages.RedirectWith(u.LoginLink)
	}

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
	return pages.StatusOK(data)
}
