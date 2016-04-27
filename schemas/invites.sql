/* An entry for every invite code. */
CREATE TABLE invites (
	/* PK: Invite's unique code. */
	code VARCHAR(32) NOT NULL,
	/* Type of invite: personal or group */
	type VARCHAR(32) NOT NULL,
	/* Id of user sending invite. FK into users.*/
	senderId VARCHAR(32) NOT NULL,
	/* Id of domain that this invite is for. FK into pageInfos. */
	domainId VARCHAR(32) NOT NULL,
	/* Date this invite was added to the table. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(code)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
