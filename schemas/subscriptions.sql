/* This table contains all the subscriptions. */
CREATE TABLE subscriptions (
	/* User id of the subscriber. FK into users. */
  userId BIGINT NOT NULL,
	/* Id of the thing (user, page, etc...) the user is subscribed to. FK into pages. */
  toId BIGINT NOT NULL,
	/* When this subscription was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(userId, toId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
