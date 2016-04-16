/* A row for every answer. An answer is a pointer to another page, and it's always
 attached to a question. */
CREATE TABLE answers (
	/* Id of this answer. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the question this answer is for. FK into pages. */
	questionId VARCHAR(32) NOT NULL,
	/* Id of the user who added this string. FK into users. */
	answerPageId VARCHAR(32) NOT NULL,
	/* Id of the user who added this answer. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Date this answer was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
