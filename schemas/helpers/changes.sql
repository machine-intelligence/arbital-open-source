/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table marks add resolvedPageId VARCHAR(32) NOT NULL;
alter table marks add column resolvedBy varchar(32) not null;
alter table marks drop column questionId;
alter table updates add column markId BIGINT NOT NULL;

alter table changeLogs modify column oldSettingsValue varchar(1024) not null;
alter table changeLogs modify column newSettingsValue varchar(1024) not null;

alter table marks add column answered BOOLEAN NOT NULL;
alter table pagePairs add column everPublished boolean not null;
update pagePairs set everPublished = 1
where
	parentId not in (select pageId from pageInfos where currentEdit <= 0) and
	childId not in (select pageId from pageInfos where currentEdit <= 0);

alter table pageInfos add column mergedInto varchar(32) not null;
alter table marks add column type varchar(32) not null;
update marks set type="query";
alter table marks add column resolvedAt datetime not null;
alter table marks add column answeredAt datetime not null;

alter table pageInfos add column likeableId bigint not null;

alter table changeLogs add column likeableId bigint not null;

/* A table for keeping track of likeableIds */
CREATE TABLE likeableIds (
	/* Id of the likeable. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table likes add column likeableId bigint not null;
-- START TRANSACTION;
-- SET @likeableId=0;
-- update pageInfos set likeableId=@likeableId:=@likeableId+1;
-- insert into likeableIds (id) select likeableId from pageInfos;
-- update likes join pageInfos on pageInfos.pageId=likes.pageId set likes.likeableId=pageInfos.likeableId;
-- COMMIT;
alter table likes drop primary key, add primary key (userId, likeableId);
alter table likes drop column pageId;

alter table subscriptions drop column userTrustSnapshotId;

DROP TABLE inviteEmailPairs;
drop table userDomainBonusTrust;

ALTER TABLE users ADD COLUMN isTrusted BOOLEAN NOT NULL;

DELETE FROM updates USING updates, pageInfos AS pi WHERE pi.pageId = updates.goToPageId
	AND pi.seeGroupId != '' AND pi.seeGroupId NOT IN (SELECT groupId FROM groupMembers AS gm WHERE gm.userId = updates.userId);

alter table users drop column inviteCode;
alter table users drop column karma;
alter table pageInfos drop column editKarmaLock;

alter table visits add column sessionId VARCHAR(32) NOT NULL after userId;
alter table visits add column ipAddress VARCHAR(64) NOT NULL after sessionId;
delete from groupMembers where userId=groupId;
insert into groupMembers (userId,createdAt,groupId) select id,createdAt,id from users;
alter table users drop column  inviteCode;

alter table pageInfos add column isEditorCommentIntention bool not null;
update pageInfos set isEditorCommentIntention=isEditorComment;

update pages set text=replace(text,"||||||||||","%%%%%%%%%%");
update pages set text=replace(text,"|||||||||","%%%%%%%%%");
update pages set text=replace(text,"||||||||","%%%%%%%%");
update pages set text=replace(text,"|||||||","%%%%%%%");
update pages set text=replace(text,"||||||","%%%%%%");
update pages set text=replace(text,"|||||","%%%%%");
update pages set text=replace(text,"||||","%%%%");
update pages set text=replace(text,"|||","%%%");
update pages set text=replace(text,"||","%%");

DELETE FROM changeLogs USING changeLogs, pageInfos AS commentInfos WHERE commentInfos.type='comment'
	AND changeLogs.type='newChild' AND commentInfos.pageId=changeLogs.auxPageId;
alter table marks add column isSubmitted boolean not null;
update marks set isSubmitted=1;

UPDATE
	pages AS comments, pageInfos AS commentInfos
SET
	comments.title = concat('"', CASE WHEN char_length(comments.text) > 30 THEN concat(substr(comments.text,1,27), '...') ELSE comments.text END, '"')
WHERE
	comments.pageId=commentInfos.pageId AND commentInfos.type='comment' AND comments.isLiveEdit;

DROP TABLE lastViews;
CREATE TABLE lastViews (
	/* Id of the likeable. */
	userId varchar(32) NOT NULL,
	/* The thing the user saw. */
	viewName varchar(64) NOT NULL,
	/* The last time the user viewed the thing. */
	viewedAt DATETIME NOT NULL,
	PRIMARY KEY(userId,viewName)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table updates add column dismissed boolean not null;

delete from updates where type="commentEdit";
alter table updates add column seen boolean not null;
update updates set seen=(NOT unseen);
update updates set type="changeLog" where type IN ("newParent","deleteParent","newChild","deleteChild","newTag","deleteTag","newUsedAsTag","deleteUsedAsTag","newRequirement","deleteRequirement","newRequiredBy","deleteRequiredBy","newSubject","deleteSubject","newTeacher","deleteTeacher","deletePage","undeletePage");
delete from updates where type="changeLog" and (changeLogId="0" OR changeLogId="");
alter table updates drop column unseen;

alter table subscriptions add column asMaintainer boolean not null;
update subscriptions join pageInfos on subscriptions.toId=pageInfos.pageId set subscriptions.asMaintainer=true where pageInfos.createdBy=subscriptions.userId;
update updates set type="changeLog" where type="pageInfoEdit";
