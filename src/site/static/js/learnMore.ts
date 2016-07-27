import app from './angular.ts';

// Directive to show the discussion section for a page
app.directive('arbLearnMore', function($compile, $location, $timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/learnMore.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			// Return true if there are any learn more suggestions to show
			$scope.hasLearnMore = function() {
				return Object.keys($scope.page.learnMoreTaughtMap).length > 0 ||
						Object.keys($scope.page.learnMoreCoveredMap).length > 0 ||
						 Object.keys($scope.page.learnMoreRequiredMap).length > 0;
			};
		},
	};
});
