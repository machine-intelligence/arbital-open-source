/* An update is a notification for the user that something new has happened, e.g.
there was a new comment. Updates are created only when a user is subscribed to
something, usually a page.

When the user visits the update pages, all the counts are zeroed out, since
the user has been made aware of all the updates.
*/
CREATE TABLE updates (
	/* Unique update id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* The update is for this user. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* The update was generated by this user. FK into users. */
	byUserId VARCHAR(32) NOT NULL,
	/* Type of update */
	type VARCHAR(32) NOT NULL,
	/* When this update was created. */
	createdAt DATETIME NOT NULL,
	/* Whether the user has seen this update. */
	unseen BOOLEAN NOT NULL,
	/* Whether the user has dismissed this update. */
	dismissed BOOLEAN NOT NULL,
	/* True if this update has been emailed out. */
	emailed BOOLEAN NOT NULL,
	/* One of these has to be set. Updates will be grouped by this key and show up
		in the same panel. */
	groupByPageId VARCHAR(32) NOT NULL,
	groupByUserId VARCHAR(32) NOT NULL,
	/* User got this update because they are subscribed to "this thing". FK into
		pages. */
	subscribedToId VARCHAR(32) NOT NULL,
	/* User will be directed to "this thing" for more information about the update. */
	goToPageId VARCHAR(32) NOT NULL,

	/* ==== Optional vars ==== */
	/* Only set if type is 'pageInfoEdit'. Used to show what changed on the updates page.
		FK into changeLogs. */
	changeLogId VARCHAR(32) NOT NULL,
	/* Only set if the update is about a mark. FK into marks. */
	markId BIGINT NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
