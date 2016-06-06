/* This file contains the recent changes to schemas, sorted from oldest to newest. */
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

alter table users add column showAdvancedEditorMode bool not null;
delete from invites where domainId="";
CREATE TABLE pageToDomainSubmissions (
	/* Id of the submitted page. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* Id of the domain it's submitted to. FK into pageInfos. */
	domainId VARCHAR(32) NOT NULL,
	/* When this submission was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the user who submitted. FK into users. */
	submitterId VARCHAR(32) NOT NULL,

	/* When this submission was approved. */
	approvedAt DATETIME NOT NULL,
	/* Id of the user who approved the submission. FK into users. */
	approverId VARCHAR(32) NOT NULL,

	PRIMARY KEY(pageId,domainId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table pageInfos add column featuredAt datetime not null;
alter table pageInfos add column isResolved bool not null;
create index pageId on visits (pageId);
