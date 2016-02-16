/* This file contains queries used for converting the ids from base10 to base36. */

DROP TABLE pagesandusers;
DROP TABLE base10tobase36;

/* This table contains a list of all page ids and user ids.
 Each row is one pageId or userId. */
CREATE TABLE pagesandusers (
	/* base10 Id of the page or user */
	base10id VARCHAR(32) NOT NULL,
	/* When this edit or user was created. */
	createdAt DATETIME NOT NULL
) CHARACTER SET utf8 COLLATE utf8_general_ci;

/* This table is used for converting the ids from base10 to base36.
 Each row is one pageId or userId. */
CREATE TABLE base10tobase36 (
	/* base10 Id of the page or user */
	base10id VARCHAR(32) NOT NULL,
	/* When this edit or user was created. */
	createdAt DATETIME NOT NULL,
	/* base36 Id of the page or user */
	base36id VARCHAR(32) NOT NULL
) CHARACTER SET utf8 COLLATE utf8_general_ci;

INSERT INTO pagesandusers (base10id, createdAt)
SELECT pageId, createdAt
FROM pages
WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt)
SELECT id, createdAt
FROM users
WHERE 1;

INSERT INTO base10tobase36 (base10id, createdAt)
SELECT DISTINCT pagesandusers.base10id, pagesandusers.createdAt
FROM pagesandusers
INNER JOIN
    (SELECT pagesandusers.base10id, MIN(pagesandusers.createdAt) AS minCreatedAt
    FROM pagesandusers
    GROUP BY pagesandusers.base10id) groupedpagesandusers
ON pagesandusers.base10id = groupedpagesandusers.base10id
AND pagesandusers.createdAt = groupedpagesandusers.minCreatedAt;





