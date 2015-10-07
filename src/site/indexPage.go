// index.go serves the index page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

type featuredDomain struct {
	DomainAlias string
	DomainName  string
	ImageUrl    string
	ChildIds    []int64
}

// indexTmplData stores the data that we pass to the index.tmpl to render the page
type indexTmplData struct {
	commonPageData
	FeaturedDomains []*featuredDomain
}

// indexPage serves the index page.
var indexPage = newPageWithOptions(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/indexPage.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl",
		"tmpl/footer.tmpl"),
	pages.PageOptions{})

// indexRenderer renders the index page.
func indexRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data indexTmplData
	data.User = u

	// Manually load some pages we like
	if sessions.Live {
		data.FeaturedDomains = append(data.FeaturedDomains,
			&featuredDomain{
				DomainAlias: "vat",
				DomainName:  "Value Alignment Theory",
				ImageUrl:    "/static/images/vatIndexLogo.png",
				ChildIds: []int64{
					8639103000879599414, // VAT
					4213693741839491939, // List: value alignment subjects
					7722661858289734773, // Value alignment problem
					3158562585659930031, // Orthogonality theses
					6820582940749120623, // Ontology identification problem
					7879626441094927809, // Real world agents should be omni-safe
					5534008569097047764, // Mindcrime
					6053065048861201341, // Diamond maximizer
				},
			}, &featuredDomain{
				DomainAlias: "arbital",
				DomainName:  "Arbital",
				ImageUrl:    "/static/images/arbitalIndexLogo.png",
				ChildIds: []int64{
					3560540392275264633, // What is Arbital
					8992241719442104138, // Page hierarchy
					7894360080081727761, // Arbital wiki page
					5933317145970853046, // Editing
					4675907493088898985, // Liking
					8676677094741262267, // Voting
				},
			},
		)
	} else {
		data.FeaturedDomains = append(data.FeaturedDomains,
			&featuredDomain{
				DomainAlias: "dom",
				DomainName:  "Domain",
				ImageUrl:    "/static/images/vatIndexLogo.png",
				ChildIds: []int64{
					8587544546020587460,
					8329543642006081477,
					8719043452159528073,
					6507930240456827790,
				},
			}, &featuredDomain{
				DomainAlias: "dom",
				DomainName:  "Domain",
				ImageUrl:    "/static/images/arbitalIndexLogo.png",
				ChildIds: []int64{
					8587544546020587460,
					8329543642006081477,
					8719043452159528073,
					6507930240456827790,
				},
			},
		)
	}
	data.PageMap = make(map[int64]*core.Page)
	for _, domain := range data.FeaturedDomains {
		for _, pageId := range domain.ChildIds {
			data.PageMap[pageId] = &core.Page{PageId: pageId}
		}
	}

	// Load pages.
	err := core.LoadPages(db, data.PageMap, u.Id, &core.LoadPageOptions{})
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(db, u.Id, data.PageMap, nil)
	if err != nil {
		return pages.Fail("Couldn't load aux data", err)
	}

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err = loadGroupNames(db, u, data.GroupMap)
	if err != nil {
		return pages.Fail("Couldn't load group names", err)
	}

	return pages.StatusOK(&data)
}
