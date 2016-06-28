/* This file contains the recent changes to schemas, sorted from oldest to newest. */

CREATE TABLE copyLenses (
	/* Id of the lens relationships. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page that has the lens. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* Id of the lens page. FK into pageInfos. */
	lensId VARCHAR(32) NOT NULL,
	/* Ordering index when sorting the page's lenses. */
	lensIndex INT NOT NULL,
	/* Lens name that shows up in the tab. */
	lensName VARCHAR(32) NOT NULL,
	/* Id of the user who created the relationship. FK into users. */
	createdBy VARCHAR(32) NOT NULL,
	/* When this lens relationship was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the last user who updated the relationship. FK into users. */
	updatedBy VARCHAR(32) NOT NULL,
	/* When this relationship was updated last. */
	updatedAt DATETIME NOT NULL,

	UNIQUE(pageId,lensId),
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

insert into copyLenses (pageId,lensId,lensIndex,lensName,createdBy,createdAt,updatedBy,updatedAt) select pp.parentId,pp.childId,pi.lensIndex,trim(substring_index(p.title,":",-1)),pp.creatorId,pp.createdAt,pp.creatorId,pp.createdAt from pagePairs as pp join pageInfos as pi on (pp.childId=pi.pageId) join pages as p on (pi.pageId=p.pageId and pi.currentEdit=p.edit) where pi.type="lens" and pp.type="parent"; 
create table copyLenses2 like copyLenses;
insert into copyLenses2 select * from copyLenses;
update copyLenses as l1 set l1.lensIndex=l1.id where l1.pageId in (select l3.pageId from copyLenses2 as l3 group by 1 having sum(l3.lensIndex=0)>1);
update copyLenses as l1 set l1.lensIndex=l1.id where l1.pageId in (select l3.pageId from copyLenses2 as l3 group by 1 having sum(l3.lensIndex=1)>1);
insert into lenses select * from copyLenses;
update lenses as l1 set l1.lensIndex=l1.lensIndex-(select min(l2.lensIndex) from copyLenses as l2 where l1.pageId=l2.pageId) order by l1.lensIndex;
drop table copyLenses;
drop table copyLenses2;
/* Might need to delete some rows? */
ALTER TABLE lenses ADD CONSTRAINT lensId UNIQUE (lensId);

update pageInfos set type="wiki" where type="lens";
alter table pageInfos drop column lensIndex;
