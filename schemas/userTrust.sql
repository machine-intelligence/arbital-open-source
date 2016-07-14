/* Store the values for user's trust and permissions for various domains. */
CREATE TABLE userTrust (
	/* The user this trust corresponds to. FK into users. */
	userId varchar(32) NOT NULL,
	/* The domain that these trust scores belong to. */
	domainId varchar(32) NOT NULL,
	/* The user's trust for general actions */
	generalTrust INT NOT NULL,
	/* The user's trust for editing actions */
	editTrust INT NOT NULL,

	UNIQUE(userId,domainId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
