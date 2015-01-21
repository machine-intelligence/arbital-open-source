// question.go serves the question page.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

type comment struct {
	Id          int64
	SupportId   int64
	CreatedAt   string
	Text        string
	ReplyToId   sql.NullInt64
	CreatorId   int64
	CreatorName string
	Replies     []*comment
}

type byDate []comment

func (a byDate) Len() int           { return len(a) }
func (a byDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDate) Less(i, j int) bool { return a[i].CreatedAt < a[j].CreatedAt }

type support struct {
	Id          int64
	CreatedAt   string
	Text        string
	AnswerIndex int
	Prior       sql.NullInt64
	CreatorId   int64
	CreatorName string
	Comments    []*comment
	Vote        *vote
}

type question struct {
	Id          int64
	Text        string
	CreatorId   int64
	CreatorName string
}

type vote struct {
	Id        int64
	SupportId int64
	Value     float32
}

// questionTmplData stores the data that we pass to the index.tmpl to render the page
type questionTmplData struct {
	User     *user.User
	Question *question
	Priors   map[int]*support
	Support  map[int][]*support // answer -> []*support
	Error    string
}

// questionPage serves the question page.
var questionPage = pages.Add(
	"/questions/{id:[0-9]+}",
	questionRenderer,
	append(baseTmpls,
		"tmpl/question.tmpl", "tmpl/support.tmpl", "tmpl/comment.tmpl", "tmpl/newComment.tmpl")...)

// loadQuestion loads and returns the question with the correeponding id from the db.
func loadQuestion(c sessions.Context, idStr string) (*question, error) {
	var question question
	var err error
	question.Id, err = strconv.ParseInt(idStr, 10, 63)
	if err != nil {
		return nil, fmt.Errorf("Incorrect id: %s", idStr)
	}

	c.Infof("querying DB for question with id = %s\n", idStr)
	sql := fmt.Sprintf("SELECT text,creatorId,creatorName FROM questions WHERE id=%s", idStr)
	exists, err := database.QueryRowSql(c, sql, &question.Text, &question.CreatorId, &question.CreatorName)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a question: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Unknown question id: %s", idStr)
	}
	return &question, nil
}

// loadSupport loads and returns the support with the corresponding question id from the db.
func loadSupport(c sessions.Context, db *sql.DB, idStr string) ([]support, error) {
	supportSlice := make([]support, 0)

	c.Infof("querying DB for support with questionId = %s\n", idStr)
	rows, err := db.Query(`
		SELECT id,createdAt,text,answerIndex,prior,creatorId,creatorName
		FROM support
		WHERE questionId=?`, idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query for support: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s support
		err := rows.Scan(
			&s.Id,
			&s.CreatedAt,
			&s.Text,
			&s.AnswerIndex,
			&s.Prior,
			&s.CreatorId,
			&s.CreatorName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for support: %v", err)
		}
		supportSlice = append(supportSlice, s)
	}
	return supportSlice, nil
}

// loadComments loads and returns all the comments for the given support ids from the db.
func loadComments(c sessions.Context, db *sql.DB, supportIds string) (map[int64]*comment, []int64, error) {
	comments := make(map[int64]*comment)
	sortedCommentIds := make([]int64, 0)

	c.Infof("querying DB for comments with supportIds = %v", supportIds)
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT id,supportId,createdAt,text,replyToId,creatorId,creatorName
		FROM comments
		WHERE supportId IN (%s)`, supportIds)
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query for comments: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ct comment
		err := rows.Scan(
			&ct.Id,
			&ct.SupportId,
			&ct.CreatedAt,
			&ct.Text,
			&ct.ReplyToId,
			&ct.CreatorId,
			&ct.CreatorName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan for comments: %v", err)
		}
		comments[ct.Id] = &ct
		sortedCommentIds = append(sortedCommentIds, ct.Id)
	}
	return comments, sortedCommentIds, nil
}

// setPriorVotes loads and sets the current user's vote values for priors.
func setPriorVotes(c sessions.Context, db *sql.DB, userId int64, supportIds string, supportMap map[int64]*support) error {
	c.Infof("querying DB for prior votes with userId=%v, supportIds=%v", userId, supportIds)
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT id,supportId,value
		FROM priorVotes
		WHERE userId=%d AND supportId IN (%s)`, userId, supportIds)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query for votes: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v vote
		err := rows.Scan(
			&v.Id,
			&v.SupportId,
			&v.Value)
		if err != nil {
			return fmt.Errorf("failed to scan for votes: %v", err)
		}
		supportMap[v.SupportId].Vote = &v
	}
	return nil
}

// questionRenderer renders the question page.
func questionRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data questionTmplData
	c := sessions.NewContext(r)

	db, err := database.GetDB(c)
	if err != nil {
		c.Errorf("error while getting DB: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the question
	idStr := mux.Vars(r)["id"]
	question, err := loadQuestion(c, idStr)
	if err != nil {
		c.Inc("question_fetch_fail")
		c.Errorf("error while fetching question id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	data.Question = question

	// Load all the support
	var supportSlice []support
	supportSlice, err = loadSupport(c, db, idStr)
	if err != nil {
		c.Inc("support_fetch_fail")
		c.Errorf("error while fetching support for question id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	var buffer bytes.Buffer
	supportMap := make(map[int64]*support)
	data.Priors = make(map[int]*support)
	data.Support = make(map[int][]*support)
	for i, s := range supportSlice {
		if s.Prior.Valid {
			data.Priors[s.AnswerIndex] = &supportSlice[i]
		} else {
			data.Support[s.AnswerIndex] = append(data.Support[s.AnswerIndex], &supportSlice[i])
		}
		buffer.WriteString(strconv.FormatInt(s.Id, 10))
		buffer.WriteString(",")
		supportMap[s.Id] = &supportSlice[i]
	}
	supportIds := buffer.String()
	supportIds = supportIds[0 : len(supportIds)-1] // remove last comma

	// Load all the comments
	var comments map[int64]*comment
	var sortedCommentKeys []int64 // need this for in-order iteration
	comments, sortedCommentKeys, err = loadComments(c, db, supportIds)
	if err != nil {
		c.Inc("comments_fetch_fail")
		c.Errorf("error while fetching comments for question id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	for _, key := range sortedCommentKeys {
		comment := comments[key]
		supportObj, ok := supportMap[comment.SupportId]
		if !ok {
			c.Errorf("couldn't find support for a comment: %d\n%v", key, err)
			return pages.InternalErrorWith(err)
		}
		if comment.ReplyToId.Valid {
			parent := comments[comment.ReplyToId.Int64]
			parent.Replies = append(parent.Replies, comments[key])
		} else {
			supportObj.Comments = append(supportObj.Comments, comments[key])
		}
	}

	// Load user, if possible
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load all the votes and update priors
	err = setPriorVotes(c, db, data.User.Id, supportIds, supportMap)
	if err != nil {
		c.Errorf("Couldn't load votes: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		"UserId":  func() int64 { return data.User.Id },
		"IsAdmin": func() bool { return data.User.IsAdmin },
	}
	c.Inc("question_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
