/* When we snapshot a user's trust state, we enter a row in this
   table for each relevant domain. These rows all have the same id */
CREATE TABLE userTrustSnapshots (
	/* Id of the userTrustStates.  Note that this is not unique per row */
	id BIGINT NOT NULL,
	/* The domain that these trust scores belong to */
	domainId varchar(32) NOT NULL,
	/* The user's trust for general actions */
	generalTrust INT NOT NULL,
	/* The user's trust for editing actions */
	editTrust INT NOT NULL,
	/* The user this trust corresponds to */
	userId varchar(32) NOT NULL,
	/* The time this snapshot was created */
	createdAt datetime NOT NULL
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;