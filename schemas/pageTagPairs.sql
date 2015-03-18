DROP TABLE IF EXISTS pageTagPairs;

/* Each row describes a tag that belong to a page. */
CREATE TABLE pageTagPairs (
	/* Tag id. FK into tags. */
  tagId BIGINT NOT NULL,
	/* Page id. Part of the FK into pages. */
  pageId BIGINT NOT NULL,
	/* Edit number. Part of the FK into pages. */
	edit INT NOT NULL,
	/* User id of the person who created this pair. FK into users. */
  createdBy BIGINT NOT NULL,
	/* When this pair was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(tagId, pageId, edit)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
