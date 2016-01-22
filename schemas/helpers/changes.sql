/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table users add column ignoreMathjax bool not null;
alter table pageInfos add column lensIndex int not null;
alter table userMasteryPairs drop column isManuallySet;
alter table userMasteryPairs add column wants boolean not null;
alter table pageInfos add column isRequisite BOOL NOT NULL;
alter table pageInfos add column indirectTeacher BOOL NOT NULL;
