/* Each row gives additional trust score to a specific user in a specific domain. */
CREATE TABLE userDomainBonusTrust (
	/* Id of User. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the domain the page belongs to. FK into groups. */
	domainId VARCHAR(32) NOT NULL,
	/* Extra EDIT trust a user has. */
	bonusEditTrust BIGINT NOT NULL,
	PRIMARY KEY(userId, domainId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
