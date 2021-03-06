'use strict';

import app from './angular.ts';

// Directive for the Dashboard page.
app.directive('arbDashboardPage', function(arb) {
	return {
		templateUrl: versionUrl('static/html/dashboardPage.html'),
		scope: {
			data: '='
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
