/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table updates add column byUserId bigint not null;
alter table pages add column metaText mediumtext not null;
alter table pages drop column privacyKey;
drop index name on groups;
drop index name_2 on groups;
drop index name_3 on groups;
drop index name_4 on groups;
CREATE UNIQUE INDEX alias ON groups(alias);
