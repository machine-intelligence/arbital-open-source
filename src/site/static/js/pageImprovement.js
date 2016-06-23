'use strict';

// arb-page-improvement is the directive which shows what improvements should be made for a page.
app.directive('arbPageImprovement', function($timeout, $http, $compile, arb) {
	return {
		templateUrl: versionUrl('static/html/pageImprovement.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			$scope.improvementTagIds = $scope.page.taggedAsIds.filter(function(tagId) {
				return arb.stateService.globalData.improvementTagIds.indexOf(tagId) >= 0;
			});

			$scope.shouldShowTagsColumn = function() {
				return $scope.improvementTagIds.length > 0;
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
		link: function(scope, element, attrs) {
		},
	};
});

