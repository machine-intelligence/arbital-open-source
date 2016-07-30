'use strict';

import app from './angular.ts';

// Directive for the Admin Dashboard page.
app.directive('arbAdminDashboardPage', function(arb) {
	return {
		templateUrl: versionUrl('static/html/adminDashboardPage.html'),
		scope: {
			data: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
