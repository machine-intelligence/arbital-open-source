"use strict";

// Directive for showing a subpage.
app.directive("arbSubpage", function ($compile, $timeout, $location, $mdToast, pageService, userService, autocompleteService, RecursionHelper) {
	return {
		templateUrl: "static/html/subpage.html",
		scope: {
			pageId: "@",  // id of this subpage
			lensId: "@",  // id of the lens this subpage belongs to
			parentSubpageId: "@",  // id of the parent subpage, if there is one
		},
		controller: function ($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.isComment = $scope.page.type === "comment";
			$scope.isQuestion = $scope.page.type === "question";
			if ($scope.isComment) {
				$scope.page.subpageIds = $scope.page.commentIds;
				$scope.page.subpageIds.sort(pageService.getChildSortFunc("oldestFirst"));
			} else if ($scope.isQuestion) {
				$scope.page.subpageIds = $scope.page.answerIds;
				$scope.page.subpageIds.sort(pageService.getChildSortFunc("likes"));
			}
			$scope.isCollapsed = false;

			if ($scope.isComment) {
				$scope.myUrl = pageService.getPageUrl($scope.lensId) + "#subpage-" + $scope.page.pageId;
			} else {
				$scope.myUrl = pageService.getPageUrl($scope.page.pageId);
			}

			// Check if this comment is selected via URL hash
			$scope.isSelected = function() {
				return $location.hash() === "subpage-" + $scope.page.pageId;
			};

			// Called when the user collapses/expands this subpage
			$scope.collapseToggle = function() {
				$scope.isCollapsed = !$scope.isCollapsed;
			};

			// Called when the user wants to edit the subpage
			$scope.editSubpage = function(event) {
				if (!event.ctrlKey) {
					pageService.loadEdit({
						pageAlias: $scope.page.pageId,
						success: function() {
							$scope.editing = true;
						},
					});
					event.preventDefault();
				}
			};

			// Called when the user is done editing the subpage
			$scope.editDone = function(result) {
				$scope.editing = false;
				if (!result.discard) {
					pageService.newCommentCreated(result.pageId);
				}
			};

			// Called when the user wants to delete the subpage
			$scope.deleteSubpage = function() {
				pageService.deletePage($scope.page.pageId, function() {
					$scope.isDeleted = true;
					// TODO: reenable toast when we fix the bug with its positioning
					/*$mdToast.show(
						$mdToast.simple()
						.textContent("Comment deleted")
						.position("top right")
						.hideDelay(3000)
					);*/
				}, function(data) {
					$scope.addMessage("delete", "Error deleting page: " + data, "error");
				});
			};

			// Called to create a new reply
			$scope.newReply = function() {
				pageService.newComment({
					parentPageId: $scope.lensId,
					replyToId: $scope.page.pageId,
					success: function(newCommentId) {
						$scope.newReplyId = newCommentId;
					},
				});
			};

			// Called when the user is done with the new reply
			$scope.newReplyDone = function(result) {
				$scope.newReplyId = undefined;
				if (!result.discard) {
					pageService.newCommentCreated(result.pageId);
				}
			};
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});

// Directive for container holding an inline comment
app.directive("arbInlineComment", function ($compile, $timeout, $location, $mdToast, pageService, userService, autocompleteService, RecursionHelper) {
	return {
		templateUrl: "static/html/inlineComment.html",
		scope: {
			commentId: "@",
			lensId: "@",  // id of the lens this comment belongs to
		},
		controller: function ($scope) {
			$scope.isExpanded = false;
			$scope.toggleExpand = function() {
				$scope.isExpanded = !$scope.isExpanded;
			};
		},
		link: function(scope, element, attrs) {
			var content = element.find(".inline-subpage");
			scope.showExpandButton = function() {
				return content.get(0).scrollHeight > content.height() && !scope.isExpanded;
			};
		},
	};
});
