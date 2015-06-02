/* This table contains all the edits for all the pages, including the original edit.
 Each row is one edit for a given page. */
CREATE TABLE pages (
	/* Id of the page the edit is for. */
	pageId BIGINT NOT NULL,
	/* The edit (version) number. Always >0 unless it's an autosave before the
	 page has been manually saved by the user. */
	edit INT NOT NULL,
	/* True iff this is the edit currently used to display the page. */
	isCurrentEdit BOOLEAN NOT NULL,
	/* True iff this is a snapshot saved by the creatorId user. */
	isSnapshot BOOLEAN NOT NULL,
	/* True iff this is an autosave for the creatorId user. There is at most one
	 autosave per user per page. */
	isAutosave BOOLEAN NOT NULL,
	/* Page's type. */
	type VARCHAR(32) NOT NULL,
	/* User id of the creator of this edit. */
	creatorId BIGINT NOT NULL,
	/* When this edit was created. */
	createdAt DATETIME NOT NULL,
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
	/* Optional name of the group this page belongs to. Only group members will be
	 able to see this page. FK into groups. */
	groupName VARCHAR(64) NOT NULL,
	/* Comma separated string of parent ids in base 36. */
	parents VARCHAR(1024) NOT NULL,
	/* If not 0, this edit has been deleted by this user id. */
	deletedBy BIGINT NOT NULL,
	PRIMARY KEY(pageId, edit)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
