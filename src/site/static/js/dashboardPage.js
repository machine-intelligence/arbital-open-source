'use strict';

// Directive for the Dashboard page.
app.directive('arbDashboardPage', function(pageService, userService) {
	return {
		templateUrl: 'static/html/dashboardPage.html',
		scope: {
			data: '='
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});
