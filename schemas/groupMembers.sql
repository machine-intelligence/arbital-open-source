/* An entry for every member in a group. */
CREATE TABLE groupMembers (
	/* Id of the group. FK into groups. */
	groupId VARCHAR(32) NOT NULL,
	/* Id of the user member. FK into users. */
	userId VARCHAR(32) NOT NULL,
  /* Date this user was added. */
  createdAt DATETIME NOT NULL,
	/* Whether this user can add new members. */
	canAddMembers BOOLEAN NOT NULL,
	/* Whether this user can change the group settings. */
	canAdmin BOOLEAN NOT NULL,
  PRIMARY KEY(userId,groupId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
