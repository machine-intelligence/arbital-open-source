/* An entry for every invite. */
CREATE TABLE invites (
	/* Id of user sending invite. FK into users. */
	fromUserId VARCHAR(32) NOT NULL,
	/* Id of domain that this invite is for. FK into pageInfos. */
	domainId VARCHAR(32) NOT NULL,
	/* Email address to send invite to. */
	toEmail VARCHAR(100) NOT NULL,
	/* Date this invite was created. */
	createdAt DATETIME NOT NULL,
	/* If a user claimed this invite, this is their id. FK into users. */
	toUserId VARCHAR(32) NOT NULL,
	/* Date this invite was claimed */
	claimedAt DATETIME NOT NULL,
	/* When the invite email was sent. */
	emailSentAt DATETIME NOT NULL,

	PRIMARY KEY(fromUserId,domainId,toEmail)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
