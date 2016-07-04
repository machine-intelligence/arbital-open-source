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

			// Determine which style of bar to show
			$scope.qualityTag = 'unassessed';
			if ($scope.page.taggedAsIds.includes('72')) {
				$scope.qualityTag = 'stub';
			} else if ($scope.page.taggedAsIds.includes('3rk')) {
				$scope.qualityTag = 'start';
			} else if ($scope.page.taggedAsIds.includes('4y7')) {
				$scope.qualityTag = 'c-class';
			} else if ($scope.page.taggedAsIds.includes('4yd')) {
				$scope.qualityTag = 'b-class';
			} else if ($scope.page.taggedAsIds.includes('4yf')) {
				$scope.qualityTag = 'a-class';
			} else if ($scope.page.taggedAsIds.includes('4yl')) {
				$scope.qualityTag = 'featured';
			}

			$scope.shouldShowTagsColumn = function() {
				return $scope.page.improvementTagIds.length > 0;
			};
			$scope.shouldShowTodosColumn = function() {
				return $scope.page.todos.length > 0;
			};
			$scope.shouldShowImprovements = function() {
				return !['b-class', 'a-class', 'featured'].includes($scope.qualityTag);
			};

			$scope.expanded = $scope.page.isSubscribedAsMaintainer;
			$scope.toggleExpand = function() {
				$scope.expanded = !$scope.expanded;
			};
		},
	};
});

