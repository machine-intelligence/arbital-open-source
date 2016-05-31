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
