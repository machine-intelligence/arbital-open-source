/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table updates add column byUserId bigint not null;
alter table pages add column metaText mediumtext not null;
alter table pages drop column privacyKey;
