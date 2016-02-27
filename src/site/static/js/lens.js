"use strict";

// Directive to show a lens' content
app.directive("arbLens", function($location, $compile, $timeout, $interval, $mdMedia, pageService, userService) {
	return {
		templateUrl: "static/html/lens.html",
		scope: {
			pageId: "@",
			lensParentId: "@",
			selectedLens: "=",
			isSimpleEmbed: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			if ($scope.lensParentId) {
				$scope.lensParentPage = pageService.pageMap[$scope.lensParentId];
			}
			$scope.isTinyScreen = !$mdMedia("gt-xs");
			$scope.isSmallScreen = !$mdMedia("gt-sm");

			$scope.mastery = pageService.masteryMap[$scope.pageId];
			if (!$scope.mastery) {
				$scope.mastery = {has: false};
			}

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

			// Listen for shortcut keys
			$(document).keyup(function(event) {
				if (!$scope.lensIsVisible) return true;
				if (!event.ctrlKey || !event.altKey) return true;
				$scope.$apply(function() {
					if (event.keyCode == 77) $scope.newInlineComment(); // M
				});
			});

			// ============ Masteries ====================

			// Compute subject ids that the user hasn't learned yet.
			$scope.subjectIds = $scope.page.subjectIds.filter(function(id) { return !pageService.hasMastery(id); });

			// Check if the user meets all requirements
			$scope.meetsAllRequirements = function(pageId) {
				var page = $scope.page;
				if (pageId) {
					page = pageService.pageMap[pageId];
				}
				for (var n = 0; n < page.requirementIds.length; n++) {
					if (!pageService.hasMastery(page.requirementIds[n])) {
						return false;
					}
				}
				return true;
			};
			$scope.showRequirementsPanel = !$scope.meetsAllRequirements();

			$scope.showRequirements = function() {
				$scope.showRequirementsPanel = true;
			};

			// Check if the user knows all the subjects
			$scope.knowsAllSubjects = function() {
				for (var n = 0; n < $scope.subjectIds.length; n++) {
					if (!pageService.hasMastery($scope.subjectIds[n])) {
						return false;
					}
				}
				return true;
			};
			$scope.showLearnedPanel = !$scope.knowsAllSubjects();

			// Toggle all requirements
			$scope.toggleRequirements = function() {
				if ($scope.meetsAllRequirements()) {
					pageService.updateMasteryMap({delete: $scope.page.requirementIds});
				} else {
					pageService.updateMasteryMap({knows: $scope.page.requirementIds});
				}
			};

			// Toggle all subjects
			$scope.toggleSubjects = function(continuePath) {
				var callback = $scope.pagesUnlocked;
				if (continuePath) {
					callback = function() {
						$location.url(pageService.getPageUrl(pageService.path.nextPageId));
					};
				}
				if ($scope.knowsAllSubjects()) {
					pageService.updateMasteryMap({delete: $scope.subjectIds, callback: callback});
				} else {
					pageService.updateMasteryMap({knows: $scope.subjectIds, callback: callback});
				}
			};

			var primaryPage = pageService.pageMap[$scope.lensParentId];
			var simplestLensId = primaryPage.lensIds[primaryPage.lensIds.length - 1];
			$scope.isSimplestLens = $scope.page.pageId === simplestLensId;

			// Compute simpler lens id if necessary
			if ($scope.showRequirementsPanel) {
				var simplerLensId = undefined;
				for (var n = $scope.page.lensIndex - 1; n >= 0; n--) {
					var lens = pageService.pageMap[primaryPage.lensIds[n]];
					if ($scope.meetsAllRequirements(lens.pageId)) {
						simplerLensId = lens.pageId;
						break;
					}
				}
				if (!simplerLensId && !$scope.isSimplestLens) {
					// We haven't found a lens for which we've met all requirements, so just suggest the simplest lens
					simplerLensId = simplestLensId;
				}
				if (simplerLensId) {
					$scope.simplerLens = pageService.pageMap[simplerLensId];
				}
			}

			$scope.getToggleSubjectsText = function() {
				if ($scope.knowsAllSubjects()) {
					if ($scope.page.subjectIds.length > 1) {
						return "Nevermind, none of them";
					} else {
						return "Nevermind, I didn't get it";
					}
				} else {
					if ($scope.page.subjectIds.length > 1) {
						return "Yes, all of them";
					} else {
						return "Yes, I got it";
					}
				}
			};

			// Check if the user can use the "yup, i got everything, let's continue" button.
			$scope.canQuickContinue = true;
			$scope.showQuickContinue = function() {
				return $scope.canQuickContinue && pageService.path && pageService.path.onPath && pageService.path.nextPageId;
			};

			// Called when the user unlocked some pages by acquiring requisites.
			$scope.pagesUnlocked = function(data) {
				$scope.canQuickContinue = false;
				$scope.unlockedIds = data && data.result && data.result.unlockedIds;
			};
		},
		link: function(scope, element, attrs) {
			if (scope.isSimpleEmbed) return;

			// Check if this lens is actually visible
			scope.lensIsVisible = true;
			scope.$on("lensTabChanged", function(event, lensId){
				var wasVisible = scope.lensIsVisible;
				scope.lensIsVisible = scope.pageId == lensId;
				if (wasVisible != scope.lensIsVisible && scope.lensIsVisible) {
					// This lens became visible. Sometimes this happens when the user is going through
					// a path and clicks "Next" at the bottom of the page. In this case we need to
					// scroll upwards to have them start reading this lens
					if ($("body").scrollTop() > element.offset().top) {
						$("body").scrollTop(element.offset().top - 100);
					}
				}
			});

			// Detach some elements and append them to the body, since they will appear
			// outside of the lens's div, and otherwise would be masked
			var $inlineCommentsDiv = element.find(".inline-comments-div");
			var $newInlineCommentButton = $inlineCommentsDiv.find(".inline-comment-icon");
			$inlineCommentsDiv.appendTo($("body"));
			var inlineIconShiftLeft = $newInlineCommentButton.outerWidth() * ($mdMedia("gt-md") ? 0.5 : 1.1);
			scope.$on("$destroy", function() {
				$inlineCommentsDiv.remove();
			});

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

			// Process an inline comment
			var processInlineComment = function(commentId) {
				if (scope.isTinyScreen) return;
				var comment = pageService.pageMap[commentId];
				if (!comment.anchorContext || !comment.anchorText) return;

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
				if (bestScore > comment.anchorContext.length / 2) return;

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
			};

			// Process all inline comments
			for (var n = 0; n < scope.page.commentIds.length; n++) {
				processInlineComment(scope.page.commentIds[n]);
			}

			// Get the style of an inline comment icon
			scope.getInlineCommentIconStyle = function(commentId) {
				var params = scope.inlineComments[commentId];
				var isVisible = element.closest(".reveal-after-render-parent").length <= 0;
				isVisible = isVisible && (!pageService.pageMap[commentId].isEditorComment || userService.showEditorComments);
				return {
					"left": $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
					"top": params.anchorNode.offset().top - $newInlineCommentButton.height() / 2,
					"visibility": isVisible ? "visible" : "hidden",
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
				if (!scope.showNewInlineCommentButton) return;
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
				$inlineCommentEditPage.remove();
				$inlineCommentEditPage = undefined;
				$markdown.find(".inline-comment-highlight").removeClass("inline-comment-highlight");
				if (!result.discard) {
					pageService.newCommentCreated(result.pageId);
					processInlineComment(result.pageId);
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
								"' class='md-whiteframe-2dp' arb-vote-bar page-id='" + pageId +
								"' is-embedded='true' show-meta-info='true'></div>")(scope);
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
