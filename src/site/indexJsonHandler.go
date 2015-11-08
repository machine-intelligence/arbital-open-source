// indexJsonHandler.go serves the index page data.
package site

import (
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type featuredDomain struct {
	DomainId string   `json:"domainId"`
	ImageUrl string   `json:"imageUrl"`
	ChildIds []string `json:"childIds"`
}

var indexHandler = siteHandler{
	URI:         "/json/index/",
	HandlerFunc: indexJsonHandler,
}

func indexJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	returnData := newHandlerData(true)
	returnData.User = u

	// Manually load some pages we like
	featuredDomains := make([]*featuredDomain, 0)
	featuredDomains = append(featuredDomains,
		&featuredDomain{
			DomainId: "8639103000879599414",
			ImageUrl: "/static/images/vatIndexLogo.png",
			ChildIds: []string{
				"8639103000879599414", // VAT
				"4213693741839491939", // List: value alignment subjects
				"7722661858289734773", // Value alignment problem
				"3158562585659930031", // Orthogonality theses
				"6820582940749120623", // Ontology identification problem
				"7879626441094927809", // Real world agents should be omni-safe
				"5534008569097047764", // Mindcrime
				"6053065048861201341", // Diamond maximizer
			},
		}, &featuredDomain{
			DomainId: "3560540392275264633",
			ImageUrl: "/static/images/arbitalIndexLogo.png",
			ChildIds: []string{
				"3560540392275264633", // What is Arbital
				"8992241719442104138", // Page hierarchy
				"7894360080081727761", // Arbital wiki page
				"5933317145970853046", // Editing
				"4675907493088898985", // Liking
				"8676677094741262267", // Voting
			},
		},
	)
	returnData.ResultMap["featuredDomains"] = featuredDomains

	for _, domain := range featuredDomains {
		domainId, _ := strconv.ParseInt(domain.DomainId, 10, 64)
		core.AddPageToMap(domainId, returnData.PageMap, core.TitlePlusLoadOptions)
		for _, pageIdStr := range domain.ChildIds {
			pageId, _ := strconv.ParseInt(pageIdStr, 10, 64)
			core.AddPageToMap(pageId, returnData.PageMap, core.TitlePlusLoadOptions)
		}
	}

	// Load pages.
	err := core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
