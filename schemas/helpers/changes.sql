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

CREATE TABLE domainFriends (
	/* Domain id. FK into domains. */
	domainId BIGINT NOT NULL,
	/* Id of another domain this domain is friends with. FK into domains. */
	friendId BIGINT NOT NULL,
	/* When this friendship was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the user who created the friendship. FK into users. */
	createdBy VARCHAR(32) NOT NULL,

	UNIQUE(domainId,friendId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table domains add column canUsersProposeComment bool not null;
update domains set canUsersProposeComment=true;
update domains set canUsersComment=false;

update domainMembers as dm set role="arbitrator" where (select d.pageId from domains as d where d.id=dm.domainId)=dm.userId;
alter table pageInfos add column submitToDomainId bigint not null;

alter table feedPages add column score double not null;

CREATE TABLE discussionSubscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of page/comment the user is subscribed to. FK into pageInfos. */
	toPageId VARCHAR(32) NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toPageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE userSubscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the user this user is subscribed to. FK into users. */
	toUserId VARCHAR(32) NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toUserId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE TABLE maintainerSubscriptions (

	/* User id of the subscriber. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the page the user is subscribed to. FK into pageInfos. */
	toPageId VARCHAR(32) NOT NULL,

	/* When this subscription was created. */
	createdAt DATETIME NOT NULL,

  	PRIMARY KEY(userId, toPageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

insert into maintainerSubscriptions (userId,toPageId,createdAt) (select userId,toId,createdAt from subscriptions where asMaintainer);
insert into userSubscriptions (userId,toUserId,createdAt) (select userId,toId,createdAt from subscriptions where toId in (select id from users));
delete from subscriptions where toId in (select id from users);
insert into discussionSubscriptions (userId,toPageId,createdAt) (select userId,toId,createdAt from subscriptions);
drop table subscriptions;

/* When a page's alias is changed, we add a row in this table. */
CREATE TABLE aliasRedirects (

	/* The old alias. */
	oldAlias VARCHAR(64) NOT NULL,

	/* The new alias. */
	newAlias VARCHAR(64) NOT NULL,

	UNIQUE(oldAlias)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table pageInfos add column votesAnonymous bool not null;
