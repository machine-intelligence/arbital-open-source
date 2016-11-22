/* This file contains the recent changes to schemas, sorted from oldest to newest. */
CREATE TABLE domains (
	/* Domain id. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the home page for this domain. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* When this page was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the user who created the page. FK into users. */
	createdBy VARCHAR(32) NOT NULL,
	/* Alias name of the domain. */
	alias VARCHAR(64) NOT NULL,

	/* ============ Various domain settings ============ */
	/* If true, any registered user can comment. */
	canUsersComment BOOL NOT NULL,
	/* If true, any registered user can propose an edit. */
	canUsersProposeEdits BOOL NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
insert into domains (pageId,createdAt,createdBy,alias,canUsersComment,canUsersProposeEdits) (select pageId,createdAt,createdBy,alias,true,true from pageInfos where type="domain");
update pageInfos set type="wiki" where type="domain";




CREATE TABLE domainMembers (

	/* Id of the domain. FK into domains. */
	domainId BIGINT NOT NULL,

	/* Id of the user member. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Date this user was added. */
	createdAt DATETIME NOT NULL,

	/* User's role in this domain. */
	role VARCHAR(32) NOT NULL,

	PRIMARY KEY(domainId,userId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

/* Add a domain for every non-user group */
insert into domains (pageId,createdAt,createdBy,alias,canUsersComment,canUsersProposeEdits) (select pageId,createdAt,createdBy,alias,true,true from pageInfos where type="group" and not pageId in (select id from users));
/* Refactor see/edit group ids into domain ids */
alter table pageInfos add column seeDomainId BIGINT NOT NULL;
alter table pageInfos add column editDomainId BIGINT NOT NULL;
update pageInfos as pi set seeDomainId=(select id from domains as d where d.pageId=pi.seeGroupId);
update pageInfos as pi set editDomainId=(select id from domains as d where d.pageId=pi.editGroupId);
alter table pageInfos drop column seeGroupId;
alter table pageInfos drop column editGroupId;
/* Convert all non-user group pages into normal wiki pages */
update pageInfos as pi set type="wiki" where type="group" and not pageId in (select id from users);
/* Convert group members into domain members */
insert into domainMembers (domainId,userId,createdAt,role) (select (select id from domains where pageId=groupId),userId,createdAt,if(canAdmin,"founder",if(canAddMembers,"arbiter","editor")) from groupMembers where groupId!=userId);
delete from domainMembers where domainId=0;
drop table groupMembers;

/* Create a domain for every user */
insert into domains (pageId,createdAt,createdBy,alias,canUsersComment,canUsersProposeEdits) (select id,u.createdAt,id,alias,true,true from users as u join pageInfos as pi on (u.id=pi.pageId));
/* ...and add the user as a member to that domain */
insert into domainMembers (domainId,userId,createdAt,role) (select d.id,u.id,d.createdAt,"reviewer" from users as u join domains as d on u.id=d.pageId);

/* Remove pageDomainPairs and leverage editDomainIds instead */
update pageInfos as pi set editDomainId=(select d.id from domains as d where d.pageId=(select domainId from pageDomainPairs as pdp where pi.pageId=pdp.pageId limit 1)) where editDomainId=0;
drop table pageDomainPairs;

/* Fill out editDomainId based on the page creator for everything else */
update pageInfos as pi set editDomainId=(select d.id from domains as d where d.pageId=pi.createdBy) where editDomainId=0;
delete from pageInfos where editDomainId=0;
drop table userTrust;
drop table userTrustSnapshots;
