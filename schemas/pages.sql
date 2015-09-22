/* This table contains all the edits for all the pages, including the original edit.
 Each row is one edit for a given page. */
CREATE TABLE pages (
	/* Id of the page the edit is for. */
	pageId BIGINT NOT NULL,
	/* The edit (version) number. Always >=0. */
	edit INT NOT NULL,
	/* The edit that came before this. Set to 0 if there was none. */
	prevEdit INT NOT NULL,
	/* True iff this is the edit currently used to display the page. */
	isCurrentEdit BOOLEAN NOT NULL,
	/* True iff this is a snapshot saved by the creatorId user. */
	isSnapshot BOOLEAN NOT NULL,
	/* True iff this is an autosave for the creatorId user. There is at most one
	 autosave per user per page. */
	isAutosave BOOLEAN NOT NULL,
	/* User id of the creator of this edit. */
	creatorId BIGINT NOT NULL,
	/* When this edit was created. */
	createdAt DATETIME NOT NULL,

	/* Page's type. */
	type VARCHAR(32) NOT NULL,
	/* Title of the page. */
	title VARCHAR(512) NOT NULL,
	/* Text of the page. */
	text MEDIUMTEXT NOT NULL,
	/* Summary of the page. */
	summary TEXT NOT NULL,
	/* Alias name of the page. */
	alias VARCHAR(64) NOT NULL,

	/* How to sort the page's children. */
	sortChildrenBy VARCHAR(32) NOT NULL,
	/* True iff the page has a probability vote. */
	hasVote BOOLEAN NOT NULL,
	/* Type of the vote this page has. If empty string, it has never been set.
	 But once voting is set, it can only be turned on/off, the type cannot be
	 changed. */
	voteType VARCHAR(32) NOT NULL,
	/* Minimum amount of karma a user needs to edit this page. */
	karmaLock INT NOT NULL,
	/* If > 0, the page is accessible only with the right link. */
	privacyKey BIGINT NOT NULL,
	/* Optional id of the group this page belongs to. FK into groups. */
	groupId BIGINT NOT NULL,
	/* Comma separated string of parent ids in base 36. We store them so that we
	 can easily revert parents when reverting to an edit. */
	parents VARCHAR(1024) NOT NULL,
	/* If not 0, this edit has been deleted by this user id. */
	deletedBy BIGINT NOT NULL,
	/* Number of TODOs in this page. */
	todoCount INT NOT NULL,

	/* == Following variables are set for inline comments and questions. == */
	/* Text of the paragraph the anchor is in. */
	anchorContext MEDIUMTEXT NOT NULL,
	/* Text the comment is attached to. */
	anchorText MEDIUMTEXT NOT NULL,
	/* Offset of the text into the context. */
	anchorOffset INT NOT NULL,

	PRIMARY KEY(pageId, edit)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
