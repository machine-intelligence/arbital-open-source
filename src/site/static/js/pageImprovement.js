'use strict';

// arb-page-improvement is the directive which shows what improvements should be made for a page.
app.directive('arbPageImprovement', function($timeout, $http, $compile, arb) {
	return {
		templateUrl: versionUrl('static/html/pageImprovement.html'),
		scope: {
			page: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.shouldShowTagsColumn = function() {
				return $scope.page.improvementTagIds.length > 0;
			};
			$scope.shouldShowTodosColumn = function() {
				return $scope.page.todos.length > 0;
			};
			$scope.shouldShowImprovements = function() {
				return $scope.shouldShowTagsColumn() || $scope.shouldShowTodosColumn();
			};

			$scope.expanded = $scope.page.isSubscribedAsMaintainer;
			$scope.toggleExpand = function() {
				$scope.expanded = !$scope.expanded;
			};
		},
	};
});

