"use strict";

// Directive for the User page.
app.directive("arbUserPage", function(pageService, userService, $location) {
	return {
		templateUrl: "/static/html/userPage.html",
		scope: {
			idsMap: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
		},
	};
});
