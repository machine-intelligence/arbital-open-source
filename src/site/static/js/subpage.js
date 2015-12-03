"use strict";

// Directive for showing a subpage.
app.directive("arbSubpage", function ($compile, $timeout, $location, pageService, userService, autocompleteService, RecursionHelper) {
	return {
		templateUrl: "/static/html/subpage.html",
		scope: {
			pageId: "@",  // id of this subpage
			lensId: "@",  // id of the lens this subpage belongs to
			parentSubpageId: "@",  // id of the parent subpage, if there is one
			newCommentFn: "&",
			loadCommentEditFn: "&",
			newSubpageCreatedFn: "&",
		},
		controller: function ($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.isComment = $scope.page.type === "comment";
			$scope.isQuestion = $scope.page.type === "question";
			if ($scope.isComment) {
				$scope.page.subpageIds = $scope.page.commentIds;
			} else if ($scope.isQuestion) {
				$scope.page.subpageIds = $scope.page.answerIds;
			}
			$scope.isCollapsed = false;

			if ($scope.isComment) {
				$scope.myUrl = pageService.getPageUrl($scope.lensId) + "#subpage-" + $scope.page.pageId;
			} else {
				$scope.myUrl = pageService.getPageUrl($scope.page.pageId);
			}

			// Gah! Stupid! TODO: move these to pageServie
			$scope.newCommentFn2 = function(options) {
				return $scope.newCommentFn({options: options});
			};
			$scope.loadCommentEditFn2 = function(options) {
				return $scope.loadCommentEditFn({options: options});
			};
			$scope.newSubpageCreatedFn2 = function(result) {
				return $scope.newSubpageCreatedFn({result: result});
			};

			// Called when the user collapses/expands this subpage
			$scope.collapseToggle = function() {
				$scope.isCollapsed = !$scope.isCollapsed;
			};

			// Called when the user wants to edit the subpage
			$scope.editSubpage = function(event) {
				if (!event.ctrlKey) {
					$scope.loadCommentEditFn({options: {
						commentId: $scope.page.pageId,
						success: function() {
							$scope.editing = true;
						},
					}});
					event.preventDefault();
				}
			};

			// Called when the user is done editing the subpage
			$scope.editDone = function(result) {
				$scope.editing = false;
				if (!result.discard) {
					$scope.newSubpageCreatedFn({result: result});
				}
			};

			// Called to create a new reply
			$scope.newReply = function() {
				$scope.newCommentFn({options: {
					replyToId: $scope.page.pageId,
					success: function(newCommentId) {
						$scope.newReplyId = newCommentId;
					},
				}});
			};

			// Called when the user is done with the new reply
			$scope.newReplyDone = function(result) {
				$scope.newReplyId = undefined;
				if (!result.discard) {
					$scope.newSubpageCreatedFn({result: result});
				}
			};
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});
