/* This file contains the recent changes to schemas, sorted from oldest to newest. */
drop table invites;
CREATE TABLE invites (
	/* Id of user sending invite. FK into users. */
	fromUserId VARCHAR(32) NOT NULL,
	/* Id of domain that this invite is for. FK into domains. */
	domainId BIGINT NOT NULL,
	/* Role the invited user will receive. */
	role VARCHAR(32) NOT NULL,
	/* Email address to send invite to. */
	toEmail VARCHAR(100) NOT NULL,
	/* Date this invite was created. */
	createdAt DATETIME NOT NULL,
	/* If a user claimed this invite, this is their id. FK into users. */
	toUserId VARCHAR(32) NOT NULL,
	/* Date this invite was claimed */
	claimedAt DATETIME NOT NULL,
	/* When the invite email was sent. */
	emailSentAt DATETIME NOT NULL,

	PRIMARY KEY(fromUserId,domainId,toEmail)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table pageInfos add column externalUrl varchar(2048) not null;


CREATE TABLE lastVisits (

	/* FK into users. */
	userId VARCHAR(64) NOT NULL,

	/* Page id. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* Date of the first visit. */
	createdAt DATETIME NOT NULL,

	/* Date of the last visit. */
	updatedAt DATETIME NOT NULL,

	UNIQUE(userId,pageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

insert into lastVisits (userId,pageId,createdAt,updatedAt) (select userId,pageId,min(createdAt),max(createdAt) from visits where userId in (select id from users) and pageId in (select pageId from pageInfos) group by 1,2);
