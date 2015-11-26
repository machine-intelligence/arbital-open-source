"use strict";

// Directive for showing a standard Arbital page.
app.directive("arbPage", function ($location, $compile, $timeout, pageService, userService) {
	return {
		templateUrl: "/static/html/page.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			scope.mastery = pageService.masteryMap[scope.pageId];
			scope.questionIds = scope.page.questionIds || [];

			// Add the primary page as the first lens.
			scope.page.lensIds.unshift(scope.page.pageId);

			// Determine which lens is selected
			scope.selectedLens = scope.page;
			if ($location.search().lens) {
				scope.selectedLens = pageService.pageMap[$location.search().lens];
			}
			scope.selectedLensIndex = scope.page.lensIds.indexOf(scope.selectedLens.pageId);

			// Manage switching between lenses, including loading the necessary data.
			scope.tabsMinHeight = 500; // hack to make the tabs body not collapse when transitioning
			var switchToLens = function(lensId) {
				if (lensId === scope.page.pageId) {
					$location.search("lens", undefined);
				} else {
					$location.search("lens", lensId);
				}
				scope.selectedLens = pageService.pageMap[lensId];
				$timeout(function() {
					scope.tabsMinHeight = 0;
				}, 1000);
			};
			scope.tabSelect = function(lensId) {
				if (scope.isLoaded(lensId)) {
					$timeout(function() {
						switchToLens(lensId);
					});
				} else {
					scope.tabsMinHeight = 500;
					pageService.loadLens(lensId, {
						success: function(data, status) {
							switchToLens(lensId);
						},
					});
				}
			};
			scope.isLoaded = function(lensId) {
				return pageService.pageMap[lensId].text.length > 0;
			};

			// Set up Page JS Controller.
			$timeout(function(){
				if (scope.page.subpageIds != null) {
					// Process subpages in two passes. First normal subpages.
					//processSubpages(false);
					$timeout(function() {
						// Inline subpages after a delay long enough for MathJax to have been processed.
						//processSubpages(true);
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
						var $comment = $compile("<arb-comment primary-page-id='" + scope.pageId +
																			"' page-id='" + subpage.pageId + "'></arb-comment>")(scope);
						$comments.prepend($comment);
					}
				}
			};
		},
	};
});
