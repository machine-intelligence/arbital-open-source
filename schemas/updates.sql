DROP TABLE IF EXISTS updates;

/* An update is a notification for the user that something new has happened, e.g.
there was a new comment. Updates are created only when a user is subscribed to
something, e.g. a comment.
 Since there could be multiple replies to a comment, and we don't want to add a
new update for each reply, we stack the updates together instead. When the
user visits the corresponding page, all the related updates are marked
as seen, and a new stack begins.
 Therefore, there could be multiple entries with the same (userId,
pageId, commentId, type) tuple. But only one of them will have seen==false. */
CREATE TABLE updates (
	/* Unique update id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* The update is for this user. FK into users. */
  userId BIGINT NOT NULL,
	/* Type of update */
	type VARCHAR(32) NOT NULL,
	/* If appropriate, this is the page in the context of which this update happened. */
	contextPageId BIGINT NOT NULL,
	/* When this update was created. */
  createdAt DATETIME NOT NULL,
	/* When this was last updated. */
	updatedAt DATETIME NOT NULL,
	/* Number of such updates. */
	count INT NOT NULL,
	/* True iff the user has seen these updates. While false, we can continue to stack similar updates together. */
	seen BOOLEAN NOT NULL,

	/* Id of the page the update came from. FK into pages. */
  fromPageId BIGINT NOT NULL,
	/* Id of the comment the update came from. FK into comments. */
  fromCommentId BIGINT NOT NULL,
	/* Id of the user the update came from. FK into users. */
	fromUserId BIGINT NOT NULL,
	/* Id of the tag the update came from. FK into tags. */
	fromTagId BIGINT NOT NULL,

  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
