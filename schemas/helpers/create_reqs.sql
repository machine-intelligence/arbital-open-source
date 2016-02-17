DELETE FROM pageInfos where pageId>="11000" AND pageId<="11999999";
INSERT INTO pageInfos
(pageId,alias,currentEdit,maxEdit,createdAt,type,sortChildrenBy,createdBy)
VALUES
("11000","11000",1,1,now(),"wiki","likes",1),
("11001","11001",1,1,now(),"wiki","likes",1),
("11002","11002",1,1,now(),"wiki","likes",1),
("11003","11003",1,1,now(),"wiki","likes",1),
("11004","11004",1,1,now(),"wiki","likes",1),
("11005","11005",1,1,now(),"lens","likes",1),
("11006","11006",1,1,now(),"wiki","likes",1),
("11007","11007",1,1,now(),"lens","likes",1),
("11008","11008",1,1,now(),"lens","likes",1),
("11009","11009",1,1,now(),"wiki","likes",1),
("11010","11010",1,1,now(),"lens","likes",1)
;

DELETE FROM pages where pageId>="11000" AND pageId<="11999";
INSERT INTO pages
(pageId,title,text,clickbait,edit,creatorId,createdAt,isCurrentEdit)
VALUES
("11000","Page 0 (I want to learn this)","I want to learn this",     "Page 0 clickbait" , 1,1,now(),true),
("11001","Page 1 (Teaches 0)"           ,"Teaching you about page 0","Page 1 clickbait" , 1,1,now(),true),
("11002","Page 2 (Req for 1)"           ,"Page 1 requires me.",      "Page 2 clickbait" , 1,1,now(),true),
("11003","Page 3 (Req for 1)"           ,"Page 1 requires me",       "Page 3 clickbait" , 1,1,now(),true),
("11004","Page 4 (Req for 1)"           ,"Page 1 requires me",       "Page 4 clickbait" , 1,1,now(),true),
("11005","Page 5 (Teaches 2)"           ,"Teaching you about page 2","Page 5 clickbait" , 1,1,now(),true),
("11006","Page 6 (Teaches 2)"           ,"Teaching you about page 2","Page 6 clickbait" , 1,1,now(),true),
("11007","Page 7 (Teaches 3)"           ,"Teaching you about page 3","Page 7 clickbait" , 1,1,now(),true),
("11008","Page 8 (Teaches 4)"           ,"Teaching you about page 4","Page 8 clickbait" , 1,1,now(),true),
("11009","Page 9 (Req for 7)"           ,"Page 7 requires me",       "Page 9 clickbait" , 1,1,now(),true),
("11010","Page 10 (Teaches 9)"          ,"Teaching you about page 9","Page 10 clickbait", 1,1,now(),true)
;

DELETE FROM pagePairs where (childId>="11000" AND childId<="11999") OR (parentId>="11000" AND parentId<="11999");
INSERT INTO pagePairs
(parentId,childId,type)
VALUES
("11006","11005","parent"),
("11006","11010","parent"),
("11006","11007","parent"),
("11006","11008","parent"),
("11002","11001","requirement"),
("11003","11001","requirement"),
("11004","11001","requirement"),
("11009","11001","requirement"),
("11009","11007","requirement"),
("11000","11001","subject"),
("11002","11005","subject"),
("11002","11006","subject"),
("11003","11007","subject"),
("11004","11008","subject"),
("11009","11010","subject")
;
