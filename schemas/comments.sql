DROP TABLE IF EXISTS comments;

/* This table contains all the comments we have in our system. */
CREATE TABLE comments (
	/* Unique comment id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the support this fact belong to. FK into support. */
  supportId BIGINT NOT NULL,
	/* Id of the comment this is a reply to. */
	replyToId BIGINT,
	/* When this was created. */
  createdAt DATETIME NOT NULL,
	/* User id of the creator. */
  creatorId BIGINT NOT NULL,
	/* Text supplied by the user. */
  text VARCHAR(2048) NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
