'use strict';

// Directive for the Dashboard page.
app.directive('arbDashboardPage', function(arb) {
	return {
		templateUrl: 'static/html/dashboardPage.html',
		scope: {
			data: '='
		},
		controller: function($scope) {
			$scope.arb = arb;
			
		},
	};
});
