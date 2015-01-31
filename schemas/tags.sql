DROP TABLE IF EXISTS tags;

/* This table contains all the tags we have in our system. */
CREATE TABLE tags (
	/* Unique tag id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* User id of the person who created this tag. FK into users. */
  createdBy BIGINT NOT NULL,
	/* When this tag was created. */
  createdAt DATETIME NOT NULL,
	/* Tag text. */
	text VARCHAR(32) NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
