/* This table contains all the maintainance subscriptions. */
CREATE TABLE maintainerSubscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the page the user is subscribed to. FK into pageInfos. */
	toPageId VARCHAR(32) NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toPageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
