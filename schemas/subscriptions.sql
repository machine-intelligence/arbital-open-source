/* This table contains all the subscriptions. */
CREATE TABLE subscriptions (
	/* User id of the subscriber. FK into users. */
  userId BIGINT NOT NULL,

	/* == At least one of these fields has to be set. == */
	/* Id of the page the user is subscribed to. FK into pages. */
  toPageId BIGINT NOT NULL,
	/* Id of the user the user is subscribed to. FK into users. */
	toUserId BIGINT NOT NULL,
	/* ================================================= */

	/* When this subscription was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(userId, toPageId, toUserId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
