'use strict';

// Directive for navigating the path
app.directive('arbPathNav', function($location, arb) {
	return {
		templateUrl: versionUrl('static/html/pathNav.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.path = arb.stateService.path;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			$scope.stopLearning = function() {
				arb.pathService.abandonPath();
			};
		},
	};
});

