// adminDashboardPageJsonHandler.go serves JSON data to display admin dashboard page.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var adminDashboardPageHandler = siteHandler{
	URI:         "/json/adminDashboardPage/",
	HandlerFunc: adminDashboardPageJsonHandler,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// processRows from the given query and return an array containing the results
// also updating the pageMap as necessary.
func processRows(rows *database.Rows, pageMap map[string]*core.Page, loadOptions *core.PageLoadOptions, schema []string) ([][]string, error) {
	data := append(make([][]string, 0), schema)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		dataRow := make([]string, 5, 5)
		err := rows.Scan(&dataRow[0], &dataRow[1], &dataRow[2], &dataRow[3], &dataRow[4])
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		dataRow = dataRow[0:len(schema)]
		for n, schemaVal := range schema {
			if schemaVal[0] == '!' {
				core.AddPageToMap(dataRow[n], pageMap, loadOptions)
			}
		}
		data = append(data, dataRow)
		return nil
	})
	return data, err
}

// adminDashboardPageJsonHandler handles the request.
func adminDashboardPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()
	var err error

	// Load additional info for all pages
	pageOptions := core.TitlePlusLoadOptions

	// Monthly active users
	rows := database.NewQuery(`
		SELECT year(v.createdAt), month(v.createdAt), count(distinct u.id), count(distinct v.userId), count(distinct v.ipAddress)
		FROM visits AS v
		LEFT JOIN users as u ON v.userId=u.id
		WHERE v.createdAt>"2015-09-00"
		GROUP BY 1, 2
		ORDER BY 1 DESC, 2 DESC`).ToStatement(db).Query()
	returnData.ResultMap["monthly_active_users"], err = processRows(rows, returnData.PageMap, pageOptions, []string{
		"Year", "Month", "Registered Users", "Sessions", "IP Addresses",
	})
	if err != nil {
		return pages.Fail("Error while loading MAUs", err)
	}

	// Users who created at least one page
	rows = database.NewQuery(`
		SELECT year(createdAt),month(createdAt),count(distinct createdBy),-1,-1
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		WHERE createdAt>"2015-09-00" AND pageId!=createdBy AND type!=?`, core.CommentPageType).Add(`
		GROUP BY 1,2
		ORDER BY 1 DESC, 2 DESC`).ToStatement(db).Query()
	returnData.ResultMap["users_who_created_at_least_one_page"], err = processRows(rows, returnData.PageMap, pageOptions, []string{
		"Year", "Month", "Count",
	})
	if err != nil {
		return pages.Fail("Error while loading page-creators", err)
	}

	// Users who created at least 5 pages in the last month
	rows = database.NewQuery(`
		SELECT pi.createdBy,concat(u.firstName," ",u.lastName),-1,-1,-1
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		JOIN users AS u
		ON (pi.createdBy=u.id)
		WHERE TIMESTAMPDIFF(DAY,pi.createdAt,NOW())<=30 AND type!=?`, core.CommentPageType).Add(`
		GROUP BY 1
		HAVING COUNT(*)>=5`).ToStatement(db).Query()
	returnData.ResultMap["users_who_created_at_least_5_pages_in_the_last_month"], err = processRows(rows, returnData.PageMap, pageOptions, []string{
		"UserId", "UserName",
	})
	if err != nil {
		return pages.Fail("Error while loading page-creators", err)
	}

	// New users
	rows = database.NewQuery(`
		SELECT year(createdAt),month(createdAt),count(*),-1,-1
		FROM users
		WHERE createdAt>"2015-09-00"
		GROUP BY 1,2
		ORDER BY 1 DESC, 2 DESC`).ToStatement(db).Query()
	returnData.ResultMap["new_users"], err = processRows(rows, returnData.PageMap, pageOptions, []string{
		"Year", "Month", "Count",
	})
	if err != nil {
		return pages.Fail("Error while loading page-creators", err)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
