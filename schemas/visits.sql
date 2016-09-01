/* Each row is a page-user pair with the date and time
when the user has last seen the page. */
CREATE TABLE visits (

	/* If the user is logged in, user's id. FK into users. */
	userId VARCHAR(64) NOT NULL,

	/* Session id. If the user is *not* logged in, the userId will be the same as this value. */
	sessionId VARCHAR(64) NOT NULL,

	/* Analytics id. Base64-encoded Sha256 hash of the session id. */
	analyticsId VARCHAR(64) NOT NULL,

	/* IP address of the user's computer. */
	ipAddress VARCHAR(64) NOT NULL,

	/* Page id. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* When this visit occured. */
	createdAt DATETIME NOT NULL
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
