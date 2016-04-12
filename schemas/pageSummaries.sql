/* This table contains all the summaries for all the pages. */
CREATE TABLE pageSummaries (

	/* Id of the page the summary is for. */
	pageId VARCHAR(32) NOT NULL,

	/* Name of the summary. */
	name VARCHAR(32) NOT NULL,

	/* Text of the summary. */
	text TEXT NOT NULL,

	PRIMARY KEY(pageId, name)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

