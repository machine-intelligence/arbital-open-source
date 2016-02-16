/* This file contains the second batch of queries used for converting the ids from base10 to base36. */

UPDATE pages SET pageId = pageIdBase36 WHERE 1;
UPDATE pages SET creatorId = creatorIdBase36 WHERE 1;
UPDATE pages SET privacyKey = privacyKeyBase36 WHERE 1;

UPDATE changeLogs SET pageId = pageIdBase36 WHERE 1;
UPDATE changeLogs SET auxPageId = auxPageIdBase36 WHERE 1;

UPDATE fixedIds SET pageId = pageIdBase36 WHERE 1;

UPDATE groupMembers SET userId = userIdBase36 WHERE 1;
UPDATE groupMembers SET groupId = groupIdBase36 WHERE 1;

UPDATE likes SET userId = userIdBase36 WHERE 1;
UPDATE likes SET pageId = pageIdBase36 WHERE 1;

UPDATE links SET parentId = parentIdBase36 WHERE 1;

UPDATE pageDomainPairs SET pageId = pageIdBase36 WHERE 1;
UPDATE pageDomainPairs SET domainId = domainIdBase36 WHERE 1;

UPDATE pageInfos SET pageId = pageIdBase36 WHERE 1;
UPDATE pageInfos SET lockedBy = lockedByBase36 WHERE 1;
UPDATE pageInfos SET seeGroupId = seeGroupIdBase36 WHERE 1;
UPDATE pageInfos SET editGroupId = editGroupIdBase36 WHERE 1;
UPDATE pageInfos SET createdBy = createdByBase36 WHERE 1;

UPDATE pagePairs SET parentId = parentIdBase36 WHERE 1;
UPDATE pagePairs SET childId = childIdBase36 WHERE 1;

UPDATE pageSummaries SET pageId = pageIdBase36 WHERE 1;

UPDATE subscriptions SET userId = userIdBase36 WHERE 1;
UPDATE subscriptions SET toId = toIdBase36 WHERE 1;

UPDATE updates SET userId = userIdBase36 WHERE 1;
UPDATE updates SET groupByPageId = groupByPageIdBase36 WHERE 1;
UPDATE updates SET groupByUserId = groupByUserIdBase36 WHERE 1;
UPDATE updates SET subscribedToId = subscribedToIdBase36 WHERE 1;
UPDATE updates SET goToPageId = goToPageIdBase36 WHERE 1;
UPDATE updates SET byUserId = byUserIdBase36 WHERE 1;

UPDATE userMasteryPairs SET userId = userIdBase36 WHERE 1;
UPDATE userMasteryPairs SET masteryId = masteryIdBase36 WHERE 1;

UPDATE users SET id = idBase36 WHERE 1;

UPDATE visits SET userId = userIdBase36 WHERE 1;
UPDATE visits SET pageId = pageIdBase36 WHERE 1;

UPDATE votes SET userId = userIdBase36 WHERE 1;
UPDATE votes SET pageId = pageIdBase36 WHERE 1;



ALTER TABLE pages DROP pageIdBase36 ;
ALTER TABLE pages DROP creatorIdBase36 ;
ALTER TABLE pages DROP privacyKeyBase36 ;

ALTER TABLE changeLogs DROP pageIdBase36 ;
ALTER TABLE changeLogs DROP auxPageIdBase36 ;

ALTER TABLE fixedIds DROP pageIdBase36 ;

ALTER TABLE groupMembers DROP userIdBase36 ;
ALTER TABLE groupMembers DROP groupIdBase36 ;

ALTER TABLE likes DROP userIdBase36 ;
ALTER TABLE likes DROP pageIdBase36 ;

ALTER TABLE links DROP parentIdBase36 ;

ALTER TABLE pageDomainPairs DROP pageIdBase36 ;
ALTER TABLE pageDomainPairs DROP domainIdBase36 ;

ALTER TABLE pageInfos DROP pageIdBase36 ;
ALTER TABLE pageInfos DROP lockedByBase36 ;
ALTER TABLE pageInfos DROP seeGroupIdBase36 ;
ALTER TABLE pageInfos DROP editGroupIdBase36 ;
ALTER TABLE pageInfos DROP createdByBase36 ;

ALTER TABLE pagePairs DROP parentIdBase36 ;
ALTER TABLE pagePairs DROP childIdBase36 ;

ALTER TABLE pageSummaries DROP pageIdBase36 ;

ALTER TABLE subscriptions DROP userIdBase36 ;
ALTER TABLE subscriptions DROP toIdBase36 ;

ALTER TABLE updates DROP userIdBase36 ;
ALTER TABLE updates DROP groupByPageIdBase36 ;
ALTER TABLE updates DROP groupByUserIdBase36 ;
ALTER TABLE updates DROP subscribedToIdBase36 ;
ALTER TABLE updates DROP goToPageIdBase36 ;
ALTER TABLE updates DROP byUserIdBase36 ;

ALTER TABLE userMasteryPairs DROP userIdBase36 ;
ALTER TABLE userMasteryPairs DROP masteryIdBase36 ;

ALTER TABLE users DROP idBase36 ;

ALTER TABLE visits DROP userIdBase36 ;
ALTER TABLE visits DROP pageIdBase36 ;

ALTER TABLE votes DROP userIdBase36 ;
ALTER TABLE votes DROP pageIdBase36 ;



ALTER TABLE pages DROP pageIdProcessed ;
ALTER TABLE pages DROP creatorIdProcessed ;
ALTER TABLE pages DROP privacyKeyProcessed ;

ALTER TABLE changeLogs DROP pageIdProcessed ;
ALTER TABLE changeLogs DROP auxPageIdProcessed ;

ALTER TABLE fixedIds DROP pageIdProcessed ;

ALTER TABLE groupMembers DROP userIdProcessed ;
ALTER TABLE groupMembers DROP groupIdProcessed ;

ALTER TABLE likes DROP userIdProcessed ;
ALTER TABLE likes DROP pageIdProcessed ;

ALTER TABLE links DROP parentIdProcessed ;

ALTER TABLE pageDomainPairs DROP pageIdProcessed ;
ALTER TABLE pageDomainPairs DROP domainIdProcessed ;

ALTER TABLE pageInfos DROP pageIdProcessed ;
ALTER TABLE pageInfos DROP lockedByProcessed ;
ALTER TABLE pageInfos DROP seeGroupIdProcessed ;
ALTER TABLE pageInfos DROP editGroupIdProcessed ;
ALTER TABLE pageInfos DROP createdByProcessed ;

ALTER TABLE pagePairs DROP parentIdProcessed ;
ALTER TABLE pagePairs DROP childIdProcessed ;

ALTER TABLE pageSummaries DROP pageIdProcessed ;

ALTER TABLE subscriptions DROP userIdProcessed ;
ALTER TABLE subscriptions DROP toIdProcessed ;

ALTER TABLE updates DROP userIdProcessed ;
ALTER TABLE updates DROP groupByPageIdProcessed ;
ALTER TABLE updates DROP groupByUserIdProcessed ;
ALTER TABLE updates DROP subscribedToIdProcessed ;
ALTER TABLE updates DROP goToPageIdProcessed ;
ALTER TABLE updates DROP byUserIdProcessed ;

ALTER TABLE userMasteryPairs DROP userIdProcessed ;
ALTER TABLE userMasteryPairs DROP masteryIdProcessed ;

ALTER TABLE users DROP idProcessed ;

ALTER TABLE visits DROP userIdProcessed ;
ALTER TABLE visits DROP pageIdProcessed ;

ALTER TABLE votes DROP userIdProcessed ;
ALTER TABLE votes DROP pageIdProcessed ;








