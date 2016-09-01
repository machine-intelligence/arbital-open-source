// Contains helpers for likeable things, like pages and changeLogs.
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

const (
	// Possible likeable types
	ChangeLogLikeableType      = "changeLog"
	PageLikeableType           = "page"
	RedLinkLikeableType        = "redLink"
	ContentRequestLikeableType = "contentRequest"
	ProjectReadLikeableType    = "projectRead"
	ProjectWriteLikeableType   = "projectWrite"
)

type Likeable struct {
	LikeableID   int64  `json:"likeableId,string"`
	LikeableType string `json:"likeableType"`
	MyLikeValue  int    `json:"myLikeValue"`
	LikeCount    int    `json:"likeCount"`
	DislikeCount int    `json:"dislikeCount"`
	// Computed from LikeCount and DislikeCount
	LikeScore int `json:"likeScore"`

	// List of user ids who liked this page
	IndividualLikes []string `json:"individualLikes"`
}

func NewLikeable(likeableType string) *Likeable {
	return &Likeable{
		LikeableType:    likeableType,
		IndividualLikes: make([]string, 0),
	}
}

// Get the likeableId of the given likeable. If it doesn't have one, create one for it.
// Returns the likeableId of the likeable.
func GetOrCreateLikeableID(tx *database.Tx, likeableType string, id string) (int64, error) {
	// Figure out which table to look into
	tableName, idField, err := GetTableAndIDFieldForLikeable(likeableType)
	if err != nil {
		return 0, err
	}

	// Look up the likeable
	var likeableID int64
	row := tx.DB.NewStatement(`
		SELECT likeableId
		FROM ` + tableName + `
		WHERE ` + idField + `=?`).WithTx(tx).QueryRow(id)
	_, err = row.Scan(&likeableID)
	if err != nil {
		return 0, fmt.Errorf("Couldn't look up likeableId: %v", err)
	}

	// If it already has a likeableId, return that
	if likeableID != 0 {
		return likeableID, nil
	}

	// Otherwise, insert a new likeableId
	result, err := tx.DB.NewInsertStatement("likeableIds", make(database.InsertMap)).WithTx(tx).Exec()
	if err != nil {
		return 0, fmt.Errorf("Couldn't insert new likeableId", err)
	}
	likeableID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Couldn't retrieve new likeableId", err)
	}

	// Update the likeable with the likeableId
	hashmap := make(database.InsertMap)
	hashmap[idField] = id
	hashmap["likeableId"] = likeableID
	result, err = tx.DB.NewInsertStatement(tableName, hashmap, "likeableId").WithTx(tx).Exec()
	if err != nil {
		return 0, err
	}

	return likeableID, nil
}

// Get the name of the table and id field for the given likeableType.
func GetTableAndIDFieldForLikeable(likeableType string) (string, string, error) {
	switch likeableType {
	case PageLikeableType:
		return "pageInfos", "pageId", nil
	case ChangeLogLikeableType:
		return "changeLogs", "id", nil
	case RedLinkLikeableType:
		return "redLinks", "alias", nil
	case ContentRequestLikeableType:
		return "contentRequests", "id", nil
	case ProjectReadLikeableType:
		return "projects", "readLikeableId", nil
	case ProjectWriteLikeableType:
		return "projects", "writeLikeableId", nil
	default:
		return "", "", fmt.Errorf("invalid likeableType")
	}
}

// Check if the given likeableType is valid.
func IsValidLikeableType(likeableType string) bool {
	switch likeableType {
	case PageLikeableType, ChangeLogLikeableType, RedLinkLikeableType, ContentRequestLikeableType:
		return true
	default:
		return false
	}
}

// LoadLikes loads likes corresponding to the given likeable objects.
// Also loads individual likes for the likeables in the individualLikeablesMap.
func LoadLikes(db *database.DB, u *CurrentUser, likeablesMap map[int64]*Likeable, individualLikeablesMap map[int64]*Likeable, userMap map[string]*User) error {
	if len(likeablesMap) <= 0 {
		return nil
	}

	likeableIDs := make([]interface{}, 0)
	for id := range likeablesMap {
		likeableIDs = append(likeableIDs, id)
	}
	for id := range individualLikeablesMap {
		likeableIDs = append(likeableIDs, id)
	}

	rows := database.NewQuery(`
		SELECT likeableId,userId,value
		FROM likes
		WHERE likeableId IN`).AddArgsGroup(likeableIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likeableID int64
		var userID string
		var value int
		err := rows.Scan(&likeableID, &userID, &value)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		if likeable, ok := likeablesMap[likeableID]; ok {
			// We count the current user's like value towards the sum in the FE rather than here.
			if userID == u.ID {
				likeable.MyLikeValue = value
			} else if value > 0 {
				likeable.LikeCount++
			} else if value < 0 {
				likeable.DislikeCount++
			}
		}

		// Store the individual likes for pages that want them
		if value > 0 && individualLikeablesMap != nil {
			if likeable, ok := individualLikeablesMap[likeableID]; ok {
				likeable.IndividualLikes = append(likeable.IndividualLikes, userID)
				AddUserIDToMap(userID, userMap)
			}
		}
		return nil
	})

	// Calculate the like score
	for _, likeable := range likeablesMap {
		// if likes >= dislikes : likes
		// if likes < dislikes : likes - (dislikes - likes) = 2*likes - dislikes
		//
		// or in other words:
		// start with the like count, and for every dislike more than the number of likes, subtract 1
		//
		// examples:
		// 10 likes, 0 dislikes : 10
		// 10 likes, 9 dislikes : 10
		// 9 likes, 10 dislikes : 8
		// 0 likes, 10 dislikes : -10
		if likeable.LikeCount >= likeable.DislikeCount {
			likeable.LikeScore = likeable.LikeCount
		} else {
			likeable.LikeScore = 2*likeable.LikeCount - likeable.DislikeCount
		}
	}

	return err
}

// LoadLikesForPages loads likes corresponding to the given pages and updates the pages.
func LoadLikesForPages(db *database.DB, u *CurrentUser, pageMap map[string]*Page, individualLikesPageMap map[string]*Page, userMap map[string]*User) error {
	// Likeables that need like counts
	likeablesMap := make(map[int64]*Likeable)
	for _, page := range pageMap {
		if page.LikeableID != 0 {
			likeablesMap[page.LikeableID] = &page.Likeable
		}
	}
	// Likeables that need lists of individual likes
	individualLikeablesMap := make(map[int64]*Likeable)
	for _, page := range individualLikesPageMap {
		if page.LikeableID != 0 {
			individualLikeablesMap[page.LikeableID] = &page.Likeable
		}
	}
	return LoadLikes(db, u, likeablesMap, individualLikeablesMap, userMap)
}
