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
				pageAlias: pageId,
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
		// For question pages, check if we need to add the anchor text
		if (page.type === "question") {
			if (page.anchorContext && page.anchorText) {
				page.text = "> " + page.anchorText + "\n\n" + page.text;
			}
		}

		// Process mastery events.
		$topParent.find(".claim-mastery").on("click", function(event) {
			pageService.updateMastery(scope, pageId, true);
		});
		$topParent.find(".discard-mastery").on("click", function(event) {
			pageService.updateMastery(scope, pageId, false);
		});

		// Set up markdown.
		arbMarkdown.init(false, pageId, page.text, $topParent, pageService, userService);

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
					//createVoteSlider($topParent.find(".page-vote"), userService, page, false);
				});
			};
			$lensTab.on("click", doCreateVoteSlider);
			// Check to see if the tab pane is visible.
			if ($topParent.closest(".tab-pane").is(":visible")) {
				doCreateVoteSlider();
			}
		}

		// Process all embedded votes.
		/*window.setTimeout(function() {
			$topParent.find("[embed-vote-id]").each(function(index) {
				var $link = $(this);
				var pageAlias = $link.attr("embed-vote-id");
				pageService.loadPages([pageAlias], {
					includeAuxData: true,
					loadVotes: true,
					success: function(data, status) {
						var pageId = pageService.pageMap[pageAlias].pageId;
						var divId = "embed-vote-" + pageId;
						var $embedDiv = $compile("<div id='" + divId + "' class='embedded-vote'><arb-vote-bar page-id='" + pageId + "'></arb-vote-bar></div>")(scope);
						$link.replaceWith($embedDiv);
						//createVoteSlider($("#" + divId), userService, pageService.pageMap[pageId], false);
					},
					error: function(data, status) {
						console.log("Couldn't load embedded votes: " + pageAlias);
					}
				});
			});
		});*/
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
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.mastery = pageService.masteryMap[$scope.pageId];
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

				if (scope.page.subpageIds != null) {
					// Process subpages in two passes. First normal subpages.
					processSubpages(false);
					$timeout(function() {
						// Inline subpages after a delay long enough for MathJax to have been processed.
						processSubpages(true);
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
			var createNewInlineSubpageToggle = function(pageId, paragraphNode, anchorOffset, anchorLength, pageType) {
				var created = false;
				var doCreate = function() {
					created = true;
					var highlightClass = "inline-comment-" + pageId;
					var $commentDiv = $(".toggle-inline-comment-div.template").clone();
					$commentDiv.attr("id", "subpage-" + pageId).removeClass("template");
					var comment = pageService.pageMap[pageId];
					var commentCount = comment.children.length + 1;
					if (pageType == "comment") {
						$commentDiv.find(".inline-comment-count").text("" + commentCount);
					}
					if (pageType == "question") {
						$commentDiv.find(".inline-comment-count").text("?");
					}
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
						pageView.toggleInlineSubpage($commentDiv, function() {
							$("." + highlightClass).addClass("inline-comment-highlight");
							if (pageType == "comment") {
								var $comment = $compile("<arb-comment primary-page-id='" + scope.page.pageId +
																				"' page-id='" + pageId + "'></arb-comment>")(scope);
								$(".inline-comment-div").append($comment);
							}
							if (pageType == "question") {
								var $comment = $compile("<arb-question primary-page-id='" + scope.page.pageId +
																				"' page-id='" + pageId + "'></arb-question>")(scope);
								$(".inline-comment-div").append($comment);
							}
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
	
						// Check if we need to expand this inline subpage because of the URL anchor.
						var expandComment = window.location.hash === "#subpage-" + pageId;
						if (!expandComment) {
							// Check if one of the children is selected.
							for (var n = 0; n < comment.children.length; n++) {
								expandComment |= window.location.hash === "#subpage-" + comment.children[n].childId;
							}
						}
						if (expandComment) {
							// Delay to allow other inline subpage buttons to compute their position correctly.
							window.setTimeout(function() {
								$commentIcon.trigger("click");
								$("html, body").animate({
				      	  scrollTop: $(anchorNode).offset().top - 100
					    	}, 1000);
							}, 100);
						}
					} else {
						$commentDiv.hide();
						console.log("ERROR: couldn't find anchor node for inline subpage");
					}
				};
				// Check that we don't have another lens selected, in which case we'll
				// postpone creating the div.
				if (pageService.primaryPage === scope.page) {
					doCreate();
				}
				pageService.primaryPageCallbacks.push(function() {
					if (created) {
						$("#subpage-" + pageId).toggle(pageService.primaryPage === scope.page);
					} else if (pageService.primaryPage === scope.page) {
						window.setTimeout(function() {  // wait until the page shows
							doCreate();
						});
					}
				});
			}

			// Dynamically create subpage elements.
			var processSubpages = function(allowInline) {
				var $comments = element.find(".comments");
				var $markdown = element.find(".markdown-text");
				var dmp = new diff_match_patch();
				dmp.Match_MaxBits = 10000;
				dmp.Match_Distance = 10000;

				// If we have inline subpages, we'll need to compute the raw text for
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

				// Go through subpages in chronological order.
				scope.page.subpageIds.sort(pageService.getChildSortFunc("recentFirst"));
				for (var n = 0; n < scope.page.subpageIds.length; n++) {
					var subpage = pageService.pageMap[scope.page.subpageIds[n]];
					// Check if the subpage is anchored and we can still find the paragraph.
					if (subpage.anchorContext && subpage.anchorText) {
						if (!allowInline) continue;
						// Find the best paragraph.
						var bestParagraphNode, bestParagraphText, bestScore = Number.MAX_SAFE_INTEGER;
						if (!paragraphTexts) {
							populateParagraphTexts();
						}
						for (var i = 0; i < paragraphTexts.length; i++) {
							var text = paragraphTexts[i];
							var diffs = dmp.diff_main(text, subpage.anchorContext);
							var score = dmp.diff_levenshtein(diffs);
							if (score < bestScore) {
								bestParagraphNode = $markdown.children().get(i);
								bestParagraphText = text;
								bestScore = score;
							}
						}
						if (bestScore > subpage.anchorContext.length / 2) {
							// This is not a good paragraph match. Continue processing as a normal subpage.
							subpage.text = "> " + subpage.anchorText + "\n\n" + subpage.text;
							if (subpage.type == "question") {
									scope.questionIds.push(subpage.pageId);
									continue;
							}
						} else {
							// Find offset into the best paragraph.
							var anchorLength;
							var anchorOffset = dmp.match_main(bestParagraphText, subpage.anchorText, subpage.anchorOffset);
							if (anchorOffset < 0) {
								// Couldn't find a match within the paragraph. We'll just use paragraph as the anchor.
								anchorOffset = 0;
								anchorLength = bestParagraphText.length;
							} else {
								// Figure out how long the highlighted anchor should be.
								var remainingText = bestParagraphText.substring(anchorOffset);
								var diffs = dmp.diff_main(remainingText, subpage.anchorText);
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

							createNewInlineSubpageToggle(subpage.pageId, bestParagraphNode, anchorOffset, anchorLength, subpage.type);
							continue;
						}
					} else if (allowInline) {
						continue;
					}
					if (subpage.type == "comment" ) {
						// Make sure this comment is not a reply (i.e. it has a parent comment)
						// If it's a reply, add it as a child to the corresponding parent comment.
						if (subpage.parents != null) {
							var hasParentComment = false;
							for (var i = 0; i < subpage.parents.length; i++) {
								var parent = pageService.pageMap[subpage.parents[i].parentId];
								hasParentComment = parent.type === "comment";
								if (hasParentComment) {
									break;
								}
							}
							if (hasParentComment) continue;
						}
						var $comment = $compile("<arb-comment primary-page-id='" + scope.pageId +
																			"' page-id='" + subpage.pageId + "'></arb-comment>")(scope);
						$comments.prepend($comment);
					}
				}
			};
		},
	};
});

// Directive for showing a vote bar.
app.directive("arbVoteBar", function($compile, $timeout, pageService, userService) {
	return {
		templateUrl: "/static/html/voteBar.html",
		scope: {
			pageId: "@",
			isPopoverVote: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			var page = pageService.pageMap[scope.pageId];
			var userId = userService.user.id;

			if (!page.hasVote) return;

			// Convert votes into a user id -> {value, createdAt} map
			var voteMap = {};
			if (page.votes) {
				for(var i = 0; i < page.votes.length; i++) {
					var vote = page.votes[i];
					voteMap[vote.userId] = {value: vote.value, createdAt: vote.createdAt};
				}
			}
		
			// Copy vote-template and add it to the parent.
			var $voteDiv = element.attr("id", "vote" + page.pageId).appendTo(element);
			var $input = $voteDiv.find(".vote-slider-input");
			$input.attr("data-slider-id", $input.attr("data-slider-id") + page.pageId);
			var userVoteStr = userId in voteMap ? ("" + voteMap[userId].value) : "";
			var mySlider = $input.bootstrapSlider({
				step: 1,
				precision: 0,
				selection: "none",
				handle: "square",
				value: +userVoteStr,
				ticks: [1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99],
				formatter: function(s) { return s + "%"; },
			});
			var $tooltip = element.find(".tooltip-main");
		
			// Set the value of the user's vote.
			var setMyVoteValue = function($voteDiv, userVoteStr) {
				$voteDiv.attr("my-vote", userVoteStr);
				$voteDiv.find(".my-vote").toggle(userVoteStr !== "");
				$voteDiv.find(".my-vote-value").text("| my vote is \"" + (+userVoteStr) + "%\"");
			}
			setMyVoteValue($voteDiv, userVoteStr);
		
			// Setup vote bars.
			// A bar represents users' votes for a given value. The tiled background
			// allows us to display each vote separately.
			var bars = {}; // voteValue -> {bar: jquery bar element, users: array of user ids who voted on this value}
			// Stuff for setting up the bars' css.
			var $barBackground = element.find(".bar-background");
			var $sliderTrack = element.find(".slider-track");
			var originLeft = $sliderTrack.offset().left;
			var originTop = $sliderTrack.offset().top;
			var barWidth = Math.max(5, $sliderTrack.width() / (99 - 1) * 2);
			// Set the correct css for the given bar object given the number of votes it has.
			var setBarCss = function(bar) {
				var $bar = bar.bar;
				var voteCount = bar.users.length;
				$bar.toggle(voteCount > 0);
				$bar.css("height", 11 * voteCount);
				$bar.css("z-index", 2 + voteCount);
				$barBackground.css("height", Math.max($barBackground.height(), $bar.height()));
				$barBackground.css("top", 10);
			}
			var highlightBar = function($bar, highlight) {
				var css = "url(/static/images/vote-bar.png)";
				var highlightColor = "rgba(128, 128, 255, 0.3)";
				if(highlight) {
					css = "linear-gradient(" + highlightColor + "," + highlightColor + ")," + css;
				}
				$bar.css("background", css);
				$bar.css("background-size", "100% 11px"); // have to set this each time
			};
			// Get the bar object corresponding to the given vote value. Create a new one if there isn't one.
			var getBar = function(vote) {
				if (!(vote in bars)) {
					var x = (vote - 1) / (99 - 1);
					var $bar = $("<div></div>").addClass("vote-bar");
					$bar.css("left", x * $sliderTrack.width() - barWidth / 2);
					$bar.css("width", barWidth);
					$barBackground.append($bar);
					bars[vote] = {bar: $bar, users: []};
				}
				return bars[vote];
			}
			for(var id in voteMap){
				// Create stacks for all the votes.
				var bar = getBar(voteMap[id].value);
				bar.users.push(id);
				setBarCss(bar);
			}
		
			// Convert mouse X position into % vote value.
			var voteValueFromMousePosX = function(mouseX) {
				var x = (mouseX - $sliderTrack.offset().left) / $sliderTrack.width();
				x = Math.max(0, Math.min(1, x));
				return Math.round(x * (99 - 1) + 1);
			};
		
			// Update the label that shows how many votes have been cast.
			var updateVoteCount = function() {
				var votesLength = Object.keys(voteMap).length;
				$voteDiv.find(".vote-count").text(votesLength + " vote" + (votesLength == 1 ? "" : "s"));
			};
			updateVoteCount();
		
			// Set handle's width.
			var $handle = element.find(".min-slider-handle");
			$handle.css("width", barWidth);
		
			// Don't track mouse movements and such for the vote in a popover.
			if (scope.isPopoverVote) {
				if (!(userId in voteMap)) {
					$handle.hide();
				}
				return;
			}
		
			// Catch mousemove event on the text, so that it doesn't propagate to parent
			// and spawn popovers, which will prevent us clicking on "x" button to delete
			// our vote.
			element.find(".text-center").on("mousemove", function(event){
				return false;
			});
		
			var mouseInParent = false;
			var mouseInPopover = false;
			// Destroy the popover that displays who voted for a given value.
			var $usersPopover = undefined;
			var destroyUsersPopover = function() {
				if ($usersPopover !== undefined) {
					$usersPopover.popover("destroy");
					highlightBar($usersPopover, false);
				}
				$usersPopover = undefined;
				mouseInPopover = false;
			};
		
			// Track mouse movement to show voter names.
			element.on("mouseenter", function(event) {
				mouseInParent = true;
				$handle.show();
				$tooltip.css("opacity", 1.0);
			});
			element.on("mouseleave", function(event) {
				mouseInParent = false;
				if (!(userId in voteMap)) {
					$handle.hide();
				} else {
					$input.bootstrapSlider("setValue", voteMap[userId].value);
				}
				$tooltip.css("opacity", 0.0);
				if (!mouseInPopover) {
					destroyUsersPopover();
				}
			});
			element.trigger("mouseleave");
			element.on("mousemove", function(event) {
				// Update slider.
				var voteValue = voteValueFromMousePosX(event.pageX);
				$input.bootstrapSlider("setValue", voteValue);
				if (mouseInPopover) return true;
		
				// We do a funky search to see if there is a vote nearby, and if so, show popover.
				var offset = 0, maxOffset = 5;
				var offsetSign = 1;
				var foundBar = false;
				while(offset <= maxOffset) {
					if(offsetSign < 0) offset++;
					offsetSign = -offsetSign;
					var hoverVoteValue = voteValue + offsetSign * offset;
					if (!(hoverVoteValue in bars)) {
						continue;
					}
					foundBar = true;
					var bar = bars[hoverVoteValue];
					// Don't do anything if it's still the same bar as last time.
					if (bar.bar === $usersPopover) {
						break;
					}
					// Destroy old one.
					destroyUsersPopover();
					// Create new popover.
					$usersPopover = bar.bar;
					highlightBar(bar.bar, true);
					$usersPopover.popover({
						html : true,
						placement: "bottom",
						trigger: "manual",
						title: "Voters (" + hoverVoteValue + "%)",
						content: function() {
							var $anchor = $(this);
							var html = "";
							for(var i = 0; i < bar.users.length; i++) {
								var userId = bar.users[i];
								html += "<arb-user-name user-id='" + userId + "'></arb-user-name> " +
									"<span class='gray-text'>{[{'" + voteMap[userId].createdAt + "' | relativeDateTime}]}</span><br>";
							}
							$timeout(function() {
								var $popover = $("#" + $anchor.attr("aria-describedby"));
								$compile($popover)(scope);
							});
							return html;
						}
					}).popover("show");
					var $popover = $barBackground.find(".popover");
					$popover.on("mouseenter", function(event){
						mouseInPopover = true;
					});
					$popover.on("mouseleave", function(event){
						mouseInPopover = false;
						if (!mouseInParent) {
							destroyUsersPopover();
						}
					});
					break;
				}
				if (!foundBar) {
					// We didn't find a bar, so destroy any existing popover.
					destroyUsersPopover();
				}
			});
		
			// Handle user's request to delete their vote.
			$voteDiv.find(".delete-my-vote-link").on("click", function(event) {
				var bar = bars[voteMap[userId].value];
				bar.users.splice(bar.users.indexOf(userId), 1);
				setBarCss(bar);
				if (bar.users.length <= 0){
					delete bars[voteMap[userId].value];
				}
		
				mouseInPopover = false;
				mouseInParent = false;
				delete voteMap[userId];
				element.trigger("mouseleave");
				element.trigger("mouseenter");
		
				postNewVote(page.pageId, 0.0);
				setMyVoteValue($voteDiv, "");
				updateVoteCount();
				return false;
			});
			
			// Track click to see when the user wants to vote / update their vote.
			element.on("click", function(event) {
				if (mouseInPopover) return true;
				if (userId === "0") {
					showSignupPopover($(event.currentTarget));
					return true;
				}
				if (userId in voteMap && voteMap[userId].value in bars) {
					// Update old bar.
					var bar = bars[voteMap[userId].value];
					bar.users.splice(bar.users.indexOf(userId), 1);
					setBarCss(bar);
					destroyUsersPopover();
					if (bar.users.length <= 0) {
						delete bars[voteMap[userId].value];
					}
				}
		
				// Set new vote and update all the things.
				var vote = voteValueFromMousePosX(event.pageX); 
				voteMap[userId] = {value: vote, createdAt: moment.utc().format("YYYY-MM-DD HH:mm:ss")};
				postNewVote(page.pageId, vote);
				setMyVoteValue($voteDiv, "" + vote);
				updateVoteCount();
		
				// Update new bar.
				var bar = getBar(vote);
				bar.users.push(userId);
				setBarCss(bar);
			});
		},
	};
});

// Directive for showing a the panel with requirements.
app.directive("arbRequirementsPanel", function(pageService, userService, autocompleteService, $timeout, $http) {
	return {
		templateUrl: "/static/html/requirementsPanel.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			if (!scope.page.requirementIds) {
				scope.page.requirementIds = [];
			}
			
			// Setup autocomplete for input field.
			autocompleteService.setupParentsAutocomplete(element.find(".tag-input"), function(event, ui) {
				var data = {
					parentId: scope.page.pageId,
					childId: ui.item.label,
					type: "requirement",
				};
				$http({method: "POST", url: "/newPagePair/", data: JSON.stringify(data)})
					.error(function(data, status){
						console.log("Error creating requirement:"); console.log(data); console.log(status);
					});

				pageService.masteryMap[data.childId] = {pageId: data.childId, isMet: true, isManuallySet: true};
				scope.page.requirementIds.push(data.childId);
				scope.$apply();
				$(event.target).val("");
				return false;
			});

			// Process deleting requirements.
			element.on("click", ".delete-requirement-link", function(event) {
				var $target = $(event.target);
				var options = {
					parentId: scope.page.pageId,
					childId: $target.attr("page-id"),
					type: "requirement",
				};
				pageService.deletePagePair(options);

				scope.page.requirementIds.splice(scope.page.requirementIds.indexOf(options.childId), 1);
				scope.$apply();
			});

			element.on("click", ".requirement-not-met", function(event) {
				pageService.updateMastery(scope, $(event.target).attr("page-id"), true);
			});

			$timeout(function() {
				// Set the rendering for tags autocomplete
				autocompleteService.setAutocompleteRendering(element.find(".tag-input"), scope);
			});
		},
	};
});
