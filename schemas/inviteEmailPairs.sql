/* An entry for every invite + email combination code. */
CREATE TABLE inviteEmailPairs (
	/* Invite's unique code. FK into invites */
	code VARCHAR(32) NOT NULL,
	PRIMARY KEY(code, email)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
