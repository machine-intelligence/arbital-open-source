/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table userMasteryPairs add column level int not null;
update userMasteryPairs set level=1 where has;

update pagePairs set level=4 where level=3;
update pagePairs set level=3 where level=2;
update pagePairs set level=2 where level=1;
update pagePairs set level=1 where level=0;
