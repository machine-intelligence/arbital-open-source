"use strict";

// arb-index directive displays a set of featured domains
app.directive("arbIndex", function(pageService, userService) {
	return {
		templateUrl: "/static/html/indexPage.html",
		scope: {
			featuredDomains: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
		},
	};
});
