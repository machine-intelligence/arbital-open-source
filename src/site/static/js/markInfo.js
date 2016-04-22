'use strict';

// Directive for showing a window for creating/editing a mark
app.directive('arbMarkInfo', function($interval, pageService, userService, autocompleteService) {
	return {
		templateUrl: 'static/html/markInfo.html',
		scope: {
			// Id of the query mark that was created.
			markId: '@',
			// Set to true if the user just created this mark.
			isNew: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.mark = pageService.markMap[$scope.markId];
			$scope.isOnPage = $scope.mark.pageId == pageService.getCurrentPageId();

			// Call to resolve the mark with the given page.
			$scope.resolveWith = function(pageId) {
				pageService.updateMark({
					markId: $scope.markId,
					resolvedPageId: $scope.mark.pageId,
				});
				$scope.mark.resolvedPageId = pageId;
				$scope.mark.resolvedBy = userService.user.id;
				$scope.hidePopup();
			};

			// Called when an author wants to resolve the mark.
			$scope.dismissMark = function() {
				pageService.updateMark({
					markId: $scope.markId,
					dismiss: true,
				});
				$scope.mark.resolvedPageId = '';
				$scope.mark.resolvedBy = userService.user.id;
				$scope.hidePopup({dismiss: true});
			};
		},
		link: function(scope, element, attrs) {
			// Hide current event window, if it makes sense.
			scope.hidePopup = function(result) {
				if (scope.isOnPage) {
					pageService.hidePopup(result);
				}
			};
		},
	};
});
