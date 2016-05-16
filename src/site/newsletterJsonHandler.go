// newsletterJsonHandler.go serves the newsletter page data.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/mailchimp"
	"zanaduu3/src/pages"
)

var newsletterHandler = siteHandler{
	URI:         "/json/newsletter/",
	HandlerFunc: newsletterJsonHandler,
	Options:     pages.PageOptions{},
}

func newsletterJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	c := params.C
	returnData := core.NewHandlerData(u).SetResetEverything()
	var err error

	if u.Email != "" {
		u.MailchimpInterests, err = mailchimp.GetInterests(c, u.Email)
		if err != nil {
			return pages.Fail("Couldn't load mailchimp subscriptions", err)
		}
	}

	return pages.Success(returnData)
}
