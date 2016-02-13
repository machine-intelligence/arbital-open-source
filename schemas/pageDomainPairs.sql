/* Each row describes a page-domain relationship. */
CREATE TABLE pageDomainPairs (
	/* Id of the page. FK into pages.*/
	pageId VARCHAR(32) NOT NULL,
	/* Id of the domain the page belongs to. FK into groups. */
  domainId VARCHAR(32) NOT NULL,
	/* Number of edges between the domain root and this page. */
  edgesFromRoot BIGINT NOT NULL,
  PRIMARY KEY(pageId, domainId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
