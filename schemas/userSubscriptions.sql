/* This table contains all the subscriptions to users. */
CREATE TABLE userSubscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the user this user is subscribed to. FK into users. */
	toUserId VARCHAR(32) NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toUserId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
