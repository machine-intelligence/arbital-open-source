// Directive to show the discussion section for a page
app.directive('arbDiscussion', function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: 'static/html/discussion.html',
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.page.subpageIds = $scope.page.commentIds || [];
			$scope.page.subpageIds.sort(pageService.getChildSortFunc('likes'));

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

			// Track (globally) whether or not to show editor comments.
			userService.showEditorComments = userService.user.id in $scope.page.creatorIds;
			if (!userService.showEditorComments && $location.hash()) {
				// If hash points to a subpage for editors, show it
				var matches = (new RegExp('^subpage-' + aliasMatch + '$')).exec($location.hash());
				if (matches) {
					var page = pageService.pageMap[matches[1]];
					if (page) {
						userService.showEditorComments = page.isEditorComment;
					}
				}
			}

			$scope.toggleEditorComments = function() {
				userService.showEditorComments = !userService.showEditorComments;
			};

			// Compute how many visible comments there are.
			$scope.visibleCommentCount = function() {
				var count = 0;
				for (var n = 0; n < $scope.page.commentIds.length; n++) {
					var commentId = $scope.page.commentIds[n];
					count += (!pageService.pageMap[commentId].isEditorComment || userService.showEditorComments) ? 1 : 0;
				}
				return count;
			};
		},
	};
});
