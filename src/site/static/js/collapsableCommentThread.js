'use strict';

// Directive for showing a comment thread that's collapsed behind a button
app.directive('arbCollapsableCommentThread', function($compile, $timeout, $location, $mdToast, $mdMedia, $anchorScroll, arb, autocompleteService, RecursionHelper) {
	return {
		templateUrl: 'static/html/collapsableCommentThread.html',
		scope: {
			commentId: '@',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.comment = pageService.pageMap[$scope.commentId];
			$scope.threadExpanded = false;
			$scope.threadLoaded = false;

			$scope.toggleThread = function() {
				$scope.threadExpanded = !$scope.threadExpanded;
				if (!$scope.threadExpanded || $scope.threadLoaded) {
					return;
				}
				pageService.loadCommentThread($scope.comment.pageId, {
					success: function() {
						$scope.threadLoaded = true;
					},
				});
			};
		},
	};
});
