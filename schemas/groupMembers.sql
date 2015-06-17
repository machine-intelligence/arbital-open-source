/* An entry for every member in a group. */
CREATE TABLE groupMembers (
	/* Id of the user member. FK into users. */
	userId BIGINT NOT NULL,
	/* Name of the group. FK into groups. */
	groupName VARCHAR(64) NOT NULL,
  /* Date this user was added. */
  createdAt DATETIME NOT NULL,
	/* Whether this user can add new members. */
	canAddMembers BOOLEAN NOT NULL,
	/* Whether this user can change the group settings. */
	canAdmin BOOLEAN NOT NULL,
  PRIMARY KEY(userId,groupName)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
