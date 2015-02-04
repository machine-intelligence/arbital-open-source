DROP TABLE IF EXISTS comments;

/* A comment is attached to a claim. It can be in context (only shown when
looking at a specific claim) or out of context (always shown). */
CREATE TABLE comments (
	/* Unique comment id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the claim this comment belongs to. FK into claims. */
	claimId BIGINT NOT NULL,
	/* This is set iff the claim is in context. If it's set, the
	 comment will only appear when the parent claim id is this value. */
	contextClaimId BIGINT NOT NULL,
	/* Id of the comment this is a reply to. */
	replyToId BIGINT NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	/* When this was last updated at. */
	updatedAt DATETIME NOT NULL,
	/* User id of the creator. */
	creatorId BIGINT NOT NULL,
	/* Full name of the user who commented. */
	creatorName VARCHAR(64) NOT NULL,
	/* Text supplied by the user. */
	text VARCHAR(2048) NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
