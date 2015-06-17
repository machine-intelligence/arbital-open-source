/* An entry for every like a user cast for a page. There could be
multiple likes from one user for the same page. */
CREATE TABLE likes (
  /* PK. Like's unique id. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the user who liked. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the page this like is for. FK into pages. */
	pageId BIGINT NOT NULL,
	/* Like value [-1,1]. */
	value TINYINT NOT NULL,
  /* Date this like was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
