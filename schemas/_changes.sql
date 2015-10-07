/* This file contains the recent changes to schemas, sorted from newest to oldest. */
alter table pages drop column privacyKey;
alter table pages add column metaText mediumtext not null;
alter table updates add column byUserId bigint not null;
