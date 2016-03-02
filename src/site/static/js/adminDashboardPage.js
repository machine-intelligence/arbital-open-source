"use strict";

// Directive for the Admin Dashboard page.
app.directive("arbAdminDashboardPage", function(pageService, userService) {
	return {
		templateUrl: "static/html/adminDashboardPage.html",
		scope: {
			data: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});
