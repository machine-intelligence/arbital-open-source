// pagePage.go serves the page page.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

type comment struct {
	Id           int64
	PageId       int64
	ReplyToId    int64
	Text         string
	CreatedAt    string
	UpdatedAt    string
	CreatorId    int64
	CreatorName  string
	LikeCount    int
	MyLikeValue  int
	IsSubscribed bool
	Replies      []*comment
}

// Helpers for soring comments by createdAt date.
type byDate []comment

func (a byDate) Len() int           { return len(a) }
func (a byDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDate) Less(i, j int) bool { return a[i].CreatedAt < a[j].CreatedAt }

// TODO: use this for context (and potentially list of links)
type input struct {
	Id          int64
	ChildId     int64
	CreatedAt   string
	UpdatedAt   string
	CreatorId   int64
	CreatorName string
}

type tag struct {
	// DB values.
	Id   int64
	Text string
}

type answer struct {
	// DB values.
	IndexId int
	Text    string
}

type richPage struct {
	// DB values.
	page
	LastVisit string

	// Computed values.
	InputCount   int
	IsSubscribed bool
	LikeCount    int
	DislikeCount int
	MyLikeValue  int
	VoteValue    float32
	VoteCount    int
	MyVoteValue  float32
	Contexts     []*richPage
	Comments     []*comment
}

// pageTmplData stores the data that we pass to the index.tmpl to render the page
type pageTmplData struct {
	User   *user.User
	Page   *richPage
	Pages  []*richPage
	Inputs []*input
}

// pagePage serves the page page.
var pagePage = pages.Add(
	"/pages/{id:[0-9]+}",
	pageRenderer,
	append(baseTmpls,
		"tmpl/pagePage.tmpl",
		"tmpl/comment.tmpl", "tmpl/newComment.tmpl",
		"tmpl/navbar.tmpl")...)

var privatePagePage = pages.Add(
	"/pages/{id:[0-9]+}/{privacyKey:[0-9]+}",
	pageRenderer,
	append(baseTmpls,
		"tmpl/pagePage.tmpl",
		"tmpl/comment.tmpl", "tmpl/newComment.tmpl",
		"tmpl/navbar.tmpl")...)

// loadMainPage loads and returns the main page.
func loadMainPage(c sessions.Context, userId int64, pageId int64) (*richPage, error) {
	c.Infof("querying DB for page with id = %d\n", pageId)

	pagePtr, err := loadFullPage(c, pageId)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	}
	mainPage := &richPage{page: *pagePtr}

	// Load contexts.
	/*query = fmt.Sprintf(`
		SELECT c.id,c.summary,c.text,c.privacyKey
		FROM inputs as i
		JOIN pages as c
		ON i.parentId=c.id
		WHERE i.childId=%d`, pageId)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var cl page
		err := rows.Scan(&cl.Id, &cl.Summary, &cl.Text, &cl.PrivacyKey)
		if err != nil {
			return fmt.Errorf("failed to scan for context page: %v", err)
		}
		mainPage.Contexts = append(mainPage.Contexts, &cl)
		return nil
	})*/
	return mainPage, err
}

// loadChildPages loads and returns all pages and inputs that have the given parent page.
/*func loadChildPages(c sessions.Context, pageId string) ([]*input, []*page, error) {
	inputs := make([]*input, 0)
	pages := make([]*page, 0)

	c.Infof("querying DB for child pages for parent id=%s\n", pageId)
	query := fmt.Sprintf(`
		SELECT i.id,i.childId,i.createdAt,i.updatedAt,i.creatorId,i.creatorName,
			c.id,c.summary,c.text,c.url,c.creatorId,c.creatorName,c.createdAt,c.updatedAt,c.privacyKey
		FROM inputs as i
		JOIN pages as c
		ON i.childId=c.id
		WHERE i.parentId=%s`, pageId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var i input
		var cl page
		err := rows.Scan(
			&i.Id, &i.ChildId,
			&i.CreatedAt, &i.UpdatedAt,
			&i.CreatorId, &i.CreatorName,
			&cl.Id, &cl.Summary, &cl.Text, &cl.Url,
			&cl.CreatorId, &cl.CreatorName,
			&cl.CreatedAt, &cl.UpdatedAt,
			&cl.PrivacyKey)
		if err != nil {
			return fmt.Errorf("failed to scan for input: %v", err)
		}
		inputs = append(inputs, &i)
		pages = append(pages, &cl)
		return nil
	})
	return inputs, pages, err
}*/

// loadInputCounts computes how many inputs each page has.
/*func loadInputCounts(c sessions.Context, pageIds string, pageMap map[int64]*page) error {
	query := fmt.Sprintf(`
		SELECT parentId,sum(1)
		FROM inputs
		WHERE parentId IN (%s)
		GROUP BY parentId`, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var parentId int64
		var count int
		err := rows.Scan(&parentId, &count)
		if err != nil {
			return fmt.Errorf("failed to scan for an input: %v", err)
		}
		pageMap[parentId].InputCount = count
		return nil
	})
	return err
}*/

// loadComments loads and returns all the comments for the given input ids from the db.
func loadComments(c sessions.Context, pageIds string) (map[int64]*comment, []int64, error) {
	commentMap := make(map[int64]*comment)
	sortedCommentIds := make([]int64, 0)

	c.Infof("querying DB for comments with pageIds = %v", pageIds)
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT id,pageId,replyToId,text,createdAt,updatedAt,creatorId,creatorName
		FROM comments
		WHERE pageId IN (%s)`, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var ct comment
		err := rows.Scan(
			&ct.Id,
			&ct.PageId,
			&ct.ReplyToId,
			&ct.Text,
			&ct.CreatedAt,
			&ct.UpdatedAt,
			&ct.CreatorId,
			&ct.CreatorName)
		if err != nil {
			return fmt.Errorf("failed to scan for comments: %v", err)
		}
		commentMap[ct.Id] = &ct
		sortedCommentIds = append(sortedCommentIds, ct.Id)
		return nil
	})
	return commentMap, sortedCommentIds, err
}

// loadLikes loads likes corresponding to the given pages and updates the pages.
func loadLikes(c sessions.Context, currentUserId int64, pageIds string, pageMap map[int64]*richPage) error {
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT userId,pageId,value
		FROM (
			SELECT *
			FROM likes
			WHERE pageId IN (%s)
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var userId int64
		var pageId int64
		var value int
		err := rows.Scan(&userId, &pageId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a like: %v", err)
		}
		page := pageMap[pageId]
		if value > 0 {
			page.LikeCount++
		} else if value < 0 {
			page.DislikeCount++
		}
		if userId == currentUserId {
			page.MyLikeValue = value
		}
		return nil
	})
	return err
}

// loadCommentLikes loads likes corresponding to the given comments and updates the comments.
func loadCommentLikes(c sessions.Context, currentUserId int64, commentIds string, commentMap map[int64]*comment) error {
	if len(commentIds) <= 0 {
		return nil
	}
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT userId,commentId,value
		FROM commentLikes
		WHERE commentId IN (%s)`, commentIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var userId int64
		var commentId int64
		var value int
		err := rows.Scan(&userId, &commentId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a comment like: %v", err)
		}
		comment := commentMap[commentId]
		if value > 0 {
			comment.LikeCount++
		}
		if userId == currentUserId {
			comment.MyLikeValue = value
		}
		return nil
	})
	return err
}

// loadVotes loads probability votes corresponding to the given pages and updates the pages.
func loadVotes(c sessions.Context, currentUserId int64, pageIds string, pageMap map[int64]*richPage) error {
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT userId,pageId,value
		FROM (
			SELECT *
			FROM votes
			WHERE pageId IN (%s) AND value>0
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var userId int64
		var pageId int64
		var value float32
		err := rows.Scan(&userId, &pageId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a vote: %v", err)
		}
		page := pageMap[pageId]
		page.VoteCount++
		page.VoteValue += value
		if userId == currentUserId {
			page.MyVoteValue = value
		}
		return nil
	})
	for _, p := range pageMap {
		if p.VoteCount > 0 {
			p.VoteValue /= float32(p.VoteCount)
		}
	}
	return err
}

// loadLastVisits loads lastVisit variable for each page.
func loadLastVisits(c sessions.Context, currentUserId int64, pageIds string, pageMap map[int64]*richPage) error {
	query := fmt.Sprintf(`
		SELECT pageId,updatedAt
		FROM visits
		WHERE userId=%d AND pageId IN (%s)`,
		currentUserId, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var updatedAt string
		err := rows.Scan(&pageId, &updatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a comment like: %v", err)
		}
		pageMap[pageId].LastVisit = updatedAt
		return nil
	})
	return err
}

// loadSubscriptions loads subscription statuses corresponding to the given
// pages and comments, and then updates the given maps.
func loadSubscriptions(
	c sessions.Context, currentUserId int64,
	pageIds string, commentIds string,
	pageMap map[int64]*richPage,
	commentMap map[int64]*comment) error {

	if len(commentIds) <= 0 {
		return nil
	}

	query := fmt.Sprintf(`
		SELECT pageId,commentId
		FROM subscriptions
		WHERE userId=%d AND (pageId IN (%s) OR commentId IN (%s))`,
		currentUserId, pageIds, commentIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var pageId int64
		var commentId int64
		err := rows.Scan(&pageId, &commentId)
		if err != nil {
			return fmt.Errorf("failed to scan for a comment like: %v", err)
		}
		if pageId > 0 {
			pageMap[pageId].IsSubscribed = true
		} else if commentId > 0 {
			commentMap[commentId].IsSubscribed = true
		}
		return nil
	})
	return err
}

// pageRenderer renders the page page.
func pageRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data pageTmplData
	c := sessions.NewContext(r)

	var err error
	/*db, err := database.GetDB(c)
	if err != nil {
		c.Errorf("error while getting DB: %v", err)
		return pages.InternalErrorWith(err)
	}*/

	// Load user, if possible
	data.User, err = user.LoadUserFromDb(c)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the parent page
	var pageId int64
	pageMap := make(map[int64]*richPage)
	pageIdStr := mux.Vars(r)["id"]
	pageId, err = strconv.ParseInt(pageIdStr, 10, 64)
	if err != nil {
		c.Inc("page_fetch_fail")
		c.Errorf("invalid id passed: %v", err)
		return pages.BadRequestWith(err)
	}
	mainPage, err := loadMainPage(c, data.User.Id, pageId)
	if err != nil {
		c.Inc("page_fetch_fail")
		c.Errorf("error while fetching a page: %v", err)
		return pages.InternalErrorWith(err)
	}
	pageMap[mainPage.PageId] = mainPage
	data.Page = mainPage

	// Check privacy setting
	if mainPage.PrivacyKey.Valid {
		privacyKey := mux.Vars(r)["privacyKey"]
		if privacyKey != fmt.Sprintf("%d", mainPage.PrivacyKey.Int64) {
			return pages.UnauthorizedWith(err)
		}
	}

	// Load all the inputs and corresponding child pages
	/*data.Inputs, data.Pages, err = loadChildPages(c, pageId)
	if err != nil {
		c.Inc("inputs_fetch_fail")
		c.Errorf("error while fetching input for page id: %s\n%v", pageId, err)
		return pages.InternalErrorWith(err)
	}

	// Get a string of all page ids and populate pageMap
	var buffer bytes.Buffer
	for _, c := range data.Pages {
		pageMap[c.Id] = c
		buffer.WriteString(fmt.Sprintf("%d", c.Id))
		buffer.WriteString(",")
	}
	buffer.WriteString(pageId)
	pageIds := buffer.String()

	// Load input counts
	err = loadInputCounts(c, pageIds, pageMap)
	if err != nil {
		c.Inc("inputs_fetch_fail")
		c.Errorf("error while fetching inputs: %v", err)
		return pages.InternalErrorWith(err)
	}*/

	// Load all the likes
	err = loadLikes(c, data.User.Id, pageIdStr, pageMap)
	if err != nil {
		c.Inc("likes_fetch_fail")
		c.Errorf("error while fetching likes: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load all the votes
	err = loadVotes(c, data.User.Id, pageIdStr, pageMap)
	if err != nil {
		c.Inc("votes_fetch_fail")
		c.Errorf("error while fetching votes: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Get last visits.
	q := r.URL.Query()
	forcedLastVisit := q.Get("lastVisit")
	if forcedLastVisit == "" {
		err = loadLastVisits(c, data.User.Id, pageIdStr, pageMap)
		if err != nil {
			c.Errorf("error while fetching a visit: %v", err)
		}
	} else {
		for _, cl := range pageMap {
			cl.LastVisit = forcedLastVisit
		}
	}

	// Load all the comments
	var commentMap map[int64]*comment // commentId -> comment
	var sortedCommentKeys []int64     // need this for in-order iteration
	commentMap, sortedCommentKeys, err = loadComments(c, pageIdStr)
	if err != nil {
		c.Inc("comments_fetch_fail")
		c.Errorf("error while fetching comments: %v", err)
		return pages.InternalErrorWith(err)
	}
	for _, key := range sortedCommentKeys {
		comment := commentMap[key]
		pageObj, ok := pageMap[comment.PageId]
		if !ok {
			c.Errorf("couldn't find page for a comment: %d\n%v", key, err)
			return pages.InternalErrorWith(err)
		}
		if comment.ReplyToId > 0 {
			parent := commentMap[comment.ReplyToId]
			parent.Replies = append(parent.Replies, commentMap[key])
		} else {
			pageObj.Comments = append(pageObj.Comments, commentMap[key])
		}
	}

	// Get a string of all comment ids.
	var buffer bytes.Buffer
	//buffer.Reset()
	for id, _ := range commentMap {
		buffer.WriteString(fmt.Sprintf("%d", id))
		buffer.WriteString(",")
	}
	commentIds := strings.TrimRight(buffer.String(), ",")

	// Load all the comment likes
	err = loadCommentLikes(c, data.User.Id, commentIds, commentMap)
	if err != nil {
		c.Inc("comment_likes_fetch_fail")
		c.Errorf("error while fetching comment likes: %v", err)
		return pages.InternalErrorWith(err)
	}

	if data.User.Id > 0 {
		// Load subscription statuses.
		err = loadSubscriptions(c, data.User.Id, pageIdStr, commentIds, pageMap, commentMap)
		if err != nil {
			c.Inc("subscriptions_fetch_fail")
			c.Errorf("error while fetching subscriptions: %v", err)
			return pages.InternalErrorWith(err)
		}

		// From here on we can render the page successfully. Further queries are nice,
		// but not mandatory, so we are not going to return an error if they fail.

		// Mark the relevant updates as read.
		query := fmt.Sprintf(
			`UPDATE updates
			SET seen=1,updatedAt='%s'
			WHERE pageId=%d AND userId=%d`,
			database.Now(), pageId, data.User.Id)
		if _, err := database.ExecuteSql(c, query); err != nil {
			c.Errorf("Couldn't update updates: %v", err)
		}

		// Update last visit date.
		values := ""
		for _, pg := range pageMap {
			values += fmt.Sprintf("(%d, %d, '%s', '%s'),",
				data.User.Id, pg.PageId, database.Now(), database.Now())
		}
		values = strings.TrimRight(values, ",")
		sql := fmt.Sprintf(`
			INSERT INTO visits (userId, pageId, createdAt, updatedAt)
			VALUES %s
			ON DUPLICATE KEY UPDATE updatedAt = VALUES(updatedAt)`, values)
		if _, err = database.ExecuteSql(c, sql); err != nil {
			c.Errorf("Couldn't update visits: %v", err)
		}

		// Load updates count.
		query = fmt.Sprintf(`
			SELECT COALESCE(SUM(count), 0)
			FROM updates
			WHERE userId=%d AND seen=0`, data.User.Id)
		_, err = database.QueryRowSql(c, query, &data.User.UpdateCount)
		if err != nil {
			c.Errorf("Couldn't retrieve updates count: %v", err)
		}
	}

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
		"IsUpdatedPage": func(p *richPage) bool {
			return p.CreatorId != data.User.Id && p.LastVisit != "" && p.CreatedAt >= p.LastVisit
		},
		"IsNewComment": func(c *comment) bool {
			lastVisit := pageMap[c.PageId].LastVisit
			return c.CreatorId != data.User.Id && lastVisit != "" && c.CreatedAt >= lastVisit
		},
		"IsUpdatedComment": func(c *comment) bool {
			lastVisit := pageMap[c.PageId].LastVisit
			return c.CreatorId != data.User.Id && lastVisit != "" && c.UpdatedAt >= lastVisit
		},
		// Check if we should even bother showing edit and delete page icons.
		"ShowEditIcons": func(p *richPage) bool {
			if data.User.IsAdmin {
				return getEditLevel(&p.page, data.User) >= 0 || getDeleteLevel(&p.page, data.User) >= 0
			}
			return getEditLevel(&p.page, data.User) > 0 || getDeleteLevel(&p.page, data.User) > 0
		},
		"GetEditLevel": func(p *richPage) int {
			return getEditLevel(&p.page, data.User)
		},
		"GetDeleteLevel": func(p *richPage) int {
			return getDeleteLevel(&p.page, data.User)
		},
		"GetPageUrl": func(p *richPage) string {
			privacyAddon := ""
			if p.PrivacyKey.Valid {
				privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey.Int64)
			}
			return fmt.Sprintf("/pages/%d%s", p.PageId, privacyAddon)
		},
		"GetPageEditUrl": func(p *richPage) string {
			privacyAddon := ""
			if p.PrivacyKey.Valid {
				privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey.Int64)
			}
			return fmt.Sprintf("/pages/edit/%d%s", p.PageId, privacyAddon)
		},
		"Sanitize": func(s string) template.HTML {
			s = template.HTMLEscapeString(s)
			s = strings.Replace(s, "\n", "<br>", -1)
			return template.HTML(s)
		},
	}
	c.Inc("page_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
