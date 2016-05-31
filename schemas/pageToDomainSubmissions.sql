/* This table contains pages that have been submitted to a domain. */
CREATE TABLE pageToDomainSubmissions (
	/* Id of the submitted page. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* Id of the domain it's submitted to. FK into pageInfos. */
	domainId VARCHAR(32) NOT NULL,
	/* When this submission was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the user who submitted. FK into users. */
	submitterId VARCHAR(32) NOT NULL,
	/* Id of the user who approved the submission. FK into users. */
	approverId VARCHAR(32) NOT NULL,

	PRIMARY KEY(pageId,domainId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
