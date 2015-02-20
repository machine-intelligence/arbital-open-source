DROP TABLE IF EXISTS subscriptions;

/* This table contains all the subscriptions for pages and comments. */
CREATE TABLE subscriptions (
	/* Unique subscription id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* User id of the subscriber. FK into users. */
  userId BIGINT NOT NULL,

	/* == At least one of these fields has to be set. == */
	/* Id of the page the user is subscribed to. FK into pages. */
  pageId BIGINT NOT NULL,
	/* Id of the comment the user is subscribed to. FK into comments. */
  commentId BIGINT NOT NULL,
	/* ================================================= */

	/* When this subscription was created. */
  createdAt DATETIME NOT NULL,
	UNIQUE(userId, pageId, commentId),
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
