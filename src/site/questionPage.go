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
	InputId     int64
	CreatedAt   string
	UpdatedAt   string
	Text        string
	ReplyToId   int64
	CreatorId   int64
	CreatorName string
	Replies     []*comment
}

type byDate []comment

func (a byDate) Len() int           { return len(a) }
func (a byDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDate) Less(i, j int) bool { return a[i].CreatedAt < a[j].CreatedAt }

type input struct {
	Id          int64
	CreatedAt   string
	UpdatedAt   string
	Text        string
	Url         string
	CreatorId   int64
	CreatorName string
	Comments    []*comment
	Vote        *vote
}

type tag struct {
	Id   int64
	Text string
}

type question struct {
	Id           int64
	CreatedAt    string
	UpdatedAt    string
	Text         string
	CreatorId    int64
	CreatorName  string
	PrivacyKey   sql.NullInt64
	IsSubscribed bool
	Tags         []*tag
	Answers      []string
	InputIds     []int64
	PriorVote    *priorVote
}

type priorVote struct {
	Value float32
}

type vote struct {
}

// questionTmplData stores the data that we pass to the index.tmpl to render the page
type questionTmplData struct {
	User     *user.User
	Question *question
	Priors   []*input
	Inputs   []*input
}

// questionPage serves the question page.
var questionPage = pages.Add(
	"/questions/{id:[0-9]+}",
	questionRenderer,
	append(baseTmpls,
		"tmpl/question.tmpl", "tmpl/input.tmpl", "tmpl/comment.tmpl", "tmpl/newComment.tmpl", "tmpl/navbar.tmpl")...)

var privateQuestionPage = pages.Add(
	"/questions/{id:[0-9]+}/{privacyKey:[0-9]+}",
	questionRenderer,
	append(baseTmpls,
		"tmpl/question.tmpl", "tmpl/input.tmpl", "tmpl/comment.tmpl", "tmpl/newComment.tmpl", "tmpl/navbar.tmpl")...)

// loadQuestion loads and returns the question with the corresponding id from the db.
func loadQuestion(c sessions.Context, userId int64, idStr string) (*question, error) {
	question := question{Answers: make([]string, 2, 2), InputIds: make([]int64, 2, 2)}
	var err error
	question.Id, err = strconv.ParseInt(idStr, 10, 63)
	if err != nil {
		return nil, fmt.Errorf("Incorrect id: %s", idStr)
	}

	c.Infof("querying DB for question with id = %s\n", idStr)
	query := fmt.Sprintf(`
		SELECT text,creatorId,creatorName,createdAt,updatedAt,privacyKey,answer1,answer2,inputId1,inputId2
		FROM questions
		WHERE id=%s`, idStr)
	exists, err := database.QueryRowSql(c, query, &question.Text,
		&question.CreatorId, &question.CreatorName,
		&question.CreatedAt, &question.UpdatedAt, &question.PrivacyKey,
		&question.Answers[0], &question.Answers[1],
		&question.InputIds[0], &question.InputIds[1])
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a question: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Unknown question id: %s", idStr)
	}

	// Get subscription status.
	var useless int
	query = fmt.Sprintf(`
		SELECT 1
		FROM subscriptions
		WHERE userId=%d AND questionId=%s`, userId, idStr)
	question.IsSubscribed, err = database.QueryRowSql(c, query, &useless)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve subscription status: %v", err)
	}

	// Load tags.
	err = loadTags(c, &question)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve question tags: %v", err)
	}

	return &question, nil
}

// loadVisit returns the date (as string) when the user has seen this question last.
// If the user has never seen this question we return an empty string.
func loadVisit(c sessions.Context, userId int64, idStr string) (string, error) {
	updatedAt := ""
	query := fmt.Sprintf(`
		SELECT updatedAt
		FROM visits
		WHERE questionId=%s AND userId=%d`, idStr, userId)
	_, err := database.QueryRowSql(c, query, &updatedAt)
	return updatedAt, err
}

// loadInputs loads and returns the inputs associated with the corresponding question id from the db.
func loadInputs(c sessions.Context, db *sql.DB, idStr string) ([]input, error) {
	inputs := make([]input, 0)

	c.Infof("querying DB for input with questionId = %s\n", idStr)
	query := fmt.Sprintf(`
		SELECT id,createdAt,updatedAt,text,url,creatorId,creatorName
		FROM inputs
		WHERE questionId=%s`, idStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var i input
		err := rows.Scan(
			&i.Id,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Text,
			&i.Url,
			&i.CreatorId,
			&i.CreatorName)
		if err != nil {
			return fmt.Errorf("failed to scan for input: %v", err)
		}
		inputs = append(inputs, i)
		return nil
	})
	return inputs, err
}

// loadComments loads and returns all the comments for the given input ids from the db.
func loadComments(c sessions.Context, db *sql.DB, inputIds string) (map[int64]*comment, []int64, error) {
	comments := make(map[int64]*comment)
	sortedCommentIds := make([]int64, 0)

	c.Infof("querying DB for comments with inputIds = %v", inputIds)
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT id,inputId,createdAt,updatedAt,text,replyToId,creatorId,creatorName
		FROM comments
		WHERE inputId IN (%s)`, inputIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var ct comment
		err := rows.Scan(
			&ct.Id,
			&ct.InputId,
			&ct.CreatedAt,
			&ct.UpdatedAt,
			&ct.Text,
			&ct.ReplyToId,
			&ct.CreatorId,
			&ct.CreatorName)
		if err != nil {
			return fmt.Errorf("failed to scan for comments: %v", err)
		}
		comments[ct.Id] = &ct
		sortedCommentIds = append(sortedCommentIds, ct.Id)
		return nil
	})
	return comments, sortedCommentIds, err
}

// loadPriorVote loads and returns the current user's most recent prior vote for this question.
func loadPriorVote(c sessions.Context, db *sql.DB, userId int64, questionId int64) (*priorVote, error) {
	c.Infof("querying DB for prior votes with userId=%v, questionId=%v", userId, questionId)
	var vote priorVote
	query := fmt.Sprintf(`
		SELECT value
		FROM priorVotes
		WHERE userId=%d AND questionId=%d
		ORDER BY createdAt DESC
		LIMIT 1`, userId, questionId)
	exists, err := database.QueryRowSql(c, query, &vote.Value)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a prior vote: %v", err)
	} else if !exists {
		return nil, nil
	}
	return &vote, nil
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

	// Load user, if possible
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the question
	idStr := mux.Vars(r)["id"]
	question, err := loadQuestion(c, data.User.Id, idStr)
	if err != nil {
		c.Inc("question_fetch_fail")
		c.Errorf("error while fetching question id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	data.Question = question

	// Check privacy setting
	if question.PrivacyKey.Valid {
		privacyKey := mux.Vars(r)["privacyKey"]
		if privacyKey != fmt.Sprintf("%d", question.PrivacyKey.Int64) {
			return pages.UnauthorizedWith(err)
		}
	}

	// Load last visit
	lastVisit := ""
	lastVisit, err = loadVisit(c, data.User.Id, idStr)
	if err != nil {
		c.Errorf("error while fetching a visit: %v", err)
	}

	// Load all the input
	var inputs []input
	inputs, err = loadInputs(c, db, idStr)
	if err != nil {
		c.Inc("input_fetch_fail")
		c.Errorf("error while fetching input for question id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	var buffer bytes.Buffer
	inputMap := make(map[int64]*input)
	data.Priors = make([]*input, 2, 2)
	data.Inputs = make([]*input, 0, len(inputs)-2)
	for i, s := range inputs {
		if s.Id == question.InputIds[0] {
			data.Priors[0] = &inputs[i]
		} else if s.Id == question.InputIds[1] {
			data.Priors[1] = &inputs[i]
		} else {
			data.Inputs = append(data.Inputs, &inputs[i])
		}
		buffer.WriteString(strconv.FormatInt(s.Id, 10))
		buffer.WriteString(",")
		inputMap[s.Id] = &inputs[i]
	}
	inputIds := buffer.String()
	inputIds = inputIds[0 : len(inputIds)-1] // remove last comma

	// Load all the comments
	var comments map[int64]*comment
	var sortedCommentKeys []int64 // need this for in-order iteration
	comments, sortedCommentKeys, err = loadComments(c, db, inputIds)
	if err != nil {
		c.Inc("comments_fetch_fail")
		c.Errorf("error while fetching comments for question id: %s\n%v", idStr, err)
		return pages.InternalErrorWith(err)
	}
	for _, key := range sortedCommentKeys {
		comment := comments[key]
		inputObj, ok := inputMap[comment.InputId]
		if !ok {
			c.Errorf("couldn't find input for a comment: %d\n%v", key, err)
			return pages.InternalErrorWith(err)
		}
		if comment.ReplyToId > 0 {
			parent := comments[comment.ReplyToId]
			parent.Replies = append(parent.Replies, comments[key])
		} else {
			inputObj.Comments = append(inputObj.Comments, comments[key])
		}
	}

	// Load prior vote
	data.Question.PriorVote, err = loadPriorVote(c, db, data.User.Id, data.Question.Id)
	if err != nil {
		c.Errorf("Couldn't load prior vote: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Now that it looks like we are going to return the page successfully, we'll
	// mark all updates related to this question as seen.
	query := fmt.Sprintf(`UPDATE updates SET seen=1 WHERE questionId=%s AND userId=%d`, idStr, data.User.Id)
	if _, err := database.ExecuteSql(c, query); err != nil {
		c.Errorf("Couldn't update updates: %v", err)
	}

	// Update last visit date.
	hashmap := make(map[string]interface{})
	hashmap["userId"] = data.User.Id
	hashmap["questionId"] = data.Question.Id
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	sql := database.GetInsertSql("visits", hashmap, "updatedAt")
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Errorf("Couldn't update visits: %v", err)
	}

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
		"IsNew": func(creatorId int64, createdAt string) bool {
			return creatorId != data.User.Id && lastVisit != "" && createdAt > lastVisit
		},
		"IsUpdated": func(creatorId int64, updatedAt string) bool {
			return creatorId != data.User.Id && lastVisit != "" && updatedAt > lastVisit
		},
	}
	c.Inc("question_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
