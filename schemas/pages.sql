/* This table contains all the edits for all the pages, including the original edit.
 Each row is one edit for a given page. */
CREATE TABLE pages (
	/* Id of the page the edit is for. */
	pageId VARCHAR(32) NOT NULL,
	/* The edit (version) number. Always >0. */
	edit INT NOT NULL,
	/* True iff this is the edit currently used to display the page. */
	isCurrentEdit BOOLEAN NOT NULL,
	/* True iff this is a minor edit. */
	isMinorEdit BOOLEAN NOT NULL,
	/* True iff this is a snapshot saved by the creatorId user. */
	isSnapshot BOOLEAN NOT NULL,
	/* True iff this is an autosave for the creatorId user. There is at most one
	 autosave per user per page. */
	isAutosave BOOLEAN NOT NULL,
	/* User id of the creator of this edit. */
	creatorId VARCHAR(32) NOT NULL,
	/* When this edit was created. */
	createdAt DATETIME NOT NULL,

	/* Title of the page. */
	title VARCHAR(512) NOT NULL,
	/* Clickbait of the page. */
	clickbait VARCHAR(512) NOT NULL,
	/* Text of the page. */
	text MEDIUMTEXT NOT NULL,
	/* Meta-text for the page. This contains meta-data like clickbait, summary,
		masteries, etc... */
	metaText MEDIUMTEXT NOT NULL,
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
