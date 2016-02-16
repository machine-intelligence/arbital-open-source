/* Each row is a page-user pair with the date and time
when the user has last seen the page. */
CREATE TABLE visits (
	/* If the user is logged in, user's id. FK into users. */
  userId VARCHAR(32) NOT NULL,
	/* Page id. FK into pages. */
  pageId VARCHAR(32) NOT NULL,
	/* When this visit occured. */
  createdAt DATETIME NOT NULL
) CHARACTER SET utf8 COLLATE utf8_general_ci;
