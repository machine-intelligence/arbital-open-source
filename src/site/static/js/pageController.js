"use strict";

// Create new PageJsController.
// page - page object corresponding to the page being displayed.
var PageJsController = function(page, $topParent, pageService, userService) {
	var page = page;
	var $topParent = $topParent;
	var pageId = page.pageId; // id of the page being displayed
	var userId = userService.user.id;
	
	// Highlight the page div. Used for selecting answers when #anchor matches.
	var highlightPageDiv = function() {
		$(".hash-anchor").removeClass("hash-anchor");
		$topParent.find(".page-body-div").addClass("hash-anchor");
	};
	if (window.location.hash === "#page-" + pageId) {
		highlightPageDiv();
	}
	
	// === Setup handlers.
	
	// Inline comments
	// Create the inline comment highlight spans for the given paragraph.
	this.createInlineCommentHighlight = function(paragraphNode, start, end, nodeClass) {
		// How many characters we passed in the anchor context (which has escaped characters).
		var charCount = 0;
		// How many characters we passed in the actual paragraph.
		var pCharCount = 0;
		// Store ranges we want to highlight.
		var ranges = [];
		// Compute context and text.
		recursivelyVisitChildren(paragraphNode, function(node, nodeText, needsEscaping) {
			if (nodeText === null) return false;
			var escapedText = needsEscaping ? escapeMarkdownChars(nodeText) : nodeText;
			var nodeWholeTextLength = node.wholeText ? node.wholeText.length : 0;
			var range = document.createRange();
			var nextCharCount = charCount + escapedText.length;
			var pNextCharCount = pCharCount + nodeWholeTextLength; //or nodeText.length???
			if (charCount <= start && nextCharCount >= end) {
				var pStart = unescapeMarkdownChars(escapedText.substring(0, start - charCount)).length;
				var pEnd = unescapeMarkdownChars(escapedText.substring(0, end - charCount)).length;
				range.setStart(node, pStart);
				range.setEnd(node, pEnd);
				ranges.push(range);
			} else if (charCount <= start && nextCharCount > start) {
				var pStart = unescapeMarkdownChars(escapedText.substring(0, start - charCount)).length;
				range.setStart(node, pStart);
				range.setEnd(node, Math.min(nodeWholeTextLength, nodeText.length));
				ranges.push(range);
			} else if (start < charCount && nextCharCount >= end) {
				range.setStart(node, 0);
				range.setEnd(node, Math.min(nodeWholeTextLength, end - charCount));
				ranges.push(range);
			} else if (start < charCount && nextCharCount > start) {
				if (nodeWholeTextLength > 0) {
					range.setStart(node, 0);
					range.setEnd(node, Math.min(nodeWholeTextLength, nodeText.length));
				} else {
					range.selectNodeContents(node);
				}
				ranges.push(range);
			} else if (start == charCount && charCount == nextCharCount) {
				// Rare occurence, but this captures MathJax divs/spans that
				// precede the script node where we actually get the text from.
				range.selectNodeContents(node);
				ranges.push(range);
			}
			charCount = nextCharCount;
			pCharCount = pNextCharCount;
			return charCount >= end;
		});
		// Highlight ranges after we did DOM traversal, so that there are no
		// modifications during the traversal.
		for (var i = 0; i < ranges.length; i++) {
			highlightRange(ranges[i], nodeClass);
		}
		return ranges.length > 0 ? ranges[0].startContainer : null;
	};

	var $newInlineCommentDiv = $(".new-inline-comment-div");
	var $markdownText = $topParent.find(".markdown-text");
	$markdownText.on("mouseup", function(event) {
		// Do setTimeout, because otherwise there is a bug when you double click to
		// select a word/paragraph, then click again and the selection var is still
		// the same (not cleared).
		window.setTimeout(function(){
			var show = !!processSelectedParagraphText();
			$newInlineCommentDiv.toggle(show);
			if (show) {
				pageView.setNewInlineCommentPrimaryPageId(pageId);
			}
		}, 0);
	});

	// Deleting a page
	$topParent.find(".delete-page-link").on("click", function(event) {
		$("#delete-page-alert").show();
		return false;
	});
	$topParent.find(".delete-page-cancel").on("click", function(event) {
		$("#delete-page-alert").hide();
	});
	$topParent.find(".delete-page-confirm").on("click", function(event) {
		var data = {
			pageId: pageId,
		};
		$.ajax({
			type: "POST",
			url: "/deletePage/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
			smartPageReload();
		});
		return false;
	});
	
	// Page (dis)liking stuff.
	// likeClick is 1 is user clicked like and -1 if they clicked dislike.
	var processLike = function(likeClick, event) {
		if (userService.user.id === "0") {
			showSignupPopover($(event.currentTarget));
			return false;
		}

		var $target = $(event.target);
		var $like = $target.closest(".page-like-div");
		var $likeCount = $like.find(".like-count");
		var $dislikeCount = $like.find(".dislike-count");
		var currentLikeValue = +$like.attr("my-like");
		var newLikeValue = (likeClick === currentLikeValue) ? 0 : likeClick;
		var likes = +$likeCount.text();
		var dislikes = +$dislikeCount.text();
	
		// Update like counts.
		// This logic has been created by looking at truth tables.
		if (currentLikeValue === 1) {
			likes -= 1;
		} else if (likeClick === 1) {
			likes += 1;
		}
		if (currentLikeValue === -1) {
			dislikes -= 1;
		} else if (likeClick === -1) {
			dislikes += 1;
		}
		$likeCount.text("" + likes);
		$dislikeCount.text("" + dislikes);
	
		// Update my-like
		$like.attr("my-like", "" + newLikeValue);
	
		// Display my like
		$like.find(".like-link").toggleClass("on", newLikeValue === 1);
		$like.find(".dislike-link").toggleClass("on", newLikeValue === -1);
		
		// Notify the server
		var data = {
			pageId: pageId,
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
	}
	$topParent.find(".like-link").on("click", function(event) {
		return processLike(1, event);
	});
	$topParent.find(".dislike-link").on("click", function(event) {
		return processLike(-1, event);
	});
	
	// Subscription stuff.
	$topParent.find(".subscribe-to-page-link").on("click", function(event) {
		if (userService.user.id === "0") {
			showSignupPopover($(event.currentTarget));
			return false;
		}

		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			pageId: pageId,
		};
		$.ajax({
			type: "POST",
			url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});

	// Process click on showing the page diff button.
	$topParent.on("click", ".show-page-diff", function(event) {
		var $pageText = $topParent.find(".page-text");
		var $editDiff = $pageText.siblings(".edit-diff");
		if ($editDiff.is(":visible")) {
			// Show the page.
			$pageText.show();
			$editDiff.hide();
		} else if ($editDiff.length > 0) {
			// We already loaded the edit from the server, just show it.
			$pageText.hide();
			$editDiff.show();
		} else {
			// Load the edit from the server.
			pageService.loadEdit({
				pageId: pageId,
				createdAtLimit: $("body").attr("last-visit"),
				success: function(data, status) {
					var dmp = new diff_match_patch();
					var diffs = dmp.diff_main(data[pageId].text, page.text);
					dmp.diff_cleanupSemantic(diffs);
					var html = dmp.diff_prettyHtml(diffs);
					$pageText.hide().after($("<div class='edit-diff'>" + html + "</div>"));
				},
			});
		}
	});

	// Start initializes things that have to be killed when this editPage stops existing.
	this.start = function($compile, scope) {
		// for question pages, check if we need to add the anchor text
//console.log("testing...");
		if (page.type === "question") {
//console.log("if (page.type === \"question\") {");
//console.log("page: %s", page);
//console.log("page.anchorContext: %s", page.anchorContext);
//console.log("page.anchorText: %s", page.anchorText);
//console.log("page.text: %s", page.text);

			if (page.anchorContext && page.anchorText) {
				page.text = "> " + page.anchorText + "\n\n" + page.text;
			}
		}

		// Set up markdown.
		arbMarkdown.init(false, pageId, page.text, $topParent, pageService);

		// Setup probability vote slider.
		// NOTE: this pretty messy, since there are some race conditions here we are
		// trying to mitigate.
		if (page.hasVote) {
			// Timeout to give use a chance to switch to correct lens tab.
			var $lensTab = $("[data-target='#lens-" + pageId + "']");
			var doCreateVoteSlider = function() {
				// Timeout to wait until the tab pane is visible.
				window.setTimeout(function() {
					// If the pane is now not visible, then don't do anything.
					if (!$topParent.closest(".tab-pane").is(":visible")) return;
					$lensTab.off("click", doCreateVoteSlider);
					createVoteSlider($topParent.find(".page-vote"), userService, page, false);
				});
			};
			$lensTab.on("click", doCreateVoteSlider);
			// Check to see if the tab pane is visible.
			if ($topParent.closest(".tab-pane").is(":visible")) {
				doCreateVoteSlider();
			}
		}

		// Process all embedded votes.
		window.setTimeout(function() {
			$topParent.find("[embed-vote-id]").each(function(index) {
				var $link = $(this);
				var pageAlias = $link.attr("embed-vote-id");
				pageService.loadPages([pageAlias], {
					includeAuxData: true,
					loadVotes: true,
					overwrite: true,
					success: function(data, status) {
						var pageId = Object.keys(data)[0];
						var divId = "embed-vote-" + pageId;
						var $embedDiv = $compile("<div id='" + divId + "' class='embedded-vote'><arb-likes-page-title page-id='" + pageId + "'></arb-likes-page-title></div>")(scope);
						$link.replaceWith($embedDiv);
						createVoteSlider($("#" + divId), userService, pageService.pageMap[pageId], false);
					},
					error: function(data, status) {
						console.log("Couldn't load embedded votes: " + pageAlias);
					}
				});
			});
		});
	};

	// Called before this controller is destroyed.
	this.stop = function() {
	};
};

// Directive for showing a standard Arbital page.
app.directive("arbPage", function (pageService, userService, $compile, $timeout) {
	return {
		templateUrl: "/static/html/page.html",
		controller: function ($scope, pageService, userService) {
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.questionIds = [];
			for (var n = 0; n < $scope.page.children.length; n++) {
				var id = $scope.page.children[n].childId;
				var page = pageService.pageMap[id];
				if (page.type === "question" && !page.anchorContext) {
					$scope.questionIds.push(id);
				}
			}

			// Sort question ids by likes, but put the ones created by current user first.
			$scope.questionIds.sort(function(id1, id2) {
				var page1 = pageService.pageMap[id1];
				var page2 = pageService.pageMap[id2];
				var ownerDiff = (page2.creatorId == userService.user.id ? 1 : 0) -
						(page1.creatorId == userService.user.id ? 1 : 0);
				if (ownerDiff != 0) return ownerDiff;
				return page2.likeScore - page1.likeScore;
			});
		},
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {

			// Set up Page JS Controller.
			$timeout(function(){
				scope.pageJsController = new PageJsController(scope.page, element, pageService, userService);
				scope.pageJsController.start($compile, scope);

				if (scope.page.commentIds != null) {
					// Process comments in two passes. First normal comments.
					processComments(false);
					$timeout(function() {
						// Inline comments after a delay long enough for MathJax to have been processed.
						processComments(true);
					}, 3000);
				}

				if (scope.page.questionIds != null) {
					// Process questions in two passes. First normal questions.
					processQuestions(false);
					$timeout(function() {
						// Inline questions after a delay long enough for MathJax to have been processed.
						processQuestions(true);
					}, 3000);
				}
			});

			// Track toggle-inline-comment offsets, so we can prevent overlap.
			var inlineCommentOffsets = [];
			var fixInlineCommentOffset = function(offset) {
				for (var i = 0; i < inlineCommentOffsets.length; i++) {
					var o = inlineCommentOffsets[i];
					if (Math.abs(offset.top - o.top) < 25) {
						if (Math.abs(offset.left - o.left) < 30) {
							offset.left = o.left + 35;
						}
					}
				}
				inlineCommentOffsets.push(offset);
			}

			// Create a toggle-inline-comment-div.
			var createNewInlineCommentToggle = function(pageId, paragraphNode, anchorOffset, anchorLength) {
				var created = false;
				var doCreate = function() {
					created = true;
					var highlightClass = "inline-comment-" + pageId;
					var $commentDiv = $(".toggle-inline-comment-div.template").clone();
					$commentDiv.attr("id", "comment-" + pageId).removeClass("template");
					var comment = pageService.pageMap[pageId];
					var commentCount = comment.children.length + 1;
					$commentDiv.find(".inline-comment-count").text("" + commentCount);
					$(".question-div").append($commentDiv);
	
					// Process mouse events.
					var $commentIcon = $commentDiv.find(".inline-comment-icon");
					$commentIcon.on("mouseenter", function(event) {
						$("." + highlightClass).addClass("inline-comment-highlight");
					});
					$commentIcon.on("mouseleave", function(event) {
						if ($commentIcon.hasClass("on")) return true;
						$("." + highlightClass).removeClass("inline-comment-highlight");
					});
					$commentIcon.on("click", function(event) {
						pageView.toggleInlineComment($commentDiv, function() {
							$("." + highlightClass).addClass("inline-comment-highlight");
							var $comment = $compile("<arb-comment primary-page-id='" + scope.page.pageId +
									"' page-id='" + pageId + "'></arb-comment>")(scope);
							$(".inline-comment-div").append($comment);
						});
						return false;
					});
	
					var commentIconLeft = $(".question-div").offset().left + 10;
					var anchorNode = scope.pageJsController.createInlineCommentHighlight(paragraphNode, anchorOffset, anchorOffset + anchorLength, highlightClass);
					if (anchorNode) {
						if (anchorNode.nodeType != Node.ELEMENT_NODE) {
							anchorNode = anchorNode.parentElement;
						}
						var offset = {left: commentIconLeft, top: $(anchorNode).offset().top};
						fixInlineCommentOffset(offset);
						$commentDiv.offset(offset);
	
						// Check if we need to expand this inline comment because of the URL anchor.
						var expandComment = window.location.hash === "#comment-" + pageId;
						if (!expandComment) {
							// Check if one of the children is selected.
							for (var n = 0; n < comment.children.length; n++) {
								expandComment |= window.location.hash === "#comment-" + comment.children[n].childId;
//console.trace();
//console.log("testing...");
//console.log("expandComment: %s", expandComment);
//console.log("window.location.hash: %s", window.location.hash);
//console.log("pageId: %s", pageId);
							}
						}
						if (expandComment) {
							// Delay to allow other inline comment buttons to compute their position correctly.
							window.setTimeout(function() {
								$commentIcon.trigger("click");
								$("html, body").animate({
				      	  scrollTop: $(anchorNode).offset().top - 100
					    	}, 1000);
							}, 100);
						}
					} else {
						$commentDiv.hide();
						console.log("ERROR: couldn't find anchor node for inline comment");
					}
				};
				// Check that we don't have another lens selected, in which case we'll
				// postpone creating the div.
				if (pageService.primaryPage === scope.page) {
					doCreate();
				}
				pageService.primaryPageCallbacks.push(function() {
					if (created) {
						$("#comment-" + pageId).toggle(pageService.primaryPage === scope.page);
					} else if (pageService.primaryPage === scope.page) {
						window.setTimeout(function() {  // wait until the page shows
							doCreate();
						});
					}
				});
			}

			// Create a toggle-inline-comment-div, for a question.
			var createNewInlineQuestionToggle = function(pageId, paragraphNode, anchorOffset, anchorLength) {
				var created = false;
				var doCreate = function() {
					created = true;
					var highlightClass = "inline-comment-" + pageId;
					var $commentDiv = $(".toggle-inline-comment-div.template").clone();
					$commentDiv.attr("id", "comment-" + pageId).removeClass("template");
					var comment = pageService.pageMap[pageId];
					var commentCount = comment.children.length + 1;
					$commentDiv.find(".inline-comment-count").text("?");
					$(".question-div").append($commentDiv);
	
					// Process mouse events.
					var $commentIcon = $commentDiv.find(".inline-comment-icon");
					$commentIcon.on("mouseenter", function(event) {
						$("." + highlightClass).addClass("inline-comment-highlight");
					});
					$commentIcon.on("mouseleave", function(event) {
						if ($commentIcon.hasClass("on")) return true;
						$("." + highlightClass).removeClass("inline-comment-highlight");
					});
					$commentIcon.on("click", function(event) {
						pageView.toggleInlineQuestion($commentDiv, function() {
							$("." + highlightClass).addClass("inline-comment-highlight");
							var $comment = $compile("<arb-question primary-page-id='" + scope.page.pageId +
									"' page-id='" + pageId + "'></arb-question>")(scope);
							$(".inline-comment-div").append($comment);
						});
						return false;
					});
	
					var commentIconLeft = $(".question-div").offset().left + 10;
					var anchorNode = scope.pageJsController.createInlineCommentHighlight(paragraphNode, anchorOffset, anchorOffset + anchorLength, highlightClass);
					if (anchorNode) {
						if (anchorNode.nodeType != Node.ELEMENT_NODE) {
							anchorNode = anchorNode.parentElement;
						}
						var offset = {left: commentIconLeft, top: $(anchorNode).offset().top};
						fixInlineCommentOffset(offset);
						$commentDiv.offset(offset);
	
						// Check if we need to expand this inline question because of the URL anchor.
						var expandComment = window.location.hash === "#comment-" + pageId;
//console.trace();
//console.log("testing...");
//console.log("expandComment: %s", expandComment);
//console.log("window.location.hash: %s", window.location.hash);
//console.log("pageId: %s", pageId);
						if (!expandComment) {
							// Check if one of the children is selected.
							for (var n = 0; n < comment.children.length; n++) {
								expandComment |= window.location.hash === "#comment-" + comment.children[n].childId;
							}
						}
						if (expandComment) {
							// Delay to allow other inline question buttons to compute their position correctly.
							window.setTimeout(function() {
								$commentIcon.trigger("click");
								$("html, body").animate({
				      	  scrollTop: $(anchorNode).offset().top - 100
					    	}, 1000);
							}, 100);
						}
					} else {
						$commentDiv.hide();
						console.log("ERROR: couldn't find anchor node for inline question");
					}
				};
				// Check that we don't have another lens selected, in which case we'll
				// postpone creating the div.
				if (pageService.primaryPage === scope.page) {
					doCreate();
				}
				pageService.primaryPageCallbacks.push(function() {
					if (created) {
						$("#comment-" + pageId).toggle(pageService.primaryPage === scope.page);
					} else if (pageService.primaryPage === scope.page) {
						window.setTimeout(function() {  // wait until the page shows
							doCreate();
						});
					}
				});
			}

			// Dynamically create comment elements.
			var processComments = function(allowInline) {
				var $comments = element.find(".comments");
				var $markdown = element.find(".markdown-text");
				var dmp = new diff_match_patch();
				dmp.Match_MaxBits = 10000;
				dmp.Match_Distance = 10000;

				// If we have inline comments, we'll need to compute the raw text for
				// each paragraph.
				var paragraphTexts = undefined;
				var populateParagraphTexts = function() {
					paragraphTexts = [];
					var i = 0;
					$markdown.children().each(function() {
						paragraphTexts.push(getParagraphText($(this).get(0)).context);
						i++;
					});
				};

				// Go through comments in chronological order.
				scope.page.commentIds.sort(pageService.getChildSortFunc({sortChildrenBy: "chronological", type: "comment"}));
				for (var n = 0; n < scope.page.commentIds.length; n++) {
					var comment = pageService.pageMap[scope.page.commentIds[n]];
					// Check if the comment in anchored and we can still find the paragraph.
					if (comment.anchorContext && comment.anchorText) {
						if (!allowInline) continue;
						// Find the best paragraph.
						var bestParagraphNode, bestParagraphText, bestScore = Number.MAX_SAFE_INTEGER;
						if (!paragraphTexts) {
							populateParagraphTexts();
						}
						for (var i = 0; i < paragraphTexts.length; i++) {
							var text = paragraphTexts[i];
							var diffs = dmp.diff_main(text, comment.anchorContext);
							var score = dmp.diff_levenshtein(diffs);
							if (score < bestScore) {
								bestParagraphNode = $markdown.children().get(i);
								bestParagraphText = text;
								bestScore = score;
							}
						}
						if (bestScore > comment.anchorContext.length / 2) {
							// This is not a good paragraph match. Continue processing as a normal comment.
							comment.text = "> " + comment.anchorText + "\n\n" + comment.text;
						} else {
							// Find offset into the best paragraph.
							var anchorLength;
							var anchorOffset = dmp.match_main(bestParagraphText, comment.anchorText, comment.anchorOffset);
							if (anchorOffset < 0) {
								// Couldn't find a match within the paragraph. We'll just use paragraph as the anchor.
								anchorOffset = 0;
								anchorLength = bestParagraphText.length;
							} else {
								// Figure out how long the highlighted anchor should be.
								var remainingText = bestParagraphText.substring(anchorOffset);
								var diffs = dmp.diff_main(remainingText, comment.anchorText);
								anchorLength = remainingText.length;
								if (diffs.length > 0) {
									// Note: we can potentially be more clever here and discount
									// edits done after anchorText.length chars higher.
									var lastDiff = diffs[diffs.length - 1];
									if (lastDiff[0] < 0) {
										anchorLength -= lastDiff[1].length;
									}
								}
							}
							createNewInlineCommentToggle(comment.pageId, bestParagraphNode, anchorOffset, anchorLength);
							continue;
						}
					} else if (allowInline) {
						continue;
					}
					// Make sure this comment is not a reply (i.e. it has a parent comment)
					// If it's a reply, add it as a child to the corresponding parent comment.
					if (comment.parents != null) {
						var hasParentComment = false;
						for (var i = 0; i < comment.parents.length; i++) {
							var parent = pageService.pageMap[comment.parents[i].parentId];
							hasParentComment = parent.type === "comment";
							if (hasParentComment) {
								if (parent.children == null) parent.children = [];
								parent.children.push({parentId: parent.pageId, childId: comment.pageId});
								break;
							}
						}
						if (hasParentComment) continue;
					}
					var $comment = $compile("<arb-comment primary-page-id='" + scope.pageId +
							"' page-id='" + comment.pageId + "'></arb-comment>")(scope);
					$comments.prepend($comment);
				}
			};

			// Dynamically create question elements.
			var processQuestions = function(allowInline) {
				var $questions = element.find(".questions");
				var $markdown = element.find(".markdown-text");
				var dmp = new diff_match_patch();
				dmp.Match_MaxBits = 10000;
				dmp.Match_Distance = 10000;

				// If we have inline questions, we'll need to compute the raw text for
				// each paragraph.
				var paragraphTexts = undefined;
				var populateParagraphTexts = function() {
					paragraphTexts = [];
					var i = 0;
					$markdown.children().each(function() {
						paragraphTexts.push(getParagraphText($(this).get(0)).context);
						i++;
					});
				};
				// Go through questions in chronological order.
				scope.page.questionIds.sort(pageService.getChildSortFunc({sortChildrenBy: "chronological", type: "question"}));
				for (var n = 0; n < scope.page.questionIds.length; n++) {
					var question = pageService.pageMap[scope.page.questionIds[n]];
					// Check if the question in anchored and we can still find the paragraph.
					if (question.anchorContext && question.anchorText) {
						if (!allowInline) continue;
						// Find the best paragraph.
						var bestParagraphNode, bestParagraphText, bestScore = Number.MAX_SAFE_INTEGER;
						if (!paragraphTexts) {
							populateParagraphTexts();
						}
						for (var i = 0; i < paragraphTexts.length; i++) {
							var text = paragraphTexts[i];
							var diffs = dmp.diff_main(text, question.anchorContext);
							var score = dmp.diff_levenshtein(diffs);
							if (score < bestScore) {
								bestParagraphNode = $markdown.children().get(i);
								bestParagraphText = text;
								bestScore = score;
							}
						}
						if (bestScore > question.anchorContext.length / 2) {
							// This is not a good paragraph match. Continue processing as a normal question.
							question.text = "> " + question.anchorText + "\n\n" + question.text;
							//question.anchorContext = null;
							scope.questionIds.push(question.pageId);
							continue;
						} else {
							// Find offset into the best paragraph.
							var anchorLength;
							var anchorOffset = dmp.match_main(bestParagraphText, question.anchorText, question.anchorOffset);
							if (anchorOffset < 0) {
								// Couldn't find a match within the paragraph. We'll just use paragraph as the anchor.
								anchorOffset = 0;
								anchorLength = bestParagraphText.length;
							} else {
								// Figure out how long the highlighted anchor should be.
								var remainingText = bestParagraphText.substring(anchorOffset);
								var diffs = dmp.diff_main(remainingText, question.anchorText);
								anchorLength = remainingText.length;
								if (diffs.length > 0) {
									// Note: we can potentially be more clever here and discount
									// edits done after anchorText.length chars higher.
									var lastDiff = diffs[diffs.length - 1];
									if (lastDiff[0] < 0) {
										anchorLength -= lastDiff[1].length;
									}
								}
							}
							createNewInlineQuestionToggle(question.pageId, bestParagraphNode, anchorOffset, anchorLength);
							continue;
						}
					} else if (allowInline) {
						continue;
					}
					// Make sure this question is not a reply (i.e. it has a parent question)
					// If it's a reply, add it as a child to the corresponding parent question.
					if (question.parents != null) {
						var hasParentQuestion = false;
						for (var i = 0; i < question.parents.length; i++) {
							var parent = pageService.pageMap[question.parents[i].parentId];
							hasParentQuestion = parent.type === "question";
							if (hasParentQuestion) {
								if (parent.children == null) parent.children = [];
								parent.children.push({parentId: parent.pageId, childId: question.pageId});
								break;
							}
						}
						if (hasParentQuestion) continue;
					}
					var $question = $compile("<arb-question primary-page-id='" + scope.pageId +
							"' page-id='" + question.pageId + "'></arb-question>")(scope);
					$questions.prepend($question);
				}
			};
		},
	};
});
