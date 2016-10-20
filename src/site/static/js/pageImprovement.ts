'use strict';

import app from './angular.ts';

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
			$scope.qualityTag = arb.pageService.getQualityTag($scope.page.tagIds);

			$scope.shouldShowTags = function() {
				return $scope.page.improvementTagIds.length > 0;
			};
			$scope.shouldShowTodos = function() {
				return $scope.page.todos.length > 0;
			};
			$scope.shouldShowRedLinks = function() {
				return Object.keys($scope.page.redAliases).length > 0;
			};
			$scope.shouldShowImprovements = function() {
				return ['a-class', 'featured'].indexOf($scope.qualityTag) < 0;
			};
		},
	};
});

