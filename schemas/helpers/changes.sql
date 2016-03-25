/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table users add column ignoreMathjax bool not null;
alter table pageInfos add column lensIndex int not null;
alter table userMasteryPairs drop column isManuallySet;
alter table userMasteryPairs add column wants boolean not null;
alter table pageInfos add column isRequisite BOOL NOT NULL;
alter table pageInfos add column indirectTeacher BOOL NOT NULL;
alter table users add column fbUserId VARCHAR(32) NOT NULL;

ALTER DATABASE zanaduu CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;

ALTER TABLE base10tobase36 CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE likes CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE links CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE pageSummaries CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE pages CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE pagesandusers CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE updates CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE email email varchar(191) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;# email shortened from 255 characters for maximum index size      
ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE visits CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE votes CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

ALTER TABLE base10tobase36 CHANGE base10id base10id varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE base10tobase36 CHANGE base36id base36id varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE type type varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE auxPageId auxPageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE auxPageIdBase10 auxPageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE changeLogs CHANGE auxPageIdBase36 auxPageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CHANGE groupId groupId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CHANGE groupIdBase10 groupIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE groupMembers CHANGE groupIdBase36 groupIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE likes CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE likes CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE likes CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE likes CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE likes CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE likes CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE links CHANGE parentId parentId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE links CHANGE parentIdBase10 parentIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE links CHANGE parentIdBase36 parentIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE links CHANGE childAlias childAlias varchar(64) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE links CHANGE childAliasBase10 childAliasBase10 varchar(64) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE links CHANGE childAliasBase36 childAliasBase36 varchar(64) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CHANGE domainId domainId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CHANGE domainIdBase10 domainIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageDomainPairs CHANGE domainIdBase36 domainIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE lockedBy lockedBy varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE lockedByBase10 lockedByBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE lockedByBase36 lockedByBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE alias alias varchar(64) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE aliasBase10 aliasBase10 varchar(64) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE aliasBase36 aliasBase36 varchar(64) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE type type varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE sortChildrenBy sortChildrenBy varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE voteType voteType varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE seeGroupId seeGroupId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE seeGroupIdBase10 seeGroupIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE seeGroupIdBase36 seeGroupIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE editGroupId editGroupId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE editGroupIdBase10 editGroupIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE editGroupIdBase36 editGroupIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE createdBy createdBy varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE createdByBase10 createdByBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageInfos CHANGE createdByBase36 createdByBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE parentId parentId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE parentIdBase10 parentIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE parentIdBase36 parentIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE childId childId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE childIdBase10 childIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE childIdBase36 childIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagePairs CHANGE type type varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageSummaries CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageSummaries CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageSummaries CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageSummaries CHANGE name name varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pageSummaries CHANGE text text text CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE creatorId creatorId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE creatorIdBase10 creatorIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE creatorIdBase36 creatorIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE title title varchar(512) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE text text mediumtext CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE text2 text2 mediumtext CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE anchorContext anchorContext mediumtext CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE anchorText anchorText mediumtext CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE clickbait clickbait varchar(512) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pages CHANGE metaText metaText mediumtext CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE pagesandusers CHANGE base10id base10id varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CHANGE toId toId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CHANGE toIdBase10 toIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE subscriptions CHANGE toIdBase36 toIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE type type varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE groupByPageId groupByPageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE groupByPageIdBase10 groupByPageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE groupByPageIdBase36 groupByPageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE groupByUserId groupByUserId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE groupByUserIdBase10 groupByUserIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE groupByUserIdBase36 groupByUserIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE subscribedToId subscribedToId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE subscribedToIdBase10 subscribedToIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE subscribedToIdBase36 subscribedToIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE goToPageId goToPageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE goToPageIdBase10 goToPageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE goToPageIdBase36 goToPageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE byUserId byUserId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE byUserIdBase10 byUserIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE updates CHANGE byUserIdBase36 byUserIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CHANGE masteryId masteryId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CHANGE masteryIdBase10 masteryIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE userMasteryPairs CHANGE masteryIdBase36 masteryIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE id id varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE idBase10 idBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE idBase36 idBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;                                                   
ALTER TABLE users CHANGE firstName firstName varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE lastName lastName varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE inviteCode inviteCode varchar(16) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE emailFrequency emailFrequency varchar(16) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE users CHANGE fbUserId fbUserId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE visits CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE visits CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE visits CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE visits CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE visits CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE visits CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE votes CHANGE userId userId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE votes CHANGE userIdBase10 userIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE votes CHANGE userIdBase36 userIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE votes CHANGE pageId pageId varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE votes CHANGE pageIdBase10 pageIdBase10 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;
ALTER TABLE votes CHANGE pageIdBase36 pageIdBase36 varchar(32) CHARACTER SET utf8mb4 NOT NULL COLLATE utf8mb4_general_ci;

REPAIR TABLE base10tobase36;
OPTIMIZE TABLE base10tobase36;
REPAIR TABLE changeLogs;
OPTIMIZE TABLE changeLogs;
REPAIR TABLE groupMembers;
OPTIMIZE TABLE groupMembers;
REPAIR TABLE likes;
OPTIMIZE TABLE likes;
REPAIR TABLE links;
OPTIMIZE TABLE links;
REPAIR TABLE pageDomainPairs;
OPTIMIZE TABLE pageDomainPairs;
REPAIR TABLE pageInfos;
OPTIMIZE TABLE pageInfos;
REPAIR TABLE pagePairs;
OPTIMIZE TABLE pagePairs;
REPAIR TABLE pageSummaries;
OPTIMIZE TABLE pageSummaries;
REPAIR TABLE pages;
OPTIMIZE TABLE pages;
REPAIR TABLE pagesandusers;
OPTIMIZE TABLE pagesandusers;
REPAIR TABLE subscriptions;
OPTIMIZE TABLE subscriptions;
REPAIR TABLE updates;
OPTIMIZE TABLE updates;
REPAIR TABLE userMasteryPairs;
OPTIMIZE TABLE userMasteryPairs;
REPAIR TABLE users;
OPTIMIZE TABLE users;
REPAIR TABLE visits;
OPTIMIZE TABLE visits;
REPAIR TABLE votes;
OPTIMIZE TABLE votes;
CREATE TABLE userPageObjectPairs (
	/* Id of the user the user is for. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the page the info is for. */
	pageId VARCHAR(32) NOT NULL,
	/* Page's published edit at the time this value was set. */
	edit INT NOT NULL,
	/* Alias name of the object. */
	object VARCHAR(64) NOT NULL,
	/* When this value was originally created at. */
	createdAt DATETIME NOT NULL,
	/* When this value was updated. */
	updatedAt DATETIME NOT NULL,

	/* Whatever value the object decides to set here. */
	value VARCHAR(512) NOT NULL,

	PRIMARY KEY(userId,pageId,object)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
alter table pageInfos add column isEditorComment bool not null;
alter table pages add column prevEdit int not null;
create table copyPages like pages;
insert copyPages select * from pages;
update pages as p set prevEdit=(select max(cp.edit) from copyPages as cp where p.pageId=cp.pageId and NOT cp.isAutosave and NOT cp.isSnapshot and cp.createdAt<p.createdAt);
drop table copyPages;
alter table pages add column snapshotText mediumtext not null;
