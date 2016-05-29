'use strict';

// Directive to show the marks section for a page
app.directive('arbMarks', function($compile, $location, $timeout, $rootScope, arb) {
	return {
		templateUrl: 'static/html/marks.html',
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			
			$scope.page = pageService.pageMap[$scope.pageId];

			// Track (globally) whether or not to show all marks.
			$scope.isPageCreator = $scope.page.creatorIds.indexOf(userService.user.id) >= 0;
			userService.showAllMarks = $scope.isPageCreator;

			// Compute which marks to show.
			var computeMarkIds = function() {
				if ($scope.$$destroyed) return;
				$scope.markIds = $scope.page.markIds.filter(function(markId) {
					var mark = pageService.markMap[markId];
					if ($location.search().markId === markId) return true;
					if (!mark.isCurrentUserOwned && !userService.showAllMarks) return false;
					return mark.type === 'query' && mark.text.length > 0 && !mark.resolvedBy;
				});
			};
			computeMarkIds();

			$scope.$watch(function() {
				return $scope.page.markIds.length;
			}, function() {
				computeMarkIds();
			});

			$scope.newQuery = function() {
				$rootScope.$broadcast('newQueryMark');
			};

			// Called to show a popup for the given mark.
			$scope.bringUpMark = function(markId) {
				pageService.showPopup({
					title: 'Edit query mark',
					$element: $compile('<div arb-query-info mark-id="' + markId +
						'" in-popup="::true' +
						'"></div>')($rootScope),
					persistent: true,
				}, function(result) {
					computeMarkIds();
				});
			};

			// Check if ?markId is set and we need to take care of it.
			var searchMarkId = $location.search().markId;
			if (searchMarkId && $scope.markIds.indexOf(searchMarkId) >= 0) {
				var mark = pageService.markMap[searchMarkId];
				if (!mark.anchorContext) {
					// The mark is not attached, so it can only be managed through the marks section
					$scope.bringUpMark(searchMarkId);
				}
			}

			$scope.toggleAllMarks = function() {
				userService.showAllMarks = !userService.showAllMarks;
				computeMarkIds();
			};
		},
	};
});
