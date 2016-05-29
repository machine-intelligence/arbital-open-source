'use strict';

// Directive for the Admin Dashboard page.
app.directive('arbAdminDashboardPage', function(arb) {
	return {
		templateUrl: 'static/html/adminDashboardPage.html',
		scope: {
			data: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
