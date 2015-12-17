"use strict";

// Directive to show a lens' content
app.directive("arbLens", function($compile, $location, $timeout, $interval, $mdMedia, pageService, userService) {
	return {
		templateUrl: "/static/html/lens.html",
		scope: {
			pageId: "@",
			lensParentId: "@",
			selectedLens: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			if ($scope.lensParentId) {
				$scope.lensParentPage = pageService.pageMap[$scope.lensParentId];
			}
			$scope.isTinyScreen = !$mdMedia("gt-xs");
			$scope.isFloatingLhs = false; //$mdMedia("gt-lg");

			$scope.mastery = pageService.masteryMap[$scope.pageId];
			if (!$scope.mastery) {
				$scope.mastery = {has: false};
			}

			// Process mastery events.
			$scope.toggleMastery = function() {
				pageService.updateMastery($scope, $scope.page.pageId, !$scope.mastery.has);
				$scope.mastery = pageService.masteryMap[$scope.pageId];
			};

			// Process click on showing the page diff button.
			$scope.showingDiff = false;
			$scope.diffHtml = undefined;
			$scope.toggleDiff = function() {
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
			};

			// Check if this lens is actually visible
			$scope.lensIsVisible = true;
			$scope.$on("lensTabChanged", function(event, lensId){
				$scope.lensIsVisible = $scope.pageId == lensId;
			});
		},
		link: function(scope, element, attrs) {
			// Detach some elements and append them to the body, since they will appear
			// outside of the lens's div, and otherwise would be masked
			var $inlineCommentsDiv = element.find(".inline-comments-div");
			var $newInlineCommentButton = $inlineCommentsDiv.find(".inline-comment-icon");
			$inlineCommentsDiv.appendTo($("body"));
			var $lensMenuDiv = element.find(".lens-menu-div");
			if (scope.isFloatingLhs) {
				$lensMenuDiv.appendTo($("body"));
			}
			var inlineIconShiftLeft = $newInlineCommentButton.outerWidth() * ($mdMedia("gt-md") ? 0.5 : 1.1);
			scope.$on("$destroy", function() {
				$inlineCommentsDiv.remove();
			});

			// Get the position for the LHS lens menu div
			scope.getLensMenuDivStyle = function() {
				if (!scope.isFloatingLhs) return {};
				return {
					"left": $markdownContainer.offset().left - $lensMenuDiv.width(),
					"top": $markdownContainer.offset().top,
					"visibility": element.closest(".reveal-after-render-parent").length > 0 ? "hidden" : "visible",
				};
			};

			// =========================== Inline comments ===========================
			var $markdownContainer = element.find(".lens-text-container");
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
				if (scope.isTinyScreen) break;
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
					"left": $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
					"top": params.anchorNode.offset().top - $newInlineCommentButton.height() / 2,
					"visibility": element.closest(".reveal-after-render-parent").length > 0 ? "hidden" : "visible",
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
				params.anchorNode.toggleClass("inline-comment-highlight-hover", mouseover);
			};

			// Hide/show the inline comment
			var closeInlineComment = function(commentId) {
				var params = scope.inlineComments[commentId];
				if (!params.container) return;
				params.container.remove();
				params.container = undefined;
				params.anchorNode.toggleClass("inline-comment-highlight-hover", params.mouseover);
				params.visible = false;
			};
			scope.toggleInlineComment = function(commentId) {
				var params = scope.inlineComments[commentId];
				params.visible = !params.visible;
				if (params.visible) {
					// Close other inline comments
					for (var id in scope.inlineComments) {
						if (id !== commentId) {
							closeInlineComment(id);
						}
					}

					// Create the container
					params.container = $compile($("<arb-inline-comment" +
						" lens-id='" + scope.page.pageId +
						"' comment-id='" + commentId + "'></arb-inline-comment>"))(scope);
					$(params.paragraphNode).after(params.container);
				} else {
					closeInlineComment(commentId);
				}
			};

			// Process creating new inline comments
			var $inlineCommentEditPage = undefined;
			var newInlineCommentButtonTop = 0;
			scope.showNewInlineCommentButton = false;
			$markdown.on("mouseup", function(event) {
				if ($inlineCommentEditPage) return;
				// Do $timeout, because otherwise there is a bug when you double click to
				// select a word/paragraph, then click again and the selection var is still
				// the same (not cleared).
				$timeout(function(){
					scope.showNewInlineCommentButton = !!processSelectedParagraphText();
					if (scope.showNewInlineCommentButton) {
						newInlineCommentButtonTop = event.pageY;
					}
				});
			});
			scope.getNewInlineCommentButtonStyle = function() {
				return {
					"left": $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
					"top": newInlineCommentButtonTop - $newInlineCommentButton.height() / 2,
				};
			};

			// Create a new inline comment
			scope.newInlineComment = function() {
				var selection = getSelectedParagraphText();
				if (!selection) return;
				pageService.newComment({
					parentPageId: scope.page.pageId,
					success: function(newCommentId) {
						var comment = pageService.editMap[newCommentId];
						comment.anchorContext = selection.context;
						comment.anchorText = selection.text;
						comment.anchorOffset = selection.offset;
						$inlineCommentEditPage = $compile($("<div arb-edit-page class='edit-comment-embed'" +
							" is-embedded='true' page-id='" + newCommentId +
							"' done-fn='newInlineCommentDone(result)'></div>"))(scope);
						$(selection.paragraphNode).after($inlineCommentEditPage);
						scope.showNewInlineCommentButton = false;
					},
				});
			};

			// Called when the user is done with the new inline comment
			scope.newInlineCommentDone = function(result) {
				if (!result.discard) {
					pageService.newCommentCreated(result.pageId);
				} else {
					$inlineCommentEditPage.remove();
					$inlineCommentEditPage = undefined;
					$markdown.find(".inline-comment-highlight").removeClass("inline-comment-highlight");
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
								"' arb-vote-bar page-id='" + pageId + "' is-embedded='true' show-meta-info='true'></div>")(scope);
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
