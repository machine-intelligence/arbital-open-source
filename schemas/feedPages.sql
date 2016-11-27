/* This table contains all the feed pages. */
CREATE TABLE feedPages (

	/* Id of the domain feed. FK into domains. */
	domainId BIGINT NOT NULL,

	/* Id of the page in the feed. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,

	/* Id of the user who submitted it to the feed. FK into users. */
	submitterId VARCHAR(32) NOT NULL,

	/* When this submission was made. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(domainId, pageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
