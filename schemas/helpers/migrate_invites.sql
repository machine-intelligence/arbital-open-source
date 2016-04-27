/* Move any invite codes associated with karma > 0 to invites table */
INSERT INTO invites (code, type, senderId, domainId, createdAt)
VALUES ("BAYES", "group", "0", "", NOW());
INSERT INTO invites (code, type, senderId, domainId, createdAt)
VALUES ("TRUTH", "group", "0", "", NOW());
INSERT INTO invites (code, type, senderId, domainId, createdAt)
VALUES ("ALEXEI", "group", "0", "", NOW());
INSERT INTO invites (code, type, senderId, domainId, createdAt)
VALUES ("LESSWRONG", "group", "0", "", NOW());

/* Add a row to inviteEmailPairs for every user with karma > 0 and either TRUTH, BAYES, ALEXEI, or LESSWRONG inviteCode */
INSERT INTO inviteEmailPairs(code, inviteeEmail, claimingUserId, claimedAt)
SELECT inviteCode, email, id, NOW()
FROM users
WHERE karma > 0
AND (inviteCode="BAYES" OR inviteCode="TRUTH" OR inviteCode="ALEXEI" OR inviteCode="LESSWRONG");

/* Add a row to inviteEmailPairs for every user with karma > 0 and inviteCode = "", set code to TRUTH */
INSERT INTO inviteEmailPairs(code, inviteeEmail, claimingUserId, claimedAt)
SELECT "TRUTH", email, id, NOW()
FROM users
WHERE karma > 0
AND inviteCode="";

/* Add a row to userDomainBonusTrust in general domain ("") for each user with karma > 0 */
INSERT INTO userDomainBonusTrust(userId, domainId, bonusEditTrust)
SELECT id, "", karma
FROM users
WHERE karma > 0
ON DUPLICATE KEY UPDATE bonusEditTrust=VALUES(bonusEditTrust);
