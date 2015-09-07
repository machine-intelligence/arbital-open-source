// group.go contains all the stuff about groups
package core

import ()

type Member struct {
	UserId        int64 `json:"userId,string"`
	CanAddMembers bool  `json:"canAddMembers"`
	CanAdmin      bool  `json:"canAdmin"`
}

type Group struct {
	Id         int64  `json:"id,string"`
	Name       string `json:"name"`
	Alias      string `json:"alias"`
	IsVisible  bool   `json:"isVisible"`
	RootPageId int64  `json:"rootPageId,string"`
	CreatedAt  string `json:"createdAt"`

	// Optionally populated.
	Members []*Member `json:"members"`
	// Member obj corresponding to the active user
	UserMember *Member `json:"userMember"`
}
