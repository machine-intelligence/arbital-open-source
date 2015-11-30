"use strict";

// $parentDiv - div that will contain the editPage element
// $commentButton - button/link the user clicked to create this edit comment div
// scope - scope which will be used to store doneFn and for compiling elements
// options: {
//   primaryPageId - page that will own this comment
//   divType - the type of the new div, either "comment" or "question"
//   parentCommentId - optionally, id of the comment this is a reply to
//   anchorContext,anchorText,anchorOffset - optionally, set for inline comments
//   callback - callback to call if edit is abandoned
// }
var createEditSubpageDiv = function($parentDiv, $commentButton, scope, options) {
	// Create and show the edit page directive.
	var createEditPage = function(newPageId) {
		// Callback for processing when the user is done creating a new comment.
		var doneFnName = "doneFn" + newPageId;
		scope[doneFnName] = function(result) {
			if (result.abandon) {
				toggleVisibility(true, false);
				$parentDiv.find("arb-edit-page").remove();
				if (options.callback) options.callback();
			} else if (result.alias) {
				smartPageReload("subpage-" + result.alias);
			}
		};
		var el = scope.$compile("<arb-edit-page page-id='" + newPageId +
				"' primary-page-id='" + options.primaryPageId +
				"' done-fn='" + doneFnName + "(result)'></arb-edit-page>")(scope);
		$parentDiv.append(el);
	};

	// Toggle the visibility of involved elements.
	// showButton - true if we should show the new/edit comment button/link
	// showLoading - true if we should show the loading spinner
	var toggleVisibility = function(showButton, showLoading) {
		$commentButton.toggle(showButton);
		$parentDiv.find(".loading-indicator").toggle(showLoading);
		$parentDiv.find("arb-edit-page").toggle(!showButton && !showLoading);
		return false;
	};

	if ($parentDiv.find("arb-edit-page").length > 0) {
		toggleVisibility(false, false);
	} else {
		toggleVisibility(false, true);
		scope.pageService.getNewPage({
			success: function(newPageId) {
				toggleVisibility(false, false);
				var page = scope.pageService.editMap[newPageId];
				page.type = options.divType;
				page.editGroupId = scope.userService.user.id;
				page.parentIds = [options.primaryPageId];
				if (options.parentCommentId) {
					page.parentIds.push(options.parentCommentId);
				}
				if (options.anchorContext) {
					page.anchorContext = options.anchorContext;
					page.anchorText = options.anchorText;
					page.anchorOffset = options.anchorOffset;
				}
				createEditPage(newPageId);
			},
		});
	}
};

// Directive for showing a subpage.
app.directive("arbSubpage", function ($compile, $timeout, $location, pageService, userService, autocompleteService, RecursionHelper) {
	return {
		templateUrl: "/static/html/subpage.html",
		scope: {
			pageId: "@",  // id of this subpage
			lensId: "@",  // id of the lens this subpage belongs to
			parentSubpageId: "@",  // id of the parent subpage, if there is one
		},
		controller: function ($scope, pageService, userService) {
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
			$scope.isCollapsed = $scope.isQuestion;

			if ($scope.isComment) {
				$scope.myUrl = pageService.getPageUrl($scope.lensId) + "#subpage-" + $scope.page.pageId;
			} else {
				$scope.myUrl = pageService.getPageUrl($scope.page.pageId);
			}

			// Called when the user collapses/expands this subpage
			$scope.collapseToggle = function() {
				$scope.isCollapsed = !$scope.isCollapsed;
			};
		},
		link: function(scope, element, attrs) {
			return;
			var $replies = element.find(".replies");
		
			// Highlight the comment div. Used for selecting comments when #anchor matches.
			var highlightCommentDiv = function() {
				$(".hash-anchor").removeClass("hash-anchor");
				element.find(".comment-content").addClass("hash-anchor");
			};
			if (window.location.hash === "#subpage-" + scope.pageId) {
				highlightCommentDiv();
			}
		
			// Comment editing stuff.
			var $comment = element.find(".comment-content");
			// Create and show the edit page directive.
			var createEditPage = function() {
				var el = $compile("<arb-edit-page page-id='" + scope.pageId +
						"' primary-page-id='" + scope.primaryPageId +
						"' done-fn='doneFn(result)'></arb-edit-page>")(scope);
				$comment.append(el);
			};
			var destroyEditPage = function() {
				$comment.find("arb-edit-page").remove();
			};
			// Reload comment from the server, loading the last, potentially non-live edit.
			var reloadComment = function() {
				$comment.find(".loading-indicator").show();
				pageService.removePageFromMap(scope.pageId);
				pageService.loadEdit({
					pageAlias: scope.pageId,
					success: function() {
						$comment.find(".loading-indicator").hide();
						createEditPage(scope.pageId);
					},
				});
			}
			// Show/hide the comment vs the edit page.
			function toggleEditComment(visible) {
				$comment.find(".comment-body").toggle(!visible);
				$comment.find("arb-edit-page").toggle(visible);
			}
			// Callback used when the user is done editing the comment.
			scope.doneFn = function(result) {
				if (result.abandon) {
					toggleEditComment(false);
					element.find(".edit-comment-link").removeClass("has-draft");
					scope.comment.hasDraft = false;
					destroyEditPage();
				} else if (result.alias) {
					smartPageReload("subpage-" + result.alias);
				}
			};
			element.find(".permalink-comment-link").on("click", function(event) {
				smartPageReload("subpage-" + scope.comment.pageId);
				return false;
			});
		
			// Editing a comment
			element.find(".edit-comment-link").on("click", function(event) {
				$(".hash-anchor").removeClass("hash-anchor");
				// Dynamically create arb-edit-page directive if it doesn't exist already.
				if ($comment.find("arb-edit-page").length <= 0) {
					if (scope.comment.hasDraft) {
						// Load the draft.
						reloadComment();
					} else {
						createEditPage(scope.comment.pageId);
					}
				}
				toggleEditComment(true);
				return false;
			});
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});

// Directive for creating a new comment.
app.directive("arbNewComment", function ($compile, pageService, userService) {
	return {
		templateUrl: "/static/html/newComment.html",
		controller: function ($scope, pageService, userService) {
		},
		scope: {
			primaryPageId: "@",  // page which this comment is ultimately attached to (i.e. primary page)
			parentCommentId: "@",  // optional id of the immediate parent comment
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.$compile = $compile;
			var $newComment = element.find(".new-comment");
			element.find(".new-comment-link").on("click", function(event) {
				if (userService.user.id === "0") {
					showSignupPopover($(event.currentTarget));
					return false;
				}
				$(".hash-anchor").removeClass("hash-anchor");
				createEditSubpageDiv($newComment, $newComment.find(".new-comment-link"), scope, {
					primaryPageId: scope.primaryPageId,
					divType: "comment",
					parentCommentId: scope.parentCommentId,
				});
				return false;
			});
		},
	};
});
