// claim.go serves the claim page.
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
	ClaimId     int64
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
	ChildId     int64
	CreatedAt   string
	UpdatedAt   string
	CreatorId   int64
	CreatorName string
}

type tag struct {
	Id   int64
	Text string
}

type claim struct {
	Id           int64
	CreatedAt    string
	UpdatedAt    string
	Text         string
	Url          string
	CreatorId    int64
	CreatorName  string
	PrivacyKey   sql.NullInt64
	IsSubscribed bool
	Tags         []*tag
}

type vote struct {
}

// claimTmplData stores the data that we pass to the index.tmpl to render the page
type claimTmplData struct {
	User   *user.User
	Claim  *claim
	Claims []*claim
	Inputs []*input
}

// claimPage serves the claim page.
var claimPage = pages.Add(
	"/claims/{id:[0-9]+}",
	claimRenderer,
	append(baseTmpls,
		"tmpl/claim.tmpl", "tmpl/input.tmpl", "tmpl/comment.tmpl", "tmpl/newComment.tmpl", "tmpl/navbar.tmpl")...)

var privateClaimPage = pages.Add(
	"/claims/{id:[0-9]+}/{privacyKey:[0-9]+}",
	claimRenderer,
	append(baseTmpls,
		"tmpl/claim.tmpl", "tmpl/input.tmpl", "tmpl/comment.tmpl", "tmpl/newComment.tmpl", "tmpl/navbar.tmpl")...)

// loadParentClaim loads and returns the parent claim.
func loadParentClaim(c sessions.Context, userId int64, claimId string) (*claim, error) {
	c.Infof("querying DB for claim with id = %s\n", claimId)
	claim := claim{}
	query := fmt.Sprintf(`
		SELECT id,text,creatorId,creatorName,createdAt,updatedAt,privacyKey
		FROM claims
		WHERE id=%s`, claimId)
	exists, err := database.QueryRowSql(c, query, &claim.Id, &claim.Text,
		&claim.CreatorId, &claim.CreatorName,
		&claim.CreatedAt, &claim.UpdatedAt, &claim.PrivacyKey)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a claim: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Unknown claim id: %s", claimId)
	}

	// Get subscription status.
	var useless int
	query = fmt.Sprintf(`
		SELECT 1
		FROM subscriptions
		WHERE userId=%d AND claimId=%s`, userId, claimId)
	claim.IsSubscribed, err = database.QueryRowSql(c, query, &useless)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve subscription status: %v", err)
	}

	// Load tags.
	err = loadTags(c, &claim)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve claim tags: %v", err)
	}

	return &claim, nil
}

// loadInputs loads and returns the inputs associated with the corresponding claim id from the db.
func loadInputs(c sessions.Context, db *sql.DB, claimId string) ([]*input, error) {
	inputs := make([]*input, 0)

	c.Infof("querying DB for inputs with parentId = %s\n", claimId)
	query := fmt.Sprintf(`
		SELECT id,childId,createdAt,updatedAt,creatorId,creatorName
		FROM inputs
		WHERE parentId=%s`, claimId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var i input
		err := rows.Scan(
			&i.Id,
			&i.ChildId,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CreatorId,
			&i.CreatorName)
		if err != nil {
			return fmt.Errorf("failed to scan for input: %v", err)
		}
		inputs = append(inputs, &i)
		return nil
	})
	return inputs, err
}

// loadChildClaims loads and returns all claims corresponding to the given inputs.
func loadChildClaims(c sessions.Context, userId int64, inputs []*input) ([]*claim, error) {
	claims := make([]*claim, 0, len(inputs))

	var buffer bytes.Buffer
	for _, i := range inputs {
		buffer.WriteString(strconv.FormatInt(i.Id, 10))
		buffer.WriteString(",")
	}
	claimIds := buffer.String()
	if claimIds == "" {
		return claims, nil
	}
	claimIds = claimIds[0 : len(claimIds)-1] // remove last comma

	c.Infof("querying DB for claims with ids = [%s]\n", claimIds)
	query := fmt.Sprintf(`
		SELECT id,text,creatorId,creatorName,createdAt,updatedAt,privacyKey
		FROM claims
		WHERE id IN (%s)`, claimIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var cl claim
		err := rows.Scan(
			&cl.Id,
			&cl.Text,
			&cl.CreatorId,
			&cl.CreatorName,
			&cl.CreatedAt,
			&cl.UpdatedAt,
			&cl.PrivacyKey)
		if err != nil {
			return fmt.Errorf("failed to scan for claim: %v", err)
		}
		claims = append(claims, &cl)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve claims: %v", err)
	}

	return claims, nil
}

// loadVisit returns the date (as string) when the user has seen this claim last.
// If the user has never seen this claim we return an empty string.
func loadVisit(c sessions.Context, userId int64, claimId string) (string, error) {
	updatedAt := ""
	query := fmt.Sprintf(`
		SELECT updatedAt
		FROM visits
		WHERE claimId=%s AND userId=%d`, claimId, userId)
	_, err := database.QueryRowSql(c, query, &updatedAt)
	return updatedAt, err
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

// claimRenderer renders the claim page.
func claimRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data claimTmplData
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

	// Load the parent claim
	claimId := mux.Vars(r)["id"]
	claim, err := loadParentClaim(c, data.User.Id, claimId)
	if err != nil {
		c.Inc("claim_fetch_fail")
		c.Errorf("error while fetching a claim: %v", err)
		return pages.InternalErrorWith(err)
	}
	data.Claim = claim

	// Check privacy setting
	if claim.PrivacyKey.Valid {
		privacyKey := mux.Vars(r)["privacyKey"]
		if privacyKey != fmt.Sprintf("%d", claim.PrivacyKey.Int64) {
			return pages.UnauthorizedWith(err)
		}
	}

	// Load all the inputs
	data.Inputs, err = loadInputs(c, db, claimId)
	if err != nil {
		c.Inc("inputs_fetch_fail")
		c.Errorf("error while fetching input for claim id: %s\n%v", claimId, err)
		return pages.InternalErrorWith(err)
	}

	// Load the child claim
	data.Claims, err = loadChildClaims(c, data.User.Id, data.Inputs)
	if err != nil {
		c.Inc("claims_fetch_fail")
		c.Errorf("error while fetching claims: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load last visit
	lastVisit := ""
	lastVisit, err = loadVisit(c, data.User.Id, claimId)
	if err != nil {
		c.Errorf("error while fetching a visit: %v", err)
	}

	/*
		// Load all the comments
		var comments map[int64]*comment
		var sortedCommentKeys []int64 // need this for in-order iteration
		comments, sortedCommentKeys, err = loadComments(c, db, inputIds)
		if err != nil {
			c.Inc("comments_fetch_fail")
			c.Errorf("error while fetching comments for claim id: %s\n%v", idStr, err)
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

	*/

	// Now that it looks like we are going to return the page successfully, we'll
	// mark all updates related to this claim as seen.
	query := fmt.Sprintf(`UPDATE updates SET seen=1 WHERE claimId=%s AND userId=%d`, claimId, data.User.Id)
	if _, err := database.ExecuteSql(c, query); err != nil {
		c.Errorf("Couldn't update updates: %v", err)
	}

	// Update last visit date.
	hashmap := make(map[string]interface{})
	hashmap["userId"] = data.User.Id
	hashmap["claimId"] = data.Claim.Id
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
	c.Inc("claim_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
