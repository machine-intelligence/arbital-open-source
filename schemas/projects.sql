/* An entry for every project we have */
CREATE TABLE projects (
	/* Project id. */
	id BIGINT NOT NULL AUTO_INCREMENT,

	/* The page which describes this project. FK into pages. */
	projectPageId VARCHAR(32) NOT NULL,

	/* The first page the reader should go to. FK into pages. */
	startPageId VARCHAR(32) NOT NULL,

	/* State of the project. "finished", "inProgress", or "requested" */
	state VARCHAR(32) NOT NULL,

	/* Id by which we track who wants to read this. FK into likes. */
	readLikeableId BIGINT NOT NULL,

	/* Id by which we track who wants to write this. FK into likes. */
	writeLikeableId BIGINT NOT NULL,

	/* Date this entry was created. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
