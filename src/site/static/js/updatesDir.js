"use strict";

// Directive for the Updates page.
app.directive("arbUpdates", function($compile, $location, $rootScope, pageService, userService) {
	return {
		templateUrl: "/static/html/updatesDir.html",
		scope: {
			updateGroups: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			element.on("click", ".group-subscribe-to-link", function(event) {
				pageService.subscribeTo($(event.target));
				return false;
			});
		},
	};
});
