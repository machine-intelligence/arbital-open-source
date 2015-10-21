// settingsPage.go serves the settings template.
package site

import (
	"zanaduu3/src/pages"
)

// settingsTmplData stores the data that we pass to the template to render the page
type settingsTmplData struct {
	commonPageData
}

// settingsPage serves the recent pages page.
var settingsPage = newPage(
	"/settings/",
	settingsRenderer,
	append(baseTmpls,
		"tmpl/settingsPage.tmpl", "tmpl/angular.tmpl.js"))

// settingsRenderer renders the page page.
func settingsRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U

	var data settingsTmplData
	data.User = u

	return pages.StatusOK(data)
}
