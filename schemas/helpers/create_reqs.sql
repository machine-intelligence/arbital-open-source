DELETE FROM pageInfos where pageId>=111111111111000000 AND pageId<=111111111111999999;
INSERT INTO pageInfos
(pageId,alias,currentEdit,maxEdit,createdAt,type,sortChildrenBy,createdBy)
VALUES
(111111111111000000,"111111111111000000",1,1,now(),"wiki","likes",1),
(111111111111000001,"111111111111000001",1,1,now(),"wiki","likes",1),
(111111111111000002,"111111111111000002",1,1,now(),"wiki","likes",1),
(111111111111000003,"111111111111000003",1,1,now(),"wiki","likes",1),
(111111111111000004,"111111111111000004",1,1,now(),"wiki","likes",1),
(111111111111000005,"111111111111000005",1,1,now(),"lens","likes",1),
(111111111111000006,"111111111111000006",1,1,now(),"wiki","likes",1),
(111111111111000007,"111111111111000007",1,1,now(),"lens","likes",1),
(111111111111000008,"111111111111000008",1,1,now(),"lens","likes",1),
(111111111111000009,"111111111111000009",1,1,now(),"wiki","likes",1),
(111111111111000010,"111111111111000010",1,1,now(),"lens","likes",1)
;

DELETE FROM pages where pageId>=111111111111000000 AND pageId<=111111111111999999;
INSERT INTO pages
(pageId,title,text,edit,creatorId,createdAt,isCurrentEdit)
VALUES
(111111111111000000,"Page 0 (I want to learn this)","I want to learn this",1,1,now(),true),
(111111111111000001,"Page 1 (Teaches 0)","Teaching you about page 0",1,1,now(),true),
(111111111111000002,"Page 2 (Req for 1)","Page 1 requires me.",1,1,now(),true),
(111111111111000003,"Page 3 (Req for 1)","Page 1 requires me",1,1,now(),true),
(111111111111000004,"Page 4 (Req for 1)","Page 1 requires me",1,1,now(),true),
(111111111111000005,"Page 5 (Teaches 2)","Teaching you about page 2",1,1,now(),true),
(111111111111000006,"Page 6 (Teaches 2)","Teaching you about page 2",1,1,now(),true),
(111111111111000007,"Page 7 (Teaches 3)","Teaching you about page 3",1,1,now(),true),
(111111111111000008,"Page 8 (Teaches 4)","Teaching you about page 4",1,1,now(),true),
(111111111111000009,"Page 9 (Req for 7)","Page 7 requires me",1,1,now(),true),
(111111111111000010,"Page 10 (Teaches 9)","Teaching you about page 9",1,1,now(),true)
;

DELETE FROM pagePairs where (childId>=111111111111000000 AND childId<=111111111111999999) OR (parentId>=111111111111000000 AND parentId<=111111111111999999);
INSERT INTO pagePairs
(parentId,childId,type)
VALUES
(111111111111000006,111111111111000005,"parent"),
(111111111111000006,111111111111000010,"parent"),
(111111111111000006,111111111111000007,"parent"),
(111111111111000006,111111111111000008,"parent"),
(111111111111000002,111111111111000001,"requirement"),
(111111111111000003,111111111111000001,"requirement"),
(111111111111000004,111111111111000001,"requirement"),
(111111111111000009,111111111111000001,"requirement"),
(111111111111000009,111111111111000007,"requirement"),
(111111111111000000,111111111111000001,"subject"),
(111111111111000002,111111111111000005,"subject"),
(111111111111000002,111111111111000006,"subject"),
(111111111111000003,111111111111000007,"subject"),
(111111111111000004,111111111111000008,"subject"),
(111111111111000009,111111111111000010,"subject")
;
