// pagesJsonHandler.go contains the handler for returning JSON with pages data.
package site

/*import (
	"fmt"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"

	"github.com/gorilla/schema"
)

// pagesJsonData contains parameters passed in via the request.
type pagesJsonData struct {
	PageAliases []string
	// Load entire page text
	IncludeText bool
	// Load auxillary data: likes, votes, subscription
	IncludeAuxData   bool
	LoadComments     bool
	LoadVotes        bool
	LoadChildren     bool
	LoadChildDraft   bool
	LoadRequirements bool
}

// pagesJsonHandler handles the request.
func pagesJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data pagesJsonData
	params.R.ParseForm()
	err := schema.NewDecoder().Decode(&data, params.R.Form)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	// If no page ids, return a new random page id.
	if len(data.PageAliases) <= 0 {
		return pages.HandlerBadRequestFail("No page ids/aliases were specified.", nil)
	}

	returnData, message, err := pagesJsonHandlerInternal(params, &data)
	if returnData == nil {
		return pages.HandlerErrorFail(message, err)
	}
	return pages.StatusOK(returnData)
}

// pagesJsonHandler handles the request.
func pagesJsonHandlerInternal(params *pages.HandlerParams, data *pagesJsonData) (map[string]interface{}, string, error) {
	db := params.DB
	u := params.U
	returnData := make(map[string]interface{})

	// Convert all aliases to ids
	pageIds := make([]int64, 0)
	strAliases := make([]interface{}, 0)
	for _, alias := range data.PageAliases {
		pageId, err := strconv.ParseInt(alias, 10, 64)
		if err == nil {
			pageIds = append(pageIds, pageId)
		} else {
			strAliases = append(strAliases, alias)
		}
	}

	// Convert actual aliases into page ids
	if len(strAliases) > 0 {
		rows := database.NewQuery(`
			SELECT pageId
			FROM pages
			WHERE isCurrentEdit AND alias IN`).AddArgsGroup(strAliases).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan for original createdAt: %v", err)
			}
			pageIds = append(pageIds, pageId)
			return nil
		})
		if err != nil {
			return nil, "couldn't convert aliases to page ids", err
		}
	}
	if len(pageIds) <= 0 {
		return nil, "All of the passed in aliases weren't found.", nil
	}

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)
	masteryMap := make(map[int64]*core.Mastery)
	sourceMap := make(map[int64]*core.Page)

	// Process pageIds
	for _, pageId := range pageIds {
		p := core.AddPageIdToMap(pageId, pageMap)
		sourceMap[pageId] = p
	}

	// Load comment ids.
	if data.LoadComments {
		err := core.LoadSubpageIds(db, pageMap, sourceMap)
		if err != nil {
			return nil, "Couldn't load subpages", err
		}
	}

	// Load links
	err := core.LoadLinks(db, pageMap, nil)
	if err != nil {
		return nil, "Couldn't load links", err
	}

	// Load children
	if data.LoadChildren {
		err := core.LoadChildrenIds(db, pageMap, &core.LoadChildrenIdsOptions{ForPages: sourceMap})
		if err != nil {
			return nil, "Couldn't load children", err
		}
	}

	// Load requirements
	if data.LoadRequirements {
		err := core.LoadRequirements(db, u.Id, pageMap, masteryMap, &core.LoadChildrenIdsOptions{ForPages: sourceMap})
		if err != nil {
			return nil, "Couldn't load children", err
		}
	}

	// Load page data
	err = core.LoadPages(db, pageMap, u, &core.LoadPageOptions{LoadText: true, LoadSummary: true})
	if err != nil {
		return nil, "error while loading pages", err
	}

	// Load the auxillary data.
	if data.IncludeAuxData {
		err := core.LoadAuxPageData(db, u.Id, pageMap, nil)
		if err != nil {
			return nil, "error while loading aux data", err
		}
	}

	// Load probability votes
	if data.LoadVotes {
		err := core.LoadVotes(db, u.Id, pageMap, userMap)
		if err != nil {
			return nil, "Couldn't load probability votes", err
		}
	}

	if data.LoadChildDraft {
		// Load child draft
		for _, p := range pageMap {
			if p.Type == core.CommentPageType {
				continue
			}
			err := core.LoadChildDraft(db, u.Id, p, pageMap)
			if err != nil {
				return nil, "Couldn't load child draft", err
			}
			break
		}
	}

	// Load all the users
	for _, p := range pageMap {
		userMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, userMap)
	if err != nil {
		return nil, "error while loading users", err
	}

	// Erase the text for pages if necessary; otherwise, mark them as visited
	// TODO: keep text only for pages in sourceMap
	visitedValues := make([]interface{}, 0)
	for k, v := range pageMap {
		if !data.IncludeText {
			v.Text = ""
		} else {
			visitedValues = append(visitedValues, u.Id, k, database.Now())
		}
	}

	// Add a visit to pages for which we loaded text.
	if len(visitedValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO visits (userId, pageId, createdAt)
			VALUES ` + database.ArgsPlaceholder(len(visitedValues), 3))
		if _, err = statement.Exec(visitedValues...); err != nil {
			return nil, "Couldn't update visits", err
		}
	}

	returnData = createReturnData(pageMap).AddUsers(userMap).AddMasteries(masteryMap)
	return returnData, "", nil
}*/
