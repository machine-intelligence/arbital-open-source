/* Each row is a page-user pair with the date and time
when the user has last seen the page. */
CREATE TABLE visits (
	/* User id. FK into users. */
  userId BIGINT NOT NULL,
	/* Page id. FK into pages. */
  pageId BIGINT NOT NULL,
	/* When this visit occured. */
  createdAt DATETIME NOT NULL
) CHARACTER SET utf8 COLLATE utf8_general_ci;
