// Directive to show the discussion section for a page
app.directive("arbDiscussion", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "static/html/discussion.html",
		scope: {
			pageId: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];

			// Sort subpages
			$scope.page.subpageIds = ($scope.page.questionIds || []).concat($scope.page.commentIds || []);
			$scope.page.subpageIds.sort(pageService.getChildSortFunc("likes"));

			// Process user clicking on New Comment button
			$scope.newCommentClick = function() {
				pageService.newComment({
					parentPageId: $scope.pageId,
					success: function(newCommentId) {
						$scope.newCommentId = newCommentId;
					},
				});
			};

			// Called when the user is done editing the new comment
			$scope.newCommentDone = function(result) {
				$scope.newCommentId = undefined;
				if (!result.discard) {
					pageService.newCommentCreated(result.pageId);
				}
			};
		},
	};
});
