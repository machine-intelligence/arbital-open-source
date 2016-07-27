'use strict';

import app from './angular.ts';

// TypeScript doesn't currently include Array.prototype.includes from ES7.
// Declare it ourselves for now.
declare global {
	interface Array<T> {
		includes(search: T): boolean;
	}
}

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
			if ($scope.page.taggedAsIds.includes('4v')) {
				$scope.qualityTag = 'wip';
			} else if ($scope.page.taggedAsIds.includes('4v4')) {
				$scope.qualityTag = 'still-needs-work';
			} else if ($scope.page.taggedAsIds.includes('72')) {
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

			$scope.shouldShowTags = function() {
				return $scope.page.improvementTagIds.length > 0;
			};
			$scope.shouldShowTodos = function() {
				return $scope.page.todos.length > 0;
			};
			$scope.shouldShowImprovements = function() {
				return $scope.shouldShowTags() || $scope.shouldShowTodos();
			};
			$scope.showQualityBar = function() {
				return !['b-class', 'a-class', 'featured'].includes($scope.qualityTag);
			};

			$scope.expanded = $scope.page.isSubscribedAsMaintainer;
			$scope.toggleExpand = function() {
				$scope.expanded = !$scope.expanded;
			};
		},
	};
});

