'use strict';

import app from './angular.ts';

// Directive for the User page.
app.directive('arbUserPage', function(arb) {
	return {
		templateUrl: versionUrl('static/html/userPage.html'),
		scope: {
			userId: '@',
			userPageData: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
