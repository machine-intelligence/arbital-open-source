// indexJsonHandler.go serves the index page data.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type featuredDomain struct {
	DomainId string   `json:"domainId"`
	ChildIds []string `json:"childIds"`
}

var indexHandler = siteHandler{
	URI:         "/json/index/",
	HandlerFunc: indexJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
	},
}

func indexJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	returnData := core.NewHandlerData(params.U, true)

	// Manually load some pages we like
	featuredDomains := make([]*featuredDomain, 0)
	// HARDCODED
	featuredDomains = append(featuredDomains,
		&featuredDomain{
			DomainId: "2v", // Value alignment theory
			ChildIds: []string{
				"2v", // VAT
				"3g", // List: value alignment subjects
				"5s", // Value alignment problem
				"1y", // Orthogonality theses
				"5c", // Ontology identification problem
				"6v", // Mindcrime
				"5g", // Diamond maximizer
			},
		}, &featuredDomain{
			DomainId: "3d", // Arbital
			ChildIds: []string{
				"3d",  // What is Arbital
				"16q", // Arbital blog
				"14q", // Arbital features
				"3n",  // Parents and children
				"3v",  // Editing
				"3p",  // Liking
				"3r",  // Voting
			},
		}, &featuredDomain{
			DomainId: "1lw", // Mathematics
			ChildIds: []string{
				"1mg", // Arithmetical hierarchy
				"1zq", // Bayes' rule: Guide
				"220", // Bayesian view of scientific virtues
				"1zg", // Multiple hypotheses and updates
				"1x5", // Bayes's Rule (odds form)
				"1wy", // Waterfall and frequency diagrams
			},
		},
	)

	returnData.ResultMap["featuredDomains"] = featuredDomains

	for _, domain := range featuredDomains {
		core.AddPageToMap(domain.DomainId, returnData.PageMap, core.TitlePlusLoadOptions)
		for _, pageIdStr := range domain.ChildIds {
			core.AddPageToMap(pageIdStr, returnData.PageMap, core.TitlePlusLoadOptions)
		}
	}
	// HARDCODED
	core.AddPageToMap("1k0", returnData.PageMap, core.TitlePlusLoadOptions)

	// Load pages.
	err := core.ExecuteLoadPipeline(db, returnData)

	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.ToJson())
}
