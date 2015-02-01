DROP TABLE IF EXISTS comments;

/* This table contains all the comments we have in our system. */
CREATE TABLE comments (
	/* Unique comment id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the input this fact belong to. FK into inputs. */
  inputId BIGINT NOT NULL,
	/* Id of the comment this is a reply to. */
	replyToId BIGINT NOT NULL,
	/* When this was created. */
  createdAt DATETIME NOT NULL,
	/* When this was updated at. */
  updatedAt DATETIME NOT NULL,
	/* User id of the creator. */
  creatorId BIGINT NOT NULL,
	/* Full name of the user who commented. */
	creatorName VARCHAR(64) NOT NULL,
	/* Text supplied by the user. */
  text VARCHAR(2048) NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
