"use strict";

// arb-group-index directive displays a set of links to pages
app.directive("arbGroupIndex", function(pageService, userService) {
	return {
		templateUrl: "/static/html/groupIndexPage.html",
		scope: {
			groupId: "@",
			idsMap: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.group = pageService.pageMap[scope.groupId];
		},
	};
});
