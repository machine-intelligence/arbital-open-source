// pageHelpers.go contains the page struct as well as helpful functions.
package site

import (
	"database/sql"
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

const (
	blogPageType         = "blog"
	questionPageType     = "question"
	infoPageType         = "info"
	maxKarmaLockFraction = 0.8
)

type page struct {
	// Data loaded from pages table
	PageId      int64
	Type        string
	Title       string
	Text        string
	CreatorId   int64
	CreatorName string
	CreatedAt   string
	KarmaLock   int
	PrivacyKey  sql.NullInt64
	DeletedBy   int64

	// Additional data.
	Tags    []*tag
	Answers []*answer
}

// loadFullPage loads and retuns a page. It also loads all the auxillary data like tags and answers.
func loadFullPage(c sessions.Context, pageId int64) (*page, error) {
	pagePtr, err := loadPage(c, pageId)
	if err != nil {
		return nil, err
	}
	err = pagePtr.loadTags(c)
	if err != nil {
		return nil, err
	}
	err = pagePtr.loadAnswers(c)
	if err != nil {
		return nil, err
	}
	return pagePtr, nil
}

// loadPage loads and returns a page with the given id from the database.
func loadPage(c sessions.Context, pageId int64) (*page, error) {
	var p page
	query := fmt.Sprintf(`
		SELECT pageId,type,title,text,creatorId,creatorName,createdAt,karmaLock,privacyKey,deletedBy
		FROM pages
		WHERE pageId=%d
		ORDER BY id DESC
		LIMIT 1`, pageId)
	exists, err := database.QueryRowSql(c, query, &p.PageId,
		&p.Type, &p.Title, &p.Text,
		&p.CreatorId, &p.CreatorName,
		&p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.DeletedBy)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Unknown page id: %d", pageId)
	}
	return &p, nil
}

// loadTags loads tags corresponding to this page.
func (p *page) loadTags(c sessions.Context) error {
	query := fmt.Sprintf(`
		SELECT t.id,t.Text
		FROM pageTagPairs AS p
		LEFT JOIN tags AS t
		ON p.tagId=t.Id
		WHERE p.pageId=%d`, p.PageId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var t tag
		err := rows.Scan(&t.Id, &t.Text)
		if err != nil {
			return fmt.Errorf("failed to scan for pageTagPair: %v", err)
		}
		p.Tags = append(p.Tags, &t)
		return nil
	})
	return err
}

// loadAnswers loads answers corresponding to this page.
func (p *page) loadAnswers(c sessions.Context) error {
	if p.Type != questionPageType {
		return nil
	}

	query := fmt.Sprintf(`
		SELECT indexId,text
		FROM answers
		WHERE pageId=%d
		ORDER BY indexId`, p.PageId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var a answer
		err := rows.Scan(&a.IndexId, &a.Text)
		if err != nil {
			return fmt.Errorf("failed to scan for an answer: %v", err)
		}
		p.Answers = append(p.Answers, &a)
		return nil
	})
	return err
}

const ()

// getMaxKarmaLock returns the highest possible karma lock a user with the
// given amount of karma can create.
func getMaxKarmaLock(karma int) int {
	return int(float32(karma) * maxKarmaLockFraction)
}

// Check if the user can edit this page. -1 = no, 0 = only as admin, 1 = yes
func getEditLevel(p *page, u *user.User) int {
	if p.Type == blogPageType {
		if p.CreatorId == u.Id {
			return 1
		} else {
			return -1
		}
	}
	if u.Karma >= p.KarmaLock {
		return 1
	} else if u.IsAdmin {
		return 0
	}
	return -1
}

// Check if the user can delete this page. -1 = no, 0 = only as admin, 1 = yes
func getDeleteLevel(p *page, u *user.User) int {
	if p.Type == blogPageType {
		if p.CreatorId == u.Id {
			return 1
		} else if u.IsAdmin {
			return 0
		} else {
			return -1
		}
	}
	if u.Karma >= p.KarmaLock {
		return 1
	} else if u.IsAdmin {
		return 0
	}
	return -1
}
