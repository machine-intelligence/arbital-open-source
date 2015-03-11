DROP TABLE IF EXISTS tags;

/* A tag is a keyword attached to a claim to allow for easier filtering / categorization. */
CREATE TABLE tags (
	/* Unique tag id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Optional tag id of the parent tag. */
	parentId BIGINT NOT NULL,
	/* User id of the person who created this tag. FK into users. */
  createdBy BIGINT NOT NULL,
	/* When this tag was created. */
  createdAt DATETIME NOT NULL,
	/* Tag text. */
	text VARCHAR(32) NOT NULL,
	/* Full tag name including parents. */
	fullName VARCHAR(255) NOT NULL,
	UNIQUE(fullName),
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
