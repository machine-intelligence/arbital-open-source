/* An entry for every invite + email combination code. */
CREATE TABLE inviteEmailPairs (
	/* Invite's unique code. FK into invites */
	code VARCHAR(32) NOT NULL,
	/* Email address to send invite to */
	email VARCHAR(100) NOT NULL,
	/* Id of user claiming invite. FK into users */
	claimingUserId VARCHAR(32) NOT NULL,
	/* Date this invite was claimed */
	claimedAt DATETIME NOT NULL,
	PRIMARY KEY(code, email)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
