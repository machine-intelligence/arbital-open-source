"use strict";

// Directive for the Updates page.
app.directive("arbUpdates", function($compile, $location, $rootScope, pageService, userService) {
	return {
		templateUrl: "/static/html/updatesPage.html",
		scope: {
			updateGroups: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});
