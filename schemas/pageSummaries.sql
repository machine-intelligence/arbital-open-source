/* This table contains all the summaries for all the pages. */
CREATE TABLE pageSummaries (
	/* Id of the page the summary is for. */
	pageId BIGINT NOT NULL,
	/* Name of the summary. */
	name VARCHAR(32) NOT NULL,
	/* Text of the summary. */
	text TEXT NOT NULL,
	PRIMARY KEY(pageId, name)
) CHARACTER SET utf8 COLLATE utf8_general_ci;

