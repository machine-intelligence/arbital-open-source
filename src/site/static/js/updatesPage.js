"use strict";

// Directive for the Updates page.
app.directive("arbUpdates", function($compile, $location, $rootScope, pageService, userService) {
	return {
		templateUrl: "/static/html/updatesPage.html",
		scope: {
			updateGroups: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
		},
	};
});
