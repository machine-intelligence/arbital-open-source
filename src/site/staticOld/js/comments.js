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
				page.parents = [{parentId: options.primaryPageId, childId: newPageId}];
				if (options.parentCommentId) {
					page.parents.push({parentId: options.parentCommentId, childId: newPageId});
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

var commentLinkFunc = function(scope, element, attrs, $compile, $timeout, pageService, userService, autocompleteService) {
	var $replies = element.find(".replies");
	// Dynamically create reply elements.
	if (scope.parentCommentId === undefined) {
		if (scope.comment.commentIds != null) {
			scope.comment.commentIds.sort(pageService.getChildSortFunc("oldestFirst"));
			for (var n = 0; n < scope.comment.commentIds.length; n++) {
				var childId = scope.comment.commentIds[n];
				if (pageService.pageMap[childId].type !== "comment") continue;
				var $comment = $compile("<arb-comment primary-page-id='" + scope.primaryPageId +
						"' page-id='" + childId +
						"' parent-comment-id='" + scope.pageId + "'></arb-comment>")(scope);
				$replies.append($comment);
			}
		}
		// Add New Comment element.
		var $newComment = $compile("<arb-new-comment primary-page-id='" + scope.primaryPageId +
				"' parent-comment-id='" + scope.pageId + "'></arb-new-comment>")(scope);
		$replies.append($newComment);
	}

	$timeout(function() {
		// Process comment's text using Markdown.
		arbMarkdown.init(false, scope.pageId, scope.comment.text, element, pageService, userService);
	});

	// Highlight the comment div. Used for selecting comments when #anchor matches.
	var highlightCommentDiv = function() {
		$(".hash-anchor").removeClass("hash-anchor");
		element.find(".comment-content").addClass("hash-anchor");
	};
	if (window.location.hash === "#subpage-" + scope.pageId) {
		highlightCommentDiv();
	}

	// Comment voting stuff.
	// likeClick is 1 is user clicked like and 0 if they clicked reset like.
	element.find(".like-comment-link").on("click", function(event) {
		var $target = $(event.target);
		var $commentRow = $target.closest(".comment-row");
		var $likeCount = $commentRow.find(".comment-like-count");
	
		// Update UI.
		$target.toggleClass("on");
		var newLikeValue = $target.hasClass("on") ? 1 : 0;
		var totalLikes = ((+$likeCount.text()) + (newLikeValue > 0 ? 1 : -1));
		if (totalLikes > 0) {
			$likeCount.text("" + totalLikes);
		} else {
			$likeCount.text("");
		}
		
		// Notify the server
		var data = {
			pageId: scope.pageId,
			value: newLikeValue,
		};
		$.ajax({
			type: "POST",
			url: "/newLike/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});

	// Process comment subscribe click.
	element.find(".subscribe-comment-link").on("click", function(event) {
		pageService.subscribeTo($target);
		return false;
	});
	
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

	// Callback used when the user confirms deleting a comment.
	scope.confirmDeleteCommentFn = function(returnedPageId) {
		pageService.deletePage(returnedPageId, (function(data, status) {
			smartPageReload();
		}));
		return false;
	};
	// Deleting a comment
	element.find(".delete-comment-link").on("click", function(event) {

		var $replies = element.find(".replies");
		var $deleteButton = element.find(".delete-comment-link");
		var position = $deleteButton.position();
		var $confirmPopover = $compile("<arb-confirm-popover message='Are you sure you want to delete this comment?" +
																 "' page-id='" + scope.comment.pageId +
																 "' x-pos='" + position.left  +
																 "' y-pos='" + position.top  +
																 "' confirm-fn='confirmDeleteCommentFn(returnedPageId)" + 
																 "'></arb-confirm-popover>")(scope);
		$replies.append($confirmPopover);

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
}

// Directive for showing a comment.
app.directive("arbComment", function ($compile, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/comment.html",
		controller: function ($scope, pageService, userService) {
			$scope.userService = userService;
			$scope.comment = pageService.pageMap[$scope.pageId];
		},
		scope: {
			primaryPageId: "@",  // id of the primary page this comment belongs to
			pageId: "@",  // id of this comment
			parentCommentId: "@",  // id of the parent comment, if there is one
		},
		link: function(scope, element, attrs) {
	    commentLinkFunc(scope, element, attrs, $compile, $timeout, pageService, userService, autocompleteService);
		},
	};
});

// Directive for showing an inline question.
app.directive("arbQuestion", function ($compile, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/inlineQuestion.html",
		controller: function ($scope, pageService, userService) {
			$scope.userService = userService;
			$scope.comment = pageService.pageMap[$scope.pageId];
		},
		scope: {
			primaryPageId: "@",  // id of the primary page this comment belongs to
			pageId: "@",  // id of this comment
		},
		link: function(scope, element, attrs) {
	    commentLinkFunc(scope, element, attrs, $compile, $timeout, pageService, userService, autocompleteService);
		},
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
