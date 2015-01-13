package dash

import (
	"database/sql"
	"flag"
	"net/http"
	"zanaduu3/src/database/dbcore"

	"github.com/hkjn/pages"
)

var (
	localDB = flag.Bool("local_db", false, "uses local MySQL database")

	activeUsers = &widget{
		Name: "Active users",
		Desc: "Shows number of active users",
		DBQuery: `SELECT
                IF (DATEDIFF(NOW(), lastActive)<=7,"active","inactive") AS "label",
                SUM(1) AS "users"
              FROM users
              GROUP BY 1 ORDER BY 1;`,
		newRow: func() *row {
			var (
				label string
				users int
			)
			return newRow(&label, &users)
		},
	}
	rewardsSeen = &widget{
		Name: "Rewards seen",
		Desc: "Shows number of rewards seen by different classes of users",
		DBQuery: `SELECT
                ur.seen AS "seen reward?",
                IF(ur.contest > 3, ">3", "<=3") AS "contests participated in",
                COUNT(DISTINCT ur.userId) AS users
              FROM (
                SELECT
                  MAX(seen) AS seen,
                  COUNT(DISTINCT contestId) AS contest,
                  userId
                FROM userRewards
                GROUP BY 3)
              AS ur
              GROUP BY 1,2 ORDER BY 2,1;`,
		newRow: func() *row {
			var (
				seen     bool
				contests string
				users    int
			)
			return newRow(&seen, &contests, &users)
		},
	}
	contestsEntered = &widget{
		Name: "Contests entered",
		Desc: "Shows user breakdown by number of contests entered",
		DBQuery: `SELECT
                IF(ur.contest > 3, "3+", ur.contest) AS "contests entered",
                COUNT(DISTINCT ur.userId) AS users
              FROM (
                SELECT
                  COUNT(DISTINCT contestId) AS "contest",
                  userId
                FROM userRewards
                GROUP BY 2)
              AS ur
              GROUP BY 1 ORDER BY 1;`,
		newRow: func() *row {
			var (
				class string
				users int
			)
			return newRow(&class, &users)
		},
	}
	seenRedeemed = &widget{
		Name: "Rewards seen / redeemed",
		Desc: "Shows user breakdown by seen / redeemed numbers for rewards",
		DBQuery: `SELECT
                ur.seen,
                ur.redeemed,
                COUNT(distinct ur.userId) AS "users"
              FROM (
                SELECT
                  MAX(seen) AS seen,
                  MAX(redeemed) AS redeemed,
                  userId
                FROM userRewards
                GROUP BY 3)
              AS ur
              GROUP BY 1,2;`,
		newRow: func() *row {
			var (
				seen, redeemed bool
				users          int
			)
			return newRow(&seen, &redeemed, &users)
		},
	}
)

// newRow returns a row for given values.
func newRow(v ...interface{}) *row {
	return &row{v}
}

type (
	// row represents the row of results of a DB query
	row struct {
		Values []interface{}
	}
	// widget shows DB results for the dashboard.
	widget struct {
		Name, DBQuery, Desc string
		EverQueried         bool
		Labels              []string    // labels for columns
		Rows                []*row      // rows of actual data
		newRow              func() *row // func to fetch row to scan into
	}
)

// query fills the widget with data queried from the DB.
func (w *widget) query(db *sql.DB) error {
	rows, err := db.Query(w.DBQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	w.Rows = []*row{}
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		w.Labels = cols
		r := w.newRow()
		err = rows.Scan(r.Values...)
		if err != nil {
			return err
		}
		w.Rows = append(w.Rows, r)
	}
	w.EverQueried = true
	return nil
}

func dashboardRenderer(w http.ResponseWriter, r *http.Request) pages.Result {
	c := pages.GetLogger(r)
	c.Infof("rendering /dash..\n")
	var (
		db  *sql.DB
		err error
	)
	if *localDB {
		db, err = dbcore.GetLocal(c)
	} else {
		db, err = dbcore.GetLive(c)
	}
	if err != nil {
		// TODO: handle returning errors for specific bits of the
		// dashboard without serving 500s, maybe?
		return pages.InternalErrorWith(err)
	}

	data := struct {
		baseData
		Widgets []*widget
	}{}

	data.Widgets = []*widget{
		activeUsers,
		rewardsSeen,
		contestsEntered,
		seenRedeemed,
	}
	for _, wd := range data.Widgets {
		err = wd.query(db)
		if err != nil {
			return pages.InternalErrorWith(err)
		}
	}
	return pages.StatusOK(data)
}
