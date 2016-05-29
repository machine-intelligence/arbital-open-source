// Directive to show the discussion section for a page
app.directive('arbPageDiscussion', function($compile, $location, $timeout, arb) {
	return {
		templateUrl: 'static/html/pageDiscussion.html',
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.pageService.pageMap[$scope.pageId];
			$scope.page.subpageIds = $scope.page.commentIds || [];
			$scope.page.subpageIds.sort(arb.pageService.getChildSortFunc('likes'));

			// Process user clicking on New Comment button
			$scope.newCommentClick = function() {
				arb.pageService.newComment({
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
					arb.pageService.newCommentCreated(result.pageId);
				}
			};

			// Track (globally) whether or not to show editor comments.
			arb.pageService.setShowEditorComments($scope.page.creatorIds.indexOf(arb.userService.user.id) >= 0);
			if (!arb.pageService.getShowEditorComments() && $location.hash()) {
				// If hash points to a subpage for editors, show it
				var matches = (new RegExp('^subpage-' + aliasMatch + '$')).exec($location.hash());
				if (matches) {
					var page = arb.pageService.pageMap[matches[1]];
					if (page) {
						arb.pageService.setShowEditorComments(page.isEditorComment);
					}
				}
			}

			$scope.toggleEditorComments = function() {
				arb.pageService.setShowEditorComments(!arb.pageService.getShowEditorComments());
			};

			// Compute how many visible comments there are.
			$scope.visibleCommentCount = function() {
				var count = 0;
				for (var n = 0; n < $scope.page.commentIds.length; n++) {
					var commentId = $scope.page.commentIds[n];
					count += (!arb.pageService.pageMap[commentId].isEditorComment || arb.pageService.getShowEditorComments()) ? 1 : 0;
				}
				return count;
			};
		},
	};
});
