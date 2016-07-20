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
		},
	};
});
