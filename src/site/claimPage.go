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
	Id      int64
	ClaimId int64
	//ContextClaimId int64
	ReplyToId   int64
	Text        string
	CreatedAt   string
	UpdatedAt   string
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
	Id            int64
	CreatedAt     string
	UpdatedAt     string
	Text          string
	Url           string
	CreatorId     int64
	CreatorName   string
	PrivacyKey    sql.NullInt64
	InputCount    int
	IsSubscribed  bool
	UpvoteCount   int
	DownvoteCount int
	MyVote        int
	Contexts      []*claim
	Comments      []*comment
	Tags          []*tag
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
		"tmpl/claimPage.tmpl", "tmpl/claim.tmpl",
		"tmpl/comment.tmpl", "tmpl/newComment.tmpl",
		"tmpl/vote.tmpl", "tmpl/navbar.tmpl")...)

var privateClaimPage = pages.Add(
	"/claims/{id:[0-9]+}/{privacyKey:[0-9]+}",
	claimRenderer,
	append(baseTmpls,
		"tmpl/claimPage.tmpl", "tmpl/claim.tmpl",
		"tmpl/comment.tmpl", "tmpl/newComment.tmpl",
		"tmpl/vote.tmpl", "tmpl/navbar.tmpl")...)

// loadParentClaim loads and returns the parent claim.
func loadParentClaim(c sessions.Context, userId int64, claimId string) (*claim, error) {
	c.Infof("querying DB for claim with id = %s\n", claimId)
	parentClaim := &claim{}
	query := fmt.Sprintf(`
		SELECT id,text,url,creatorId,creatorName,createdAt,updatedAt,privacyKey
		FROM claims
		WHERE id=%s`, claimId)
	exists, err := database.QueryRowSql(c, query, &parentClaim.Id, &parentClaim.Text, &parentClaim.Url,
		&parentClaim.CreatorId, &parentClaim.CreatorName,
		&parentClaim.CreatedAt, &parentClaim.UpdatedAt, &parentClaim.PrivacyKey)
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
	parentClaim.IsSubscribed, err = database.QueryRowSql(c, query, &useless)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve subscription status: %v", err)
	}

	// Load tags.
	err = loadTags(c, parentClaim)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve claim tags: %v", err)
	}

	// Load contexts.
	query = fmt.Sprintf(`
		SELECT c.id,c.text,c.privacyKey
		FROM inputs as i
		JOIN claims as c
		ON i.parentId=c.id
		WHERE i.childId=%s`, claimId)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var cl claim
		err := rows.Scan(&cl.Id, &cl.Text, &cl.PrivacyKey)
		if err != nil {
			return fmt.Errorf("failed to scan for context claim: %v", err)
		}
		parentClaim.Contexts = append(parentClaim.Contexts, &cl)
		return nil
	})
	return parentClaim, err
}

// loadChildClaims loads and returns all claims and inputs that have the given parent claim.
func loadChildClaims(c sessions.Context, claimId string) ([]*input, []*claim, error) {
	inputs := make([]*input, 0)
	claims := make([]*claim, 0)

	c.Infof("querying DB for child claims for parent id=%s\n", claimId)
	query := fmt.Sprintf(`
		SELECT i.id,i.childId,i.createdAt,i.updatedAt,i.creatorId,i.creatorName,
			c.id,c.text,c.url,c.creatorId,c.creatorName,c.createdAt,c.updatedAt,c.privacyKey
		FROM inputs as i
		JOIN claims as c
		ON i.childId=c.id
		WHERE i.parentId=%s`, claimId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var i input
		var cl claim
		err := rows.Scan(
			&i.Id, &i.ChildId,
			&i.CreatedAt, &i.UpdatedAt,
			&i.CreatorId, &i.CreatorName,
			&cl.Id, &cl.Text, &cl.Url,
			&cl.CreatorId, &cl.CreatorName,
			&cl.CreatedAt, &cl.UpdatedAt,
			&cl.PrivacyKey)
		if err != nil {
			return fmt.Errorf("failed to scan for input: %v", err)
		}
		inputs = append(inputs, &i)
		claims = append(claims, &cl)
		return nil
	})
	return inputs, claims, err
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

// loadInputCounts computes how many inputs each claim has.
func loadInputCounts(c sessions.Context, claimIds string, claimMap map[int64]*claim) error {
	query := fmt.Sprintf(`
		SELECT parentId,sum(1)
		FROM inputs
		WHERE parentId IN (%s)
		GROUP BY parentId`, claimIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var parentId int64
		var count int
		err := rows.Scan(&parentId, &count)
		if err != nil {
			return fmt.Errorf("failed to scan for an input: %v", err)
		}
		claimMap[parentId].InputCount = count
		return nil
	})
	return err
}

// loadComments loads and returns all the comments for the given input ids from the db.
func loadComments(c sessions.Context, claimIds string) (map[int64]*comment, []int64, error) {
	comments := make(map[int64]*comment)
	sortedCommentIds := make([]int64, 0)

	c.Infof("querying DB for comments with claimIds = %v", claimIds)
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT id,claimId,replyToId,text,createdAt,updatedAt,creatorId,creatorName
		FROM comments
		WHERE claimId IN (%s)`, /*AND (contextClaimId=0 OR contextClaimId=%s)`*/ claimIds /*, claimId*/)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var ct comment
		err := rows.Scan(
			&ct.Id,
			&ct.ClaimId,
			//&ct.ContextClaimId,
			&ct.ReplyToId,
			&ct.Text,
			&ct.CreatedAt,
			&ct.UpdatedAt,
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

// loadVotes loads votes corresponding to the given claims and updates the claims.
func loadVotes(c sessions.Context, currentUserId int64, claimIds string, claimMap map[int64]*claim) error {
	// Workaround for: https://github.com/go-sql-driver/mysql/issues/304
	query := fmt.Sprintf(`
		SELECT userId,claimId,value
		FROM (SELECT * FROM votes ORDER BY id DESC) AS v
		WHERE claimId IN (%s)
		GROUP BY userId,claimId`, claimIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var userId int64
		var claimId int64
		var value int
		err := rows.Scan(&userId, &claimId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a vote: %v", err)
		}
		claim := claimMap[claimId]
		if value > 0 {
			claim.UpvoteCount++
		} else if value < 0 {
			claim.DownvoteCount++
		}
		if userId == currentUserId {
			claim.MyVote = value
		}
		return nil
	})
	return err
}

// claimRenderer renders the claim page.
func claimRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data claimTmplData
	c := sessions.NewContext(r)

	var err error
	/*db, err := database.GetDB(c)
	if err != nil {
		c.Errorf("error while getting DB: %v", err)
		return pages.InternalErrorWith(err)
	}*/

	// Load user, if possible
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the parent claim
	claimMap := make(map[int64]*claim)
	claimId := mux.Vars(r)["id"]
	parentClaim, err := loadParentClaim(c, data.User.Id, claimId)
	if err != nil {
		c.Inc("claim_fetch_fail")
		c.Errorf("error while fetching a claim: %v", err)
		return pages.InternalErrorWith(err)
	}
	claimMap[parentClaim.Id] = parentClaim
	data.Claim = parentClaim

	// Check privacy setting
	if parentClaim.PrivacyKey.Valid {
		privacyKey := mux.Vars(r)["privacyKey"]
		if privacyKey != fmt.Sprintf("%d", parentClaim.PrivacyKey.Int64) {
			return pages.UnauthorizedWith(err)
		}
	}

	// Load all the inputs and corresponding child claims
	data.Inputs, data.Claims, err = loadChildClaims(c, claimId)
	if err != nil {
		c.Inc("inputs_fetch_fail")
		c.Errorf("error while fetching input for claim id: %s\n%v", claimId, err)
		return pages.InternalErrorWith(err)
	}
	for _, c := range data.Claims {
		claimMap[c.Id] = c
	}

	// Load last visit
	lastVisit := ""
	lastVisit, err = loadVisit(c, data.User.Id, claimId)
	if err != nil {
		c.Errorf("error while fetching a visit: %v", err)
	}

	// Get a string of all claim ids.
	var buffer bytes.Buffer
	for _, c := range data.Claims {
		buffer.WriteString(strconv.FormatInt(c.Id, 10))
		buffer.WriteString(",")
	}
	buffer.WriteString(claimId)
	claimIds := buffer.String()

	// Load input counts
	err = loadInputCounts(c, claimIds, claimMap)
	if err != nil {
		c.Inc("inputs_fetch_fail")
		c.Errorf("error while fetching inputs: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load all the comments
	var comments map[int64]*comment
	var sortedCommentKeys []int64 // need this for in-order iteration
	comments, sortedCommentKeys, err = loadComments(c, claimIds)
	if err != nil {
		c.Inc("comments_fetch_fail")
		c.Errorf("error while fetching comments: %v", err)
		return pages.InternalErrorWith(err)
	}
	for _, key := range sortedCommentKeys {
		comment := comments[key]
		claimObj, ok := claimMap[comment.ClaimId]
		if !ok {
			c.Errorf("couldn't find claim for a comment: %d\n%v", key, err)
			return pages.InternalErrorWith(err)
		}
		if comment.ReplyToId > 0 {
			parent := comments[comment.ReplyToId]
			parent.Replies = append(parent.Replies, comments[key])
		} else {
			claimObj.Comments = append(claimObj.Comments, comments[key])
		}
	}

	// Load all the votes
	err = loadVotes(c, data.User.Id, claimIds, claimMap)
	if err != nil {
		c.Inc("votes_fetch_fail")
		c.Errorf("error while fetching votes: %v", err)
		return pages.InternalErrorWith(err)
	}

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
		"GetClaimUrl": func(c *claim) string {
			privacyAddon := ""
			if c.PrivacyKey.Valid {
				privacyAddon = fmt.Sprintf("/%d", c.PrivacyKey.Int64)
			}
			return fmt.Sprintf("/claims/%d%s", c.Id, privacyAddon)
		},
	}
	c.Inc("claim_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
