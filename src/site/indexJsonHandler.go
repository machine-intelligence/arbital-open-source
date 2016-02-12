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
	u := params.U

	returnData := newHandlerData(true)
	returnData.User = u

	// Manually load some pages we like
	featuredDomains := make([]*featuredDomain, 0)
	// HARDCODED
	/*
		featuredDomains = append(featuredDomains,
			&featuredDomain{
				DomainId: "8639103000879599414",
				ChildIds: []string{
					"8639103000879599414", // VAT
					"4213693741839491939", // List: value alignment subjects
					"7722661858289734773", // Value alignment problem
					"3158562585659930031", // Orthogonality theses
					"6820582940749120623", // Ontology identification problem
					"5534008569097047764", // Mindcrime
					"6053065048861201341", // Diamond maximizer
				},
			}, &featuredDomain{
				DomainId: "3560540392275264633",
				ChildIds: []string{
					"3560540392275264633", // What is Arbital
					"8138584842800103864", // Arbital blog
					"5092144177314150382", // Arbital features
					"8992241719442104138", // Parents and children
					"5933317145970853046", // Editing
					"4675907493088898985", // Liking
					"8676677094741262267", // Voting
				},
			},
		)
	*/
	featuredDomains = append(featuredDomains,
		&featuredDomain{
			DomainId: "30",
			ChildIds: []string{
				"30", // VAT
				"3r", // List: value alignment subjects
				"67", // Value alignment problem
				"1r", // Orthogonality theses
				"5r", // Ontology identification problem
				"7c", // Mindcrime
				"5v", // Diamond maximizer
			},
		}, &featuredDomain{
			DomainId: "3n",
			ChildIds: []string{
				"3n",  // What is Arbital
				"189", // Arbital blog
				"15m", // Arbital features
				"3y",  // Parents and children
				"45",  // Editing
				"40",  // Liking
				"42",  // Voting
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
	// Display this page fully
	// HARDCODED
	//core.AddPageToMap("3440973961008233681", returnData.PageMap, core.PrimaryPageLoadOptions)
	core.AddPageToMap("1m5", returnData.PageMap, core.PrimaryPageLoadOptions)

	db.C.Debugf("returnData.PageMap: %v", returnData.PageMap)

	// Load pages.
	err := core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)

	db.C.Debugf("returnData.PageMap: %v", returnData.PageMap)

	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	db.C.Debugf("returnData: %v", returnData)
	db.C.Debugf("returnData.toJson(): %v", returnData.toJson())

	return pages.StatusOK(returnData.toJson())
}
