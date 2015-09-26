// index.go serves the index page.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
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
		"tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl",
		"tmpl/footer.tmpl"),
	newPageOptions{})

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)

	data, err := indexInternalRenderer(w, r, u)
	if err != nil {
		c.Inc("index_page_served_fail")
		c.Errorf("%s", err)
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("index_page_served_success")
	return pages.StatusOK(data)
}

// indexInternalRenderer renders the index page.
func indexInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*indexTmplData, error) {
	var err error
	var data indexTmplData
	data.User = u
	c := sessions.NewContext(r)
	data.PageMap = make(map[int64]*core.Page)

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
					3560540392275264633, // What is Zanaduu
					8992241719442104138, // Page hierarchy
					7894360080081727761, // Zanaduu wiki page
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
	for _, domain := range data.FeaturedDomains {
		for _, pageId := range domain.ChildIds {
			data.PageMap[pageId] = &core.Page{PageId: pageId}
		}
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, &core.LoadPageOptions{})
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(c, u.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load aux data: %v", err)
	}

	// Load all the groups.
	data.GroupMap = make(map[int64]*core.Group)
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	return &data, nil
}
