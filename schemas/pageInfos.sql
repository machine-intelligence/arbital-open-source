/* This table contains various information about the pages. This info is not
 dependent on any specific edit number. */
CREATE TABLE pageInfos (
	/* Id of the page the info is for. */
	pageId VARCHAR(32) NOT NULL,
	/* Likeable id for this page. Partial FK into likes.
	   Note that this is not set until the first time this page is liked. */
	likeableId BIGINT NOT NULL,

	/* Edit number currently used to display the page. 0 if this page hasn't
		been published. */
	currentEdit INT NOT NULL,
	/* Maximum edit number used by this page. */
	maxEdit INT NOT NULL,
	/* When this page was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the user who created the page. FK into users. */
	createdBy VARCHAR(32) NOT NULL,
	/* Alias name of the page. */
	alias VARCHAR(64) NOT NULL,
	/* Page's type. */
	type VARCHAR(32) NOT NULL,
	/* How to sort the page's children. */
	sortChildrenBy VARCHAR(32) NOT NULL,
	/* True iff the page has a probability vote. */
	hasVote BOOLEAN NOT NULL,
	/* Type of the vote this page has. If empty string, it has never been set.
	 But once voting is set, it can only be turned on/off, the type cannot be
	 changed. */
	voteType VARCHAR(32) NOT NULL,
	/* If true, this page is used as a requisite somewhere. */
	isRequisite BOOL NOT NULL,
	/* If true, this page teaches its requisites indirectly (e.g. by providing links). */
	indirectTeacher BOOL NOT NULL,
	/* True iff this page is currently in a deleted state. */
	isDeleted BOOLEAN NOT NULL,
	/* If set, this page has been merged into the mergedInto page id. FK into pageInfos. */
	mergedInto VARCHAR(32) NOT NULL,
	/* Number of different users who looked at this page. */
	viewCount BIGINT NOT NULL,

	/* When this page has been added to the Featured section */
	featuredAt DATETIME NOT NULL,

	/* === Permission settings === */
	/* see: who can see the page */
	/* act: who can perform actions on the page (e.g. vote, comment) */
	/* edit: who can edit the page */
	/* If set, only this group can see the page. FK into pages. */
	seeGroupId VARCHAR(32) NOT NULL,
	/* If set, only this group can edit the page. FK into pages. */
	editGroupId VARCHAR(32) NOT NULL,

	/* If set, the page is locked by this user. FK into users. */
	lockedBy VARCHAR(32) NOT NULL,
	/* Time until the user has this lock. */
	lockedUntil DATETIME NOT NULL,

	/* == Following variables are set for some specific pages. == */
	/* If this page is a lens, this is its ordering index when sorting its parent's
		lenses from most simple to most technical. */
	lensIndex INT NOT NULL,
	/* If true, this comment is meant for editors only. */
	isEditorComment BOOL NOT NULL,
	/* If true, this comment thread is resolved and should be hidden. */
	isResolved BOOL NOT NULL,
	/* The value of isEditorComment the user wanted. We might have disallowed it
		because the creator lacked the right permissions. */
	isEditorCommentIntention BOOL NOT NULL,

	PRIMARY KEY(pageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
