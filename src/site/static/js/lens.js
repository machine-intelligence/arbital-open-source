"use strict";

// Directive to show a lens' content
app.directive("arbLens", function($compile, $location, $timeout, $interval, $anchorScroll, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/lens.html",
		scope: {
			pageId: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];

			$scope.mastery = pageService.masteryMap[$scope.pageId];
			if (!$scope.mastery) {
				$scope.mastery = {has: false};
			}

			// Process mastery events.
			$scope.toggleMastery = function() {
				pageService.updateMastery($scope, $scope.page.pageId, !$scope.mastery.has);
				$scope.mastery = pageService.masteryMap[$scope.pageId];
			};

			// Called from the page when the user shows/hides the diff
			$scope.showingDiff = false;
			$scope.diffHtml = undefined;
			$scope.$on("toggleDiff", function(e, lensId) {
				if (lensId !== $scope.page.pageId) return;
				$scope.showingDiff = !$scope.showingDiff;
				if (!$scope.showingDiff) return;
				if ($scope.diffHtml) return;

				var earliest = $scope.page.lastVisit;
				if (moment($scope.page.createdAt).isBefore(earliest)) {
					earliest = $scope.page.createdAt;
				}
				// Load the edit from the server.
				pageService.loadEdit({
					pageAlias: $scope.page.pageId,
					createdAtLimit: earliest,
					skipProcessDataStep: true,
					success: function(data, status) {
						var dmp = new diff_match_patch();
						var diffs = dmp.diff_main(data[$scope.page.pageId].text, $scope.page.text);
						dmp.diff_cleanupSemantic(diffs);
						$scope.diffHtml = dmp.diff_prettyHtml(diffs).replace(/&para;/g, "");
					},
				});
			});
		},
		link: function(scope, element, attrs) {

			// =========================== Inline comments ===========================
			element.find(".inline-comments-div").appendTo($("body"));
			var $markdown = element.find(".lens-text");
			scope.inlineComments = {};
			var dmp = new diff_match_patch();
			dmp.Match_MaxBits = 10000;
			dmp.Match_Distance = 10000;

			// Compute the raw text for each paragraph; on demand
			var paragraphTexts = undefined;
			var populateParagraphTexts = function() {
				paragraphTexts = [];
				var i = 0;
				$markdown.children().each(function() {
					paragraphTexts.push(getParagraphText($(this).get(0)).context);
					i++;
				});
			};

			// Process all inline comments
			for (var n = 0; n < scope.page.commentIds.length; n++) {
				var comment = pageService.pageMap[scope.page.commentIds[n]];
				if (!comment.anchorContext || !comment.anchorText) continue;

				// Find the best paragraph
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

				// Check if it's a close enough match
				if (bestScore > comment.anchorContext.length / 2) continue;

				// Find offset into the best paragraph
				var anchorLength;
				var anchorOffset = dmp.match_main(bestParagraphText, comment.anchorText, comment.anchorOffset);
				if (anchorOffset < 0) {
					// Couldn't find a match within the paragraph. We'll just use paragraph as the anchor
					anchorOffset = 0;
					anchorLength = bestParagraphText.length;
				} else {
					// Figure out how long the highlighted anchor should be
					var remainingText = bestParagraphText.substring(anchorOffset);
					var diffs = dmp.diff_main(remainingText, comment.anchorText);
					anchorLength = remainingText.length;
					if (diffs.length > 0) {
						// Note: we can potentially be more clever here and discount
						// edits done after anchorText.length chars
						var lastDiff = diffs[diffs.length - 1];
						if (lastDiff[0] < 0) {
							anchorLength -= lastDiff[1].length;
						}
					}
				}

				// Create the span corresponding to the anchor text
				var highlightClass = "inline-comment-" + comment.pageId;
				createInlineCommentHighlight(bestParagraphNode, anchorOffset, anchorOffset + anchorLength, highlightClass);

				// Add to the array of valid inline comments
				scope.inlineComments[comment.pageId] = {
					paragraphNode: bestParagraphNode,
					anchorNode: $("." + highlightClass),
				};
			}

			// Get the style of an inline comment icon
			scope.getInlineCommentIconStyle = function(commentId) {
				var params = scope.inlineComments[commentId];
				return {
					"left": $markdown.position().left + $markdown.width(),
					"top": params.anchorNode.offset().top,
				};
			};

			// Return true iff the comment icon is selected
			scope.isInlineCommentIconSelected = function(commentId) {
				var params = scope.inlineComments[commentId];
				return params.mouseover || params.visible;
			};

			// Called when the use hovers the mouse over the icon
			scope.inlineCommentIconMouseover = function(commentId, mouseover) {
				var params = scope.inlineComments[commentId];
				params.mouseover = mouseover;
				if (params.visible) return;
				params.anchorNode.toggleClass("inline-comment-text", mouseover);
			};

			// Hide/show the inline comment
			scope.toggleInlineComment = function(commentId) {
				var params = scope.inlineComments[commentId];
				params.visible = !params.visible;
				if (params.visible) {
					params.container = $compile($("<div arb-subpage" +
						" class='inline-comment-container md-whiteframe-4dp' lens-id='" + scope.page.pageId +
						"' page-id='" + commentId + "'></div>"))(scope);
					$(params.paragraphNode).after(params.container);
					$location.hash("subpage-" + commentId);
				} else {
					params.container.remove();
					params.container = undefined;
					params.anchorNode.toggleClass("inline-comment-text", params.mouseover);
				}
			};

			// Process all embedded votes
			$timeout(function() {
				element.find("[embed-vote-id]").each(function(index) {
					var $link = $(this);
					var pageAlias = $link.attr("embed-vote-id");
					pageService.loadIntrasitePopover(pageAlias, {
						success: function(data, status) {
							var pageId = pageService.pageMap[pageAlias].pageId;
							var divId = "embed-vote-" + pageId;
							var $embedDiv = $compile("<div id='" + divId +
								"' class='embedded-vote'><arb-vote-bar page-id='" + pageId + "'></arb-vote-bar></div>")(scope);
							$link.replaceWith($embedDiv);
						},
						error: function(data, status) {
							console.error("Couldn't load embedded votes: " + pageAlias);
						}
					});
				});
			});
		},
	};
});

