// pageHelpers.go contains the page struct as well as helpful functions.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

type loadEditOptions struct {
	// If true, the last edit will be loaded for the given user, even if it's an
	// autosave or a snapshot.
	loadNonliveEdit bool
	// Don't convert loaded parents string into an array of parents
	ignoreParents bool
	// If set, we'll only load from edits less than this
	loadEditWithLimit int
	// If set, we'll only load from edits with createdAt timestamp before this
	createdAtLimit string
}

// loadFullEdit loads and retuns the last edit for the given page id and user id,
// even if it's not live. It also loads all the auxillary data like tags.
// If the page couldn't be found, (nil, nil) will be returned.
func loadFullEdit(c sessions.Context, pageId, userId int64) (*core.Page, error) {
	return loadFullEditWithOptions(c, pageId, userId, nil)
}

// loadFullEditWithOptions is just like loadFullEdit, but takes the option
// parameters. Pass nil to use default options.
func loadFullEditWithOptions(c sessions.Context, pageId, userId int64, options *loadEditOptions) (*core.Page, error) {
	if options == nil {
		options = &loadEditOptions{loadNonliveEdit: true}
	}
	pagePtr, err := loadEdit(c, pageId, userId, *options)
	if err != nil {
		return nil, err
	}
	if pagePtr == nil {
		return nil, nil
	}
	return pagePtr, nil
}

// loadEdit loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If userId is given, the last edit of the given pageId will be returned. It
// might be an autosave or a snapshot, and thus not the current live page.
// If the page couldn't be found, (nil, nil) will be returned.
func loadEdit(c sessions.Context, pageId, userId int64, options loadEditOptions) (*core.Page, error) {
	var p core.Page
	whereClause := "p.isCurrentEdit"
	if options.loadNonliveEdit {
		whereClause = fmt.Sprintf(`
			p.edit=(
				SELECT MAX(edit)
				FROM pages
				WHERE pageId=%d AND deletedBy<=0 AND (creatorId=%d OR NOT (isSnapshot OR isAutosave))
			)`, pageId, userId)
	} else if options.loadEditWithLimit > 0 {
		whereClause = fmt.Sprintf(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=%d AND edit<%d AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.loadEditWithLimit)
	} else if options.createdAtLimit != "" {
		whereClause = fmt.Sprintf(`
			p.edit=(
				SELECT max(edit)
				FROM pages
				WHERE pageId=%d AND createdAt<'%s' AND NOT isSnapshot AND NOT isAutosave
			)`, pageId, options.createdAtLimit)
	}
	// TODO: we often don't need maxEditEver
	query := fmt.Sprintf(`
		SELECT p.pageId,p.edit,p.type,p.title,p.text,p.summary,p.alias,p.creatorId,
			p.sortChildrenBy,p.hasVote,p.voteType,p.createdAt,p.karmaLock,p.privacyKey,
			p.groupId,p.parents,p.deletedBy,p.isAutosave,p.isSnapshot,p.isCurrentEdit,
			(SELECT max(isCurrentEdit) FROM pages WHERE pageId=%[1]d) AS wasPublished,
			(SELECT max(edit) FROM pages WHERE pageId=%[1]d) AS maxEditEver,
			(SELECT ifnull(max(voteType),"") FROM pages WHERE pageId=%[1]d AND NOT isAutosave AND NOT isSnapshot AND voteType!="") AS lockedVoteType
		FROM pages AS p
		WHERE p.pageId=%[1]d AND %[2]s AND
			(p.groupId=0 OR p.groupId IN (SELECT groupId FROM groupMembers WHERE userId=%[3]d))`,
		pageId, whereClause, userId)
	exists, err := database.QueryRowSql(c, query, &p.PageId, &p.Edit,
		&p.Type, &p.Title, &p.Text, &p.Summary, &p.Alias, &p.CreatorId, &p.SortChildrenBy,
		&p.HasVote, &p.VoteType, &p.CreatedAt, &p.KarmaLock, &p.PrivacyKey, &p.GroupId,
		&p.ParentsStr, &p.DeletedBy, &p.IsAutosave, &p.IsSnapshot, &p.IsCurrentEdit,
		&p.WasPublished, &p.MaxEditEver, &p.LockedVoteType)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a page: %v", err)
	} else if !exists {
		return nil, nil
	}

	p.TextLength = len(p.Text)
	if p.DeletedBy > 0 {
		return &core.Page{PageId: p.PageId, DeletedBy: p.DeletedBy}, nil
	} else if !options.ignoreParents {
		if err := p.ProcessParents(c, nil); err != nil {
			return nil, fmt.Errorf("Couldn't process parents: %v", err)
		}
	}
	return &p, nil
}

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If the page couldn't be found, (nil, nil) will be returned.
func loadPage(c sessions.Context, pageId int64, userId int64) (*core.Page, error) {
	return loadEdit(c, pageId, userId, loadEditOptions{})
}

// loadPage loads and returns a page with the given id from the database.
// If the page is deleted, minimum amount of data will be returned.
// If the page couldn't be found, (nil, nil) will be returned.
func loadPageByAlias(c sessions.Context, pageAlias string, userId int64) (*core.Page, error) {
	pageId, err := strconv.ParseInt(pageAlias, 10, 64)
	if err != nil {
		query := fmt.Sprintf(`
			SELECT pageId
			FROM aliases
			WHERE fullName="%s"`, pageAlias)
		exists, err := database.QueryRowSql(c, query, &pageId)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load an alias: %v", err)
		} else if !exists {
			return nil, nil
		}
	}
	return loadPage(c, pageId, userId)
}

// loadPageIds from the given query and return an array containing them, while
// also updating the pageMap as necessary.
func loadPageIds(c sessions.Context, query string, pageMap map[int64]*core.Page) ([]string, error) {
	ids := make([]string, 0, indexPanelLimit)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadChildDraft(c sessions.Context, userId int64, p *core.Page, pageMap map[int64]*core.Page) error {
	if p.Type != core.QuestionPageType {
		// Load potential question draft.
		query := fmt.Sprintf(`
			SELECT pageId
			FROM pages
			WHERE type="question" AND creatorId=%d AND deletedBy<=0 AND parents REGEXP "(^|,)%s($|,)"
			GROUP BY pageId
			HAVING SUM(isCurrentEdit)<=0`, userId, strconv.FormatInt(p.PageId, core.PageIdEncodeBase))
		_, err := database.QueryRowSql(c, query, &p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load question draft: %v", err)
		}
	} else {
		// Load potential answer draft.
		query := fmt.Sprintf(`
			SELECT pageId
			FROM pages
			WHERE type="answer" AND creatorId=%d AND deletedBy<=0 AND parents REGEXP "(^|,)%s($|,)"
			GROUP BY pageId
			HAVING SUM(isCurrentEdit)<=0`, userId, strconv.FormatInt(p.PageId, core.PageIdEncodeBase))
		_, err := database.QueryRowSql(c, query, &p.ChildDraftId)
		if err != nil {
			return fmt.Errorf("Couldn't load answer draft id: %v", err)
		}
		if p.ChildDraftId > 0 {
			p, err := loadFullEdit(c, p.ChildDraftId, userId)
			if err != nil {
				return fmt.Errorf("Couldn't load answer draft: %v", err)
			}
			pageMap[p.PageId] = p
		}
	}
	return nil
}

// loadLinks loads the links for the given page.
func loadLinks(c sessions.Context, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	// List of all aliases we need to get titles for
	aliasesList := make([]string, 0, 0)
	// Map of each page alias to a list of pages which have it as a link.
	linkMap := make(map[string]string)

	// Load all links.
	pageIdsStr := core.PageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT parentId,childAlias
		FROM links
		WHERE parentId IN (%s)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var parentId int64
		var childAlias string
		err := rows.Scan(&parentId, &childAlias)
		if err != nil {
			return fmt.Errorf("failed to scan for an alias: %v", err)
		}
		aliasesList = append(aliasesList, fmt.Sprintf(`"%s"`, childAlias))
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
	aliasesStr := strings.Join(aliasesList, ",")
	if len(aliasesStr) > 0 {
		query = fmt.Sprintf(`
			SELECT alias,title
			FROM pages
			WHERE isCurrentEdit AND deletedBy=0 AND alias IN (%s)`, aliasesStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var alias, title string
			err := rows.Scan(&alias, &title)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			linkMap[alias] = title
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
func loadChildrenIds(c sessions.Context, pageMap map[int64]*core.Page, options loadChildrenIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIdsStr := core.PageIdsStringFromMap(sourcePageMap)
	newPages := make(map[int64]*core.Page)
	query := fmt.Sprintf(`
		SELECT pp.parentId,pp.childId,p.type
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE parentId IN (%s)
		) AS pp JOIN (
			SELECT pageId,type
			FROM pages
			WHERE isCurrentEdit AND type!="comment"
		) AS p
		ON (p.pageId=pp.childId)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
		pageIdsStr = core.PageIdsStringFromMap(newPages)
		query := fmt.Sprintf(`
			SELECT pp.parentId,sum(1)
			FROM (
				SELECT parentId,childId
				FROM pagePairs
				WHERE parentId IN (%s)
			) AS pp JOIN (
				SELECT pageId
				FROM pages
				WHERE isCurrentEdit AND type!="comment"
			) AS p
			ON (p.pageId=pp.childId)
			GROUP BY 1`, pageIdsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadCommentIds(c sessions.Context, pageMap map[int64]*core.Page, sourcePageMap map[int64]*core.Page) error {
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIdsStr := core.PageIdsStringFromMap(sourcePageMap)
	query := fmt.Sprintf(`
		SELECT pp.parentId,pp.childId
		FROM (
			SELECT parentId,childId
			FROM pagePairs
			WHERE parentId IN (%s)
		) AS pp JOIN (
			SELECT pageId
			FROM pages
			WHERE isCurrentEdit AND type="comment"
		) AS p
		ON (p.pageId=pp.childId)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadParentsIds(c sessions.Context, pageMap map[int64]*core.Page, options loadParentsIdsOptions) error {
	sourcePageMap := pageMap
	if options.ForPages != nil {
		sourcePageMap = options.ForPages
	}
	if len(sourcePageMap) <= 0 {
		return nil
	}
	pageIdsStr := core.PageIdsStringFromMap(sourcePageMap)
	newPages := make(map[int64]*core.Page)
	query := fmt.Sprintf(`
		SELECT parentId,childId
		FROM pagePairs
		WHERE childId IN (%s)`, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
		return err
	}
	if options.LoadHasParents && len(newPages) > 0 {
		pageIdsStr = core.PageIdsStringFromMap(newPages)
		query := fmt.Sprintf(`
			SELECT childId,sum(1)
			FROM pagePairs
			WHERE childId IN (%s)
			GROUP BY 1`, pageIdsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId int64
			var parents int
			err := rows.Scan(&pageId, &parents)
			if err != nil {
				return fmt.Errorf("failed to scan for grandparents: %v", err)
			}
			pageMap[pageId].HasParents = parents > 0
			return nil
		})
	}
	return err
}

// loadDraftExistence computes for each page whether or not the user has a
// work-in-progress draft for it.
// This only makes sense to call for pages which were loaded for isCurrentEdit=true.
func loadDraftExistence(c sessions.Context, userId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT pageId,MAX(
				IF((isSnapshot OR isAutosave) AND creatorId=%d AND deletedBy=0 AND
					(groupId=0 OR groupId IN (SELECT groupId FROM groupMembers WHERE userId=%d)),
				edit, -1)
			) as myMaxEdit, MAX(IF(isCurrentEdit, edit, -1)) AS currentEdit
		FROM pages
		WHERE pageId IN (%s)
		GROUP BY pageId
		HAVING myMaxEdit > currentEdit`,
		userId, userId, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadLastVisits(c sessions.Context, currentUserId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIdsStr := core.PageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT pageId,max(createdAt)
		FROM visits
		WHERE userId=%d AND pageId IN (%s)
		GROUP BY 1`,
		currentUserId, pageIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadSubscriptions(c sessions.Context, currentUserId int64, pageMap map[int64]*core.Page) error {
	if len(pageMap) <= 0 {
		return nil
	}
	pageIds := core.PageIdsStringFromMap(pageMap)
	query := fmt.Sprintf(`
		SELECT toPageId
		FROM subscriptions
		WHERE userId=%d AND toPageId IN (%s)`,
		currentUserId, pageIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadUserSubscriptions(c sessions.Context, currentUserId int64, userMap map[int64]*core.User) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIds := pageIdsStringFromUserMap(userMap)
	query := fmt.Sprintf(`
		SELECT toUserId
		FROM subscriptions
		WHERE userId=%d AND toUserId IN (%s)`,
		currentUserId, userIds)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
func loadAuxPageData(c sessions.Context, userId int64, pageMap map[int64]*core.Page, options *loadAuxPageDataOptions) error {
	if options == nil {
		options = &loadAuxPageDataOptions{}
	}

	// Load likes
	err := loadLikes(c, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load likes: %v", err)
	}

	// Load all the subscription statuses.
	if userId > 0 {
		err = loadSubscriptions(c, userId, pageMap)
		if err != nil {
			return fmt.Errorf("Couldn't load subscriptions: %v", err)
		}
	}

	// Load whether or not pages have drafts.
	err = loadDraftExistence(c, userId, pageMap)
	if err != nil {
		return fmt.Errorf("Couldn't load draft existence: %v", err)
	}

	// Load original creation date.
	if len(pageMap) > 0 {
		pageIdsStr := core.PageIdsStringFromMap(pageMap)
		query := fmt.Sprintf(`
			SELECT pageId,MIN(createdAt)
			FROM pages
			WHERE pageId IN (%s) AND NOT isAutosave AND NOT isSnapshot
			GROUP BY 1`, pageIdsStr)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
	err = loadLastVisits(c, userId, pageMap)
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

// pageIdsStringFromUserMap returns a comma separated string of all userIds in the given map.
func pageIdsStringFromUserMap(userMap map[int64]*core.User) string {
	var buffer bytes.Buffer
	for id, _ := range userMap {
		buffer.WriteString(fmt.Sprintf("%d,", id))
	}
	str := buffer.String()
	if len(str) >= 1 {
		str = str[0 : len(str)-1]
	}
	return str
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
// "blog" = can't perform action because this is a blog page the user doesn't own
// "comment" = can't perform action because this is a comment page the user doesn't own
// "###" = user doesn't have at least ### karma
func getEditLevel(p *core.Page, u *user.User) string {
	if p.Type == core.BlogPageType || p.Type == core.CommentPageType {
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
// "blog" = can't perform action because this is a blog page the user doesn't own
// "###" = user doesn't have at least ### karma
func getDeleteLevel(p *core.Page, u *user.User) string {
	if p.Type == core.BlogPageType || p.Type == core.CommentPageType {
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
