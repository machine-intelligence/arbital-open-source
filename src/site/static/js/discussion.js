// Directive to show the discussion section for a page
app.directive("arbDiscussion", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/discussion.html",
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

			// Load the comment edit 
			// options: {
			//	commentId: comment to load
			//	success: callback
			// }
			$scope.loadCommentEdit = function(options) {
				pageService.loadEdit({
					pageAlias: options.commentId,
					success: function() {
						if (options.success) options.success(options.commentId);
					},
					error: function(error) {
						// TODO
					},
				});
			};

			// Create a new comment; optionally it's a reply to the given commentId
			// options: {
			//	replyToId: (optional) comment id this will be a reply to
			//	success: callback
			// }
			$scope.newComment = function(options) {
				var parentIds = [$scope.page.pageId];
				if (options.replyToId) {
					parentIds.push(options.replyToId);
				}
				// Create new comment
				pageService.getNewPage({
					type: "comment",
					parentIds: parentIds,
					success: function(newCommentId) {
						$scope.loadCommentEdit({commentId: newCommentId, success: options.success});
					},
				});
			};

			// Called when the user created a new subpage
			$scope.newSubpageCreated = function(result) {
				// TODO: dynamically add the comment
				window.location.href = pageService.getPageUrl($scope.pageId) + "#subpage-" + result.pageId;
				window.location.reload();
			};

			// Process user clicking on New Comment button
			$scope.newCommentClick = function() {
				$scope.newComment({
					success: function(newCommentId) {
						$scope.newCommentId = newCommentId;
					},
				});
			};

			// Called when the user is done editing the new comment
			$scope.newCommentDone = function(result) {
				$scope.newCommentId = undefined;
				if (!result.discard) {
					$scope.newSubpageCreated(result);
				}
			};
		},
	};
});
