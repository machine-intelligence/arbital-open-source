/* This table contains all the subscriptions to discussions (page or comments). */
CREATE TABLE discussionSubscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of page/comment the user is subscribed to. FK into pageInfos. */
	toPageId VARCHAR(32) NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toPageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
