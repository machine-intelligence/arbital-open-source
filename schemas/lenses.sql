/* This table contains all information about lens relationships. */
CREATE TABLE lenses (
	/* Id of the lens relationships. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page that has the lens. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* Id of the lens page. FK into pageInfos. */
	lensId VARCHAR(32) NOT NULL,
	/* Ordering index when sorting the page's lenses. */
	lensIndex INT NOT NULL,
	/* Lens name that shows up in the tab. */
	lensName VARCHAR(32) NOT NULL,
	/* Id of the user who created the relationship. FK into users. */
	createdBy VARCHAR(32) NOT NULL,
	/* When this lens relationship was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the last user who updated the relationship. FK into users. */
	updatedBy VARCHAR(32) NOT NULL,
	/* When this relationship was updated last. */
	updatedAt DATETIME NOT NULL,

	UNIQUE(lensId),
	/* This constraint should apply, but makes it very difficult to update the lensIndex for multiple rows */
	/*UNIQUE(pageId,lensIndex),*/
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
