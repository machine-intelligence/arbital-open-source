/* This table contains all the subscriptions. */
CREATE TABLE subscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the thing (user, page, etc...) the user is subscribed to. FK into pages. */
	toId VARCHAR(32) NOT NULL,

	/* Whether the user is subscribed as a maintainer of the thing. */
	asMaintainer BOOLEAN NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
