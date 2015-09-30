// pageHelpers.go contains the page struct as well as helpful functions.
package site

import (
	"fmt"
	"strconv"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

type loadEditOptions struct {
	// If true, the last edit will be loaded for the given user, even if it's an
	// autosave or a snapshot.
	loadNonliveEdit bool
	// Don't convert loaded parents string into an array of parents
	ignoreParents bool

	// If set, we'll load this edit of the page
	loadSpecificEdit int
	// If set, we'll only load from edits less than this
	loadEditWithLimit int
	// If set, we'll only load from edits with createdAt timestamp before this
	createdAtLimit string
}

// loadFullEdit loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If userId is given, the last edit of the given pageId will be returned. It
// might be an autosave or a snapshot, and thus not the current live page.
// If the page couldn't be found, (nil, nil) will be returned.
func loadFullEdit(db *database.DB, pageId, userId int64, options *loadEditOptions) (*core.Page, error) {
	if options == nil {
		options = &loadEditOptions{}
	}
	var p core.Page

	whereClause := database.NewQuery("p.isCurrentEdit")
	if options.loadSpecificEdit > 0 {
		whereClause = database.NewQuery("p.edit=?", options.loadSpecificEdit)
	} else if options.loadNonliveEdit {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT MAX(edit)
				FROM pages
				WHERE pageId=? AND deletedBy<=0 AND (creatorId=? OR NOT (isSnapshot OR isAutosave))
			)`, pageId, userId)
	} else if options.loadEditWithLimit > 0 {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND edit<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.loadEditWithLimit)
	} else if options.createdAtLimit != "" {
		whereClause = database.NewQuery(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=? AND createdAt<? AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.createdAtLimit)
	}
	statement := database.NewQuery(`
		SELECT p.pageId,p.edit,p.prevEdit,p.type,p.title,p.clickbait,p.text,p.summary,p.alias,p.creatorId,
			p.sortChildrenBy,p.hasVote,p.voteType,p.createdAt,p.karmaLock,p.privacyKey,
			p.groupId,p.parents,p.deletedBy,p.isAutosave,p.isSnapshot,p.isCurrentEdit,p.isMinorEdit,
			p.todoCount,i.currentEdit>0,i.maxEdit,i.lockedBy,i.lockedUntil
		FROM pages AS p
		JOIN (
			SELECT *
			FROM pageInfos
			WHERE pageId=?`, pageId).Add(`
		) AS i
		ON (p.pageId=i.pageId)
		WHERE`).AddPart(whereClause).Add(`AND
			(p.groupId=0 OR p.groupId IN (SELECT id FROM groups WHERE isVisible) OR p.groupId IN (SELECT groupId FROM groupMembers WHERE userId=?))`, userId).ToStatement(db)
	row := statement.QueryRow()
	exists, err := row.Scan(&p.PageId, &p.Edit, &p.PrevEdit,
		&p.Type, &p.Title, &p.Clickbait, &p.Text, &p.Summary, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.GroupId,
		&p.ParentsStr, &p.DeletedBy, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit, &p.IsMinorEdit,
		&p.TodoCount, &p.WasPublished, &p.MaxEditEver, &p.LockedBy, &p.LockedUntil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}

	p.TextLength = len(p.Text)
	if p.DeletedBy > 0 {
		return &core.Page{PageId: p.PageId, DeletedBy: p.DeletedBy}, nil
	} else if !options.ignoreParents {
		if err := p.ProcessParents(db.C, nil); err != nil {
			return nil, fmt.Errorf("Couldn't process parents: %v", err)
		}
	}
	return &p, nil
}

// loadPageIds from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func loadPageIds(rows *database.Rows, pageMap map[int64]*core.Page) ([]string, error) {
	ids := make([]string, 0, indexPanelLimit)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId int64
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}

		p, ok := pageMap[pageId]
		if !ok {
			p = &core.Page{PageId: pageId}
			pageMap[pageId] = p
		}
		ids = append(ids, fmt.Sprintf("%d", p.PageId))
		return nil
	})
	return ids, err
}

// loadChildDraft loads a potentially existing draft for the given page. If it's
// loaded, it'll be added to the give map.
func loadChildDraft(db *database.DB, userId int64, p *core.Page, pageMap map[int64]*core.Page) error {
	parentRegexp := fmt.Sprintf("(^|,)%s($|,)", strconv.FormatInt(p.PageId, core.PageIdEncodeBase))
	if p.Type != core.QuestionPageType {
		// Load potential question draft.
		row := db.NewStatement(`
			SELECT pageId
			FROM (
				SELECT pageId,creatorId
				FROM pages
				WHERE type="question" AND deletedBy<=0 AND parents REGEXP ?
				GROUP BY pageId
				HAVING SUM(isCurrentEdit)<=0
			) AS p
			WHERE creatorId=?
			LIMIT 1`).QueryRow(parentRegexp, userId)
		_, err := row.Scan(&p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load question draft: %v", err)
		}
	} else {
		// Load potential answer draft.
		row := db.NewStatement(`
			SELECT pageId
			FROM (
				SELECT pageId,creatorId
				FROM pages
				WHERE type="answer" AND deletedBy<=0 AND parents REGEXP ?
				GROUP BY pageId
				HAVING SUM(isCurrentEdit)<=0
			) AS p
			WHERE creatorId=?
			LIMIT 1`).QueryRow(parentRegexp, userId)
		_, err := row.Scan(&p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load answer draft id: %v", err)
		}
		if p.ChildDraftId > 0 {
			p, err := loadFullEdit(db, p.ChildDraftId, userId, &loadEditOptions{loadNonliveEdit: true})
			if err != nil {
				return fmt.Errorf("Couldn't load answer draft: %v", err)
			}
			pageMap[p.PageId] = p
		}
	}
	return nil
}

// loadLikes loads likes corresponding to the given pages and updates the pages.
func loadLikes(db *database.DB, currentUserId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT userId,pageId,value
		FROM (
			SELECT *
			FROM likes
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId int64
		var pageId int64
		var value int
		err := rows.Scan(&userId, &pageId, &value)
		if err != nil {
			return fmt.Errorf("failed to scan for a like: %v", err)
		}
		page := pageMap[pageId]
		if value > 0 {
			if page.LikeCount >= page.DislikeCount {
				page.LikeScore++
			} else {
				page.LikeScore += 2
			}
			page.LikeCount++
		} else if value < 0 {
			if page.DislikeCount >= page.LikeCount {
				page.LikeScore--
			}
			page.DislikeCount++
		}
		if userId == currentUserId {
			page.MyLikeValue = value
		}
		return nil
	})
	return err
}

// loadVotes loads probability votes corresponding to the given pages and updates the pages.
func loadVotes(db *database.DB, currentUserId int64, pageMap map[int64]*core.Page, usersMap map[int64]*core.User) error {
	pageIds := core.PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT userId,pageId,value,createdAt
		FROM (
			SELECT *
			FROM votes
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
			ORDER BY id DESC
		) AS v
		GROUP BY userId,pageId`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var v core.Vote
		var pageId int64
		err := rows.Scan(&v.UserId, &pageId, &v.Value, &v.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a vote: %v", err)
		}
		if v.Value == 0 {
			return nil
		}
		page := pageMap[pageId]
		if page.Votes == nil {
			page.Votes = make([]*core.Vote, 0, 0)
		}
		page.Votes = append(page.Votes, &v)
		if _, ok := usersMap[v.UserId]; !ok {
			usersMap[v.UserId] = &core.User{Id: v.UserId}
		}
		return nil
	})
	return err
}

// loadLinks loads the links for the given page.
func loadLinks(db *database.DB, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	// List of all aliases we need to get titles for
	aliasesList := make([]interface{}, 0)
	// Map of each page alias to a list of pages which have it as a link.
	linkMap := make(map[string]string)

	// Load all links.
	pageIds := core.PageIdsListFromMap(pageMap)
	rows := db.NewStatement(`
		SELECT parentId,childAlias
		FROM links
		WHERE parentId IN ` + database.InArgsPlaceholder(len(pageIds))).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId int64
		var childAlias string
		err := rows.Scan(&parentId, &childAlias)
		if err != nil {
			return fmt.Errorf("failed to scan for an alias: %v", err)
		}
		aliasesList = append(aliasesList, childAlias)
		if pageMap[parentId].Links == nil {
			pageMap[parentId].Links = make(map[string]string)
		}
		pageMap[parentId].Links[childAlias] = ""
		return nil
	})
	if err != nil {
		return err
	}

	// Get the page titles for all the links.
	if len(aliasesList) > 0 {
		// Double up aliases list because we'll use it twice in the query.
		placeholder := database.InArgsPlaceholder(len(aliasesList))
		aliasesList = append(aliasesList, aliasesList...)
		rows = db.NewStatement(`
			SELECT pageId,alias,title
			FROM pages
			WHERE isCurrentEdit AND deletedBy=0 AND
				(alias IN ` + placeholder + ` OR pageId IN ` + placeholder + ` )`).Query(aliasesList...)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId, alias, title string
			err := rows.Scan(&pageId, &alias, &title)
			lowercaseAlias := strings.ToLower(alias)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			linkMap[lowercaseAlias] = title
			if pageId != alias {
				linkMap[pageId] = title
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Set the links for all pages.
	for _, p := range pageMap {
		for alias, _ := range p.Links {
			p.Links[alias] = linkMap[alias]
		}
	}
	return nil
}

type loadChildrenIdsOptions struct {
	// If set, the children will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*core.Page
	// Load whether or not each child has children of its own.
	LoadHasChildren bool
}

// loadChildrenIds loads the page ids for all the children of the pages in the given pageMap.
func loadChildrenIds(db *database.DB, pageMap map[int64]*core.Page, options loadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsListFromMap(sourcePageMap)
	newPages := make(map[int64]*core.Page)
	rows := db.NewStatement(`
		SELECT pp.parentId,pp.childId,p.type
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE parentId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
		) AS pp JOIN (
			SELECT pageId,type
			FROM pages
			WHERE isCurrentEdit AND type!="comment"
		) AS p
		ON (p.pageId=pp.childId)`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p core.PagePair
		var childType string
		err := rows.Scan(&p.ParentId, &p.ChildId, &childType)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[p.ChildId]
		if !ok {
			newPage = &core.Page{PageId: p.ChildId, Type: childType}
			pageMap[newPage.PageId] = newPage
			newPages[newPage.PageId] = newPage
		}
		newPage.Parents = append(newPage.Parents, &p)

		parent := sourcePageMap[p.ParentId]
		if newPage.Type == core.LensPageType {
			parent.LensIds = append(parent.LensIds, fmt.Sprintf("%d", newPage.PageId))
		} else {
			parent.Children = append(parent.Children, &p)
			parent.HasChildren = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if options.LoadHasChildren && len(newPages) > 0 {
		pageIds = core.PageIdsListFromMap(newPages)
		rows := db.NewStatement(`
			SELECT pp.parentId,sum(1)
			FROM (
				SELECT parentId,childId
				FROM pagePairs
				WHERE parentId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
			) AS pp JOIN (
				SELECT pageId
				FROM pages
				WHERE isCurrentEdit AND type!="comment"
			) AS p
			ON (p.pageId=pp.childId)
			GROUP BY 1`).Query(pageIds...)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var children int
			err := rows.Scan(&pageId, &children)
			if err != nil {
				return fmt.Errorf("failed to scan for grandchildren: %v", err)
			}
			pageMap[pageId].HasChildren = children > 0
			return nil
		})
	}
	return err
}

// loadCommentIds loads the page ids for all the comments of the pages in the given pageMap.
func loadCommentIds(db *database.DB, pageMap map[int64]*core.Page, sourcePageMap map[int64]*core.Page) error {
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsListFromMap(sourcePageMap)
	rows := db.NewStatement(`
		SELECT pp.parentId,pp.childId
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE parentId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
		) AS pp JOIN (
			SELECT pageId
			FROM pages
			WHERE isCurrentEdit AND type="comment"
		) AS p
		ON (p.pageId=pp.childId)`).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p core.PagePair
		err := rows.Scan(&p.ParentId, &p.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for comments: %v", err)
		}
		newPage, ok := pageMap[p.ChildId]
		if !ok {
			newPage = &core.Page{PageId: p.ChildId, Type: core.CommentPageType}
			pageMap[newPage.PageId] = newPage
		}
		newPage.Parents = append(newPage.Parents, &p)
		sourcePageMap[p.ParentId].CommentIds = append(sourcePageMap[p.ParentId].CommentIds, fmt.Sprintf("%d", p.ChildId))
		return nil
	})
	return err
}

type loadParentsIdsOptions struct {
	// If set, the parents will be loaded for these pages, but added to the
	// map passed in as the argument.
	ForPages map[int64]*core.Page
	// Load whether or not each parent has parents of its own.
	LoadHasParents bool
}

// loadParentsIds loads the page ids for all the parents of the pages in the given pageMap.
func loadParentsIds(db *database.DB, pageMap map[int64]*core.Page, options loadParentsIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}

	pageIds := core.PageIdsListFromMap(sourcePageMap)
	newPages := make(map[int64]*core.Page)
	rows := db.NewStatement(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE childId IN ` + database.InArgsPlaceholder(len(pageIds))).Query(pageIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var p core.PagePair
		err := rows.Scan(&p.ParentId, &p.ChildId)
		if err != nil {
			return fmt.Errorf("failed to scan for page pairs: %v", err)
		}
		newPage, ok := pageMap[p.ParentId]
		if !ok {
			newPage = &core.Page{PageId: p.ParentId}
			pageMap[newPage.PageId] = newPage
			newPages[newPage.PageId] = newPage
		}
		newPage.Children = append(newPage.Children, &p)
		sourcePageMap[p.ChildId].Parents = append(sourcePageMap[p.ChildId].Parents, &p)
		sourcePageMap[p.ChildId].HasParents = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to load parents: %v", err)
	}
	if options.LoadHasParents && len(newPages) > 0 {
		pageIds = core.PageIdsListFromMap(newPages)
		rows := db.NewStatement(`
			SELECT childId,sum(1)
			FROM pagePairs
			WHERE childId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
			GROUP BY 1`).Query(pageIds...)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var parents int
			err := rows.Scan(&pageId, &parents)
			if err != nil {
				return fmt.Errorf("failed to scan for grandparents: %v", err)
			}
			pageMap[pageId].HasParents = parents > 0
			return nil
		})
		if err != nil {
			return fmt.Errorf("Failed to load grandparents: %v", err)
		}
	}
	return nil
}

// loadDraftExistence computes for each page whether or not the user has an
// autosave draft for it.
// This only makes sense to call for pages which were loaded for isCurrentEdit=true.
func loadDraftExistence(db *database.DB, userId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId,MAX(
				IF(isAutosave AND creatorId=? AND deletedBy=0, edit, -1)
			) as myMaxEdit, MAX(IF(isCurrentEdit, edit, -1)) AS currentEdit
		FROM pages`, userId).Add(`
		WHERE pageId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY pageId
		HAVING myMaxEdit > currentEdit`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId int64
		var blank int
		err := rows.Scan(&pageId, &blank, &blank)
		if err != nil {
			return fmt.Errorf("failed to scan a page id: %v", err)
		}
		pageMap[pageId].HasDraft = true
		return nil
	})
	return err
}

// loadLastVisits loads lastVisit variable for each page.
func loadLastVisits(db *database.DB, currentUserId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT pageId,max(createdAt)
		FROM visits
		WHERE userId=?`, currentUserId).Add(`AND pageId IN`).AddArgsGroup(pageIds).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId int64
		var createdAt string
		err := rows.Scan(&pageId, &createdAt)
		if err != nil {
			return fmt.Errorf("failed to scan for a comment like: %v", err)
		}
		pageMap[pageId].LastVisit = createdAt
		return nil
	})
	return err
}

// loadSubscriptions loads subscription statuses corresponding to the given
// pages, and then updates the given maps.
func loadSubscriptions(db *database.DB, currentUserId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsListFromMap(pageMap)
	rows := database.NewQuery(`
		SELECT toPageId
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toPageId IN`).AddArgsGroup(pageIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toPageId int64
		err := rows.Scan(&toPageId)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		pageMap[toPageId].IsSubscribed = true
		return nil
	})
	return err
}

// loadUserSubscriptions loads subscription statuses corresponding to the given
// users, and then updates the given map.
func loadUserSubscriptions(db *database.DB, currentUserId int64, userMap map[int64]*core.User) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIds := core.IdsListFromUserMap(userMap)
	rows := database.NewQuery(`
		SELECT toUserId
		FROM subscriptions
		WHERE userId=?`, currentUserId).Add(`AND toUserId IN`).AddArgsGroup(userIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var toUserId int64
		err := rows.Scan(&toUserId)
		if err != nil {
			return fmt.Errorf("failed to scan for a subscription: %v", err)
		}
		userMap[toUserId].IsSubscribed = true
		return nil
	})
	return err
}

type loadAuxPageDataOptions struct {
	// If set, pretend that we last visited all the pages on this date.
	// Used when we refresh the page, but don't want to erase the new/updated stars just yet.
	ForcedLastVisit string
}

// loadAuxPageData loads the auxillary page data for the given pages.
func loadAuxPageData(db *database.DB, userId int64, pageMap map[int64]*core.Page, options *loadAuxPageDataOptions) error {
	if options == nil {
		options = &loadAuxPageDataOptions{}
	}

	// Load likes
	err := loadLikes(db, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load likes: %v", err)
	}

	// Load all the subscription statuses.
	if userId > 0 {
		err = loadSubscriptions(db, userId, pageMap)
		if err != nil {
			return fmt.Errorf("Couldn't load subscriptions: %v", err)
		}
	}

	// Load whether or not pages have drafts.
	err = loadDraftExistence(db, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load draft existence: %v", err)
	}

	// Load original creation date.
	if len(pageMap) > 0 {
		pageIds := core.PageIdsListFromMap(pageMap)
		rows := db.NewStatement(`
			SELECT pageId,MIN(createdAt)
			FROM pages
			WHERE pageId IN ` + database.InArgsPlaceholder(len(pageIds)) + `
				AND NOT isAutosave AND NOT isSnapshot
			GROUP BY 1`).Query(pageIds...)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var originalCreatedAt string
			err := rows.Scan(&pageId, &originalCreatedAt)
			if err != nil {
				return fmt.Errorf("failed to scan for original createdAt: %v", err)
			}
			pageMap[pageId].OriginalCreatedAt = originalCreatedAt
			return nil
		})
		if err != nil {
			return fmt.Errorf("Couldn't load original createdAt: %v", err)
		}
	}

	// Load last visit time.
	err = loadLastVisits(db, userId, pageMap)
	if err != nil {
		return fmt.Errorf("error while fetching a visit: %v", err)
	}
	if options.ForcedLastVisit != "" {
		// Reset the last visit date for all the pages we actually visited
		for _, p := range pageMap {
			if p.LastVisit > options.ForcedLastVisit {
				p.LastVisit = options.ForcedLastVisit
			}
		}
	}

	return nil
}

// getMaxKarmaLock returns the highest possible karma lock a user with the
// given amount of karma can create.
func getMaxKarmaLock(karma int) int {
	return int(float32(karma) * core.MaxKarmaLockFraction)
}

// getPageUrl returns the domain relative url for accessing the given page.
func getPageUrl(p *core.Page) string {
	privacyAddon := ""
	if p.PrivacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey)
	}
	return fmt.Sprintf("/pages/%s%s", p.Alias, privacyAddon)
}

// getEditPageUrl returns the domain relative url for editing the given page.
func getEditPageUrl(p *core.Page) string {
	var privacyAddon string
	if p.PrivacyKey > 0 {
		privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey)
	}
	return fmt.Sprintf("/edit/%d%s", p.PageId, privacyAddon)
}

// Check if the user can edit this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "comment" = can't perform action because this is a comment page the user doesn't own
// "###" = user doesn't have at least ### karma
func getEditLevel(p *core.Page, u *user.User) string {
	if p.Type == core.CommentPageType {
		if p.CreatorId == u.Id {
			return ""
		} else {
			return p.Type
		}
	}
	karmaReq := p.KarmaLock
	if karmaReq < core.EditPageKarmaReq && p.WasPublished {
		karmaReq = core.EditPageKarmaReq
	}
	if u.Karma < karmaReq {
		if u.IsAdmin {
			return "admin"
		}
		return fmt.Sprintf("%d", karmaReq)
	}
	return ""
}

// Check if the user can delete this page. Possible return values:
// "" = user has correct permissions to perform the action
// "admin" = user can perform the action, but only because they are an admin
// "###" = user doesn't have at least ### karma
func getDeleteLevel(p *core.Page, u *user.User) string {
	if p.Type == core.CommentPageType {
		if p.CreatorId == u.Id {
			return ""
		} else if u.IsAdmin {
			return "admin"
		} else {
			return p.Type
		}
	}
	karmaReq := p.KarmaLock
	if karmaReq < core.DeletePageKarmaReq {
		karmaReq = core.DeletePageKarmaReq
	}
	if u.Karma < karmaReq {
		if u.IsAdmin {
			return "admin"
		}
		return fmt.Sprintf("%d", karmaReq)
	}
	return ""
}
