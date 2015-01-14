DROP TABLE IF EXISTS users;

/* An entry for every user that has ever done anything in our system. */
CREATE TABLE users (
  /* PK. User's unique id. */
  id BIGINT NOT NULL AUTO_INCREMENT,
  /* Date this user was added to the table. */
  createdAt DATETIME,
	/* User's email. */
  userEmail VARCHAR(255) NOT NULL,
  /* User's self-assigned first name. */
  firstName VARCHAR(64),
  /* User's self-assigned last name. */
  lastName VARCHAR(64),
  /* Date of the last website visit. */
  lastWebsiteVisit DATETIME,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
