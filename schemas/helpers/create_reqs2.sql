DELETE FROM pageInfos where pageId>="22000" AND pageId<="22999";
INSERT INTO pageInfos
(pageId,alias,currentEdit,maxEdit,createdAt,type,sortChildrenBy,createdBy)
VALUES
("22000","22000",1,1,now(),"wiki","likes",1),
("22001","22001",1,1,now(),"wiki","likes",1),
("22002","22002",1,1,now(),"wiki","likes",1),
("22003","22003",1,1,now(),"wiki","likes",1),
("22004","22004",1,1,now(),"wiki","likes",1),
("22005","22005",1,1,now(),"wiki","likes",1)
;

DELETE FROM pages where pageId>="22000" AND pageId<="22999";
INSERT INTO pages
(pageId,title,text,clickbait,edit,creatorId,createdAt,isLiveEdit)
VALUES
("22000","Page 0 (I want to learn this)","I want to learn this",     "Page 0 clickbait",1,1,now(),true),
("22001","Page 1 (Req for 0)"           ,"Page 0 requires me.",      "Page 1 clickbait",1,1,now(),true),
("22002","Page 2 (Req for 0)"           ,"Page 0 requires me.",      "Page 2 clickbait",1,1,now(),true),
("22003","Page 3 (Req for 0)"           ,"Page 0 requires me.",      "Page 3 clickbait",1,1,now(),true),
("22004","Page 4 (Teaches 1)"           ,"Teaching you about page 1","Page 4 clickbait",1,1,now(),true),
("22005","Page 5 (Req for 2)"           ,"Page 2 requires me.",      "Page 5 clickbait",1,1,now(),true)
;

DELETE FROM pagePairs where (childId>="22000" AND childId<="22999") OR (parentId>="22000" AND parentId<="22999");
INSERT INTO pagePairs
(parentId,childId,type)
VALUES
("22t","22000","tag"),
("22t","22003","tag"),
("22001","22000","requirement"),
("22002","22000","requirement"),
("22003","22000","requirement"),
("22005","22002","requirement"),
("22001","22004","subject")
;

