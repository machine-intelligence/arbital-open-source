'use strict';
// jscs:disable requireCamelCaseOrUpperCaseIdentifiers

// Directive to show a lens' content
app.directive('arbLens', function($location, $compile, $timeout, $interval, $mdMedia, pageService, userService) {
	return {
		templateUrl: 'static/html/lens.html',
		scope: {
			pageId: '@',
			lensParentId: '@',
			isSimpleEmbed: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			if ($scope.lensParentId) {
				$scope.lensParentPage = pageService.pageMap[$scope.lensParentId];
			}
			$scope.isTinyScreen = !$mdMedia('gt-xs');
			$scope.isSmallScreen = !$mdMedia('gt-sm');

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
						var dmp = new diff_match_patch(); // jscs:ignore requireCapitalizedConstructors
						var diffs = dmp.diff_main(data[$scope.page.pageId].text, $scope.page.text);
						dmp.diff_cleanupSemantic(diffs);
						$scope.diffHtml = dmp.diff_prettyHtml(diffs).replace(/&para;/g, '');
					},
				});
			};

			// Compute how many visible comments there are.
			$scope.visibleCommentCount = function() {
				var count = 0;
				for (var n = 0; n < $scope.page.commentIds.length; n++) {
					var commentId = $scope.page.commentIds[n];
					count += (!pageService.pageMap[commentId].isEditorComment || userService.showEditorComments) ? 1 : 0;
				}
				return count;
			};

			// Listen for shortcut keys
			$(document).keyup(function(event) {
				if (!event.ctrlKey || !event.altKey) return true;
				$scope.$apply(function() {
					if (event.keyCode == 77) $scope.newInlineComment(); // M
					else if (event.keyCode == 85) $scope.newConfusedMark(); // U
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
						$timeout.cancel(callbackPromise);
						if (pageService.path.nextPageId) {
							// Go to the next page.
							$location.url(pageService.getPageUrl(pageService.path.nextPageId));
						} else {
							// This is the end of the path.
							pageService.abandonPath();
						}
					};
					// Make sure we execute the callback if we don't hear back from the server.
					var callbackPromise = $timeout(callback, 500);
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
				for (var n = $scope.page.lensIndex + 1; n < primaryPage.lensIds.length; n++) {
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
						return 'Nevermind, none of them';
					} else {
						return 'Nevermind, I didn\'t get it';
					}
				} else {
					if ($scope.page.subjectIds.length > 1) {
						return 'Yes, all of them';
					} else {
						return 'Yes, I got it';
					}
				}
			};

			// Check if the user can use the "yup, i got everything, let's continue" button.
			$scope.canQuickContinue = true;
			$scope.showQuickContinue = function() {
				return $scope.canQuickContinue && pageService.path && pageService.path.onPath;
			};
			$scope.getQuickContinueText = function() {
				if (pageService.path.nextPageId) {
					return 'Yes, I got this. Let\'s continue!';
				}
				return 'Yes, I got this. Now, I\'m all done!';
			};

			// Called when the user unlocked some pages by acquiring requisites.
			$scope.pagesUnlocked = function(data) {
				$scope.canQuickContinue = false;
				$scope.unlockedIds = data && data.result && data.result.unlockedIds;
			};
		},
		link: function(scope, element, attrs) {
			if (scope.isSimpleEmbed) return;

			// Detach some elements and append them to the body, since they will appear
			// outside of the lens's div, and otherwise would be masked
			var $inlineCommentsDiv = element.find('.inline-comments-div');
			var inlineCommentButtonHeight = 40;
			$inlineCommentsDiv.appendTo($('body'));
			var inlineIconShiftLeft = inlineCommentButtonHeight * ($mdMedia('gt-md') ? 0.5 : 1.1);
			scope.$on('$destroy', function() {
				$inlineCommentsDiv.remove();
			});

			// =========================== Inline elements ===========================
			var $markdownContainer = element.find('.lens-text-container');
			var $markdown = element.find('.lens-text');
			scope.inlineComments = {};
			scope.inlineMarks = {};
			var orderedInlineButtons = [];
			var dmp = new diff_match_patch(); // jscs:ignore requireCapitalizedConstructors
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

			// Given the anchor parameters, find the corresponding place in DOM.
			var computeInlineHightlightParams = function(anchorContext, anchorText, anchorOffset) {
				var results = {
					bestParagraphNode: undefined,
					bestParagraphIndex: 0,
					anchorOffset: 0,
					anchorLength: 0,
				}
				// Find the best paragraph
				var bestParagraphText;
				var bestScore = 9007199254740991; // Number.MAX_SAFE_INTEGER doesn't exist in IE
				if (!paragraphTexts) {
					populateParagraphTexts();
				}
				for (var i = 0; i < paragraphTexts.length; i++) {
					var text = paragraphTexts[i];
					var diffs = dmp.diff_main(text, anchorContext);
					var score = dmp.diff_levenshtein(diffs);
					if (score < bestScore) {
						results.bestParagraphNode = $markdown.children().get(i);
						bestParagraphText = text;
						bestScore = score;
						results.bestParagraphIndex = i;
					}
				}

				// Check if it's a close enough match
				if (bestScore > anchorContext.length / 2) return undefined;

				// Find offset into the best paragraph
				results.anchorOffset = dmp.match_main(bestParagraphText, anchorText, anchorOffset);
				if (results.anchorOffset < 0) {
					// Couldn't find a match within the paragraph. We'll just use paragraph as the anchor
					results.anchorOffset = 0;
					results.anchorLength = bestParagraphText.length;
				} else {
					// Figure out how long the highlighted anchor should be
					var remainingText = bestParagraphText.substring(results.anchorOffset);
					var diffs = dmp.diff_main(remainingText, anchorText);
					results.anchorLength = remainingText.length;
					if (diffs.length > 0) {
						// Note: we can potentially be more clever here and discount
						// edits done after anchorText.length chars
						var lastDiff = diffs[diffs.length - 1];
						if (lastDiff[0] < 0) {
							results.anchorLength -= lastDiff[1].length;
						}
					}
				}
				return results;
			};

			// Process a mark.
			var processMark = function(markId) {
				if (scope.isTinyScreen) return;
				var mark = pageService.markMap[markId];
				if (!mark.anchorContext || !mark.anchorText) return;

				// Create the span corresponding to the anchor text
				var highlightParams = computeInlineHightlightParams(mark.anchorContext,
						mark.anchorText, mark.anchorOffset);
				if (!highlightParams) return;
				var highlightClass = 'inline-mark-' + mark.id;
				createInlineCommentHighlight(highlightParams.bestParagraphNode, highlightParams.anchorOffset,
						highlightParams.anchorOffset + highlightParams.anchorLength, highlightClass);

				// Add to the array of valid inline comments
				var inlineMark = {
					paragraphNode: highlightParams.bestParagraphNode,
					anchorNode: $('.' + highlightClass),
					paragraphIndex: highlightParams.bestParagraphIndex,
					anchorOffset: highlightParams.anchorOffset,
					markId: mark.id,
				};
				scope.inlineMarks[mark.id] = inlineMark;
				orderedInlineButtons.push(inlineMark);
			};

			// Process an inline comment
			var processInlineComment = function(commentId) {
				if (scope.isTinyScreen) return;
				var comment = pageService.pageMap[commentId];
				if (!comment.anchorContext || !comment.anchorText) return;

				// Create the span corresponding to the anchor text
				var highlightParams = computeInlineHightlightParams(comment.anchorContext,
						comment.anchorText, comment.anchorOffset);
				if (!highlightParams) return;
				var highlightClass = 'inline-comment-' + comment.pageId;
				createInlineCommentHighlight(highlightParams.bestParagraphNode, highlightParams.anchorOffset,
						highlightParams.anchorOffset + highlightParams.anchorLength, highlightClass);

				// Add to the array of valid inline comments
				var inlineComment = {
					paragraphNode: highlightParams.bestParagraphNode,
					anchorNode: $('.' + highlightClass),
					paragraphIndex: highlightParams.bestParagraphIndex,
					anchorOffset: highlightParams.anchorOffset,
					pageId: comment.pageId,
				};
				scope.inlineComments[comment.pageId] = inlineComment;
				orderedInlineButtons.push(inlineComment);
			};

			// Process all inline comments
			for (var n = 0; n < scope.page.commentIds.length; n++) {
				processInlineComment(scope.page.commentIds[n]);
			}
			// Process all marks
			if ($location.search().markId) {
				processMark($location.search().markId);
			}
			for (var n = 0; n < scope.page.markIds.length; n++) {
				processMark(scope.page.markIds[n]);
			}
			
			// Process all RHS buttons to compute their position, zIndex, etc...
			// This fixes any potential overlapping issues.
			var preprocessInlineCommentButtons = function() {
				orderedInlineButtons.sort(function(a, b) {
					// Create arrays of values which we compare, breaking ties with the next item in the array.
					var arrayA = [a.paragraphIndex, a.anchorOffset, a.pageId, a.markId];
					var arrayB = [b.paragraphIndex, b.anchorOffset, b.pageId, b.markId];
					for (var i = 0; i < arrayA.length; i++) {
						if (arrayA[i] < arrayB[i]) { return -1; }
						if (arrayA[i] > arrayB[i]) { return 1; }
					}
					return 0;
				});
				var minTop = 0;
				for (n = 0; n < orderedInlineButtons.length; n++) {
					var inlineButton = orderedInlineButtons[n];
					var preferredTop = inlineButton.anchorNode.offset().top;
					var top = Math.max(minTop, preferredTop);
					// Use this to recompute the actual top when absolute positions are better known
					inlineButton.topOffset = top - preferredTop;
					inlineButton.zIndex = n;
					// Subtract 8 pixels to allow small overlap between buttons
					minTop = top + inlineCommentButtonHeight - 8;
				}
			};
			preprocessInlineCommentButtons();

			// Get the style of an inline comment icon
			scope.getInlineCommentIconStyle = function(commentId) {
				var params = scope.inlineComments[commentId];
				var isVisible = element.closest('.reveal-after-render-parent').length <= 0;
				isVisible = isVisible && (!pageService.pageMap[commentId].isEditorComment || userService.showEditorComments);
				return {
					'left': $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
					'top': params.anchorNode.offset().top - inlineCommentButtonHeight / 2 + params.topOffset,
					'visibility': isVisible ? 'visible' : 'hidden',
					'zIndex': params.zIndex,
				};
			};
			// Get the style of an inline mark icon
			scope.getInlineMarkIconStyle = function(markId) {
				var params = scope.inlineMarks[markId];
				var isVisible = element.closest('.reveal-after-render-parent').length <= 0;
				return {
					'left': $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
					'top': params.anchorNode.offset().top - inlineCommentButtonHeight / 2 + params.topOffset,
					'visibility': isVisible ? 'visible' : 'hidden',
					'zIndex': params.zIndex,
				};
			};

			// Return true iff the comment icon is selected
			scope.isInlineCommentIconSelected = function(commentId) {
				var params = scope.inlineComments[commentId];
				return params.mouseover || params.visible;
			};

			// Return true iff the mark icon is selected
			scope.isInlineMarkIconSelected = function(markId) {
				var params = scope.inlineMarks[markId];
				return params.mouseover || params.visible;
			};

			// Called when the user hovers the mouse over the inline comment icon
			scope.inlineCommentIconMouseover = function(commentId, mouseover) {
				var params = scope.inlineComments[commentId];
				params.mouseover = mouseover;
				if (params.visible) return;
				params.anchorNode.toggleClass('inline-comment-highlight-hover', mouseover);
			};

			// Called when the user hovers the mouse over the inline mark icon
			scope.inlineMarkIconMouseover = function(markId, mouseover) {
				var params = scope.inlineMarks[markId];
				params.mouseover = mouseover;
				if (params.visible) return;
				params.anchorNode.toggleClass('inline-comment-highlight-hover', mouseover);
			};

			// Hide/show the inline comment
			var closeInlineComment = function(commentId) {
				var params = scope.inlineComments[commentId];
				if (!params.container) return;
				params.container.remove();
				params.container = undefined;
				params.anchorNode.toggleClass('inline-comment-highlight-hover', params.mouseover);
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
					params.container = $compile($('<arb-inline-comment' +
						' lens-id=\'' + scope.page.pageId +
						'\' comment-id=\'' + commentId + '\'></arb-inline-comment>'))(scope);
					$(params.paragraphNode).after(params.container);
				} else {
					closeInlineComment(commentId);
				}
			};

			// Hide/show the inline mark.
			scope.toggleInlineMark = function(markId) {
				var params = scope.inlineMarks[markId];
				params.visible = !params.visible;
				pageService.hideEvent();
				if (params.visible) {
					showConfusionEventWindow(markId, false);
				}
			};

			// Process creating new inline comments
			var $inlineCommentEditPage = undefined;
			var newInlineCommentButtonTop = 0;
			scope.showNewInlineCommentButton = false;
			$markdownContainer.on('mouseup', function(event) {
				if ($inlineCommentEditPage) return;
				// Do $timeout, because otherwise there is a bug when you double click to
				// select a word/paragraph, then click again and the selection var is still
				// the same (not cleared).
				$timeout(function() {
					scope.showNewInlineCommentButton = !!processSelectedParagraphText();
					if (scope.showNewInlineCommentButton) {
						newInlineCommentButtonTop = event.pageY;
					}
				});
			});
			scope.getRhsButtonsStyle = function() {
				return {
					'left': $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
					'top': newInlineCommentButtonTop - inlineCommentButtonHeight / 2,
					'zIndex': orderedInlineButtons.length,
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
						$inlineCommentEditPage = $compile($('<div arb-edit-page class=\'edit-comment-embed\'' +
							' is-embedded=\'true\' page-id=\'' + newCommentId +
							'\' done-fn=\'newInlineCommentDone(result)\'></div>'))(scope);
						$(selection.paragraphNode).after($inlineCommentEditPage);
						scope.showNewInlineCommentButton = false;
					},
				});
			};

			// Called when the user is done with the new inline comment
			scope.newInlineCommentDone = function(result) {
				$inlineCommentEditPage.remove();
				$inlineCommentEditPage = undefined;
				$markdown.find('.inline-comment-highlight').removeClass('inline-comment-highlight');
				if (!result.discard) {
					pageService.newCommentCreated(result.pageId);
					processInlineComment(result.pageId);
					preprocessInlineCommentButtons();
				}
			};

			// Show all unresolved marks on this lens.
			scope.loadedUnresolvedMarks = false;
			scope.loadUnresolvedMarks = function() {
				scope.loadedUnresolvedMarks = true;
				pageService.loadUnresolvedMarks({pageId: scope.page.pageId}, function(data){
					for (var markId in data.marks) {
						processMark(markId);
					}
					preprocessInlineCommentButtons();
				});
			};

			// =========================== Inline questions ===========================

			// Create a new confused mark
			var showConfusionEventWindow = function(markId, isNew) {
				scope.showNewInlineCommentButton = false;
				pageService.showEvent({
					title: isNew ? 'New confusion mark' : 'Update your confusion mark',
					$element: $compile('<div arb-confusion-window mark-id="' + markId +
						'" is-new="::' + isNew +
						'"></div>')(scope),
				}, function() {
					var params = scope.inlineMarks[markId];
					params.visible = false;
					$markdown.find('.inline-comment-highlight-hover').removeClass('inline-comment-highlight-hover');
				});
			};
			scope.newConfusedMark = function() {
				if (!scope.showNewInlineCommentButton) return;
				var selection = getSelectedParagraphText();
				if (!selection) return;
				pageService.newMark({
						pageId: scope.pageId,
						edit: scope.page.edit,
						anchorContext: selection.context,
						anchorText: selection.text,
						anchorOffset: selection.offset,
					},
					function(data) {
						processMark(data.result.markId);
						preprocessInlineCommentButtons();
						$location.search("markId", data.result.markId);
						showConfusionEventWindow(data.result.markId, true);
					}
				);
			};

			// Process all embedded votes
			$timeout(function() {
				element.find('[embed-vote-id]').each(function(index) {
					var $link = $(this);
					var pageAlias = $link.attr('embed-vote-id');
					pageService.loadIntrasitePopover(pageAlias, {
						success: function(data, status) {
							var pageId = pageService.pageMap[pageAlias].pageId;
							var divId = 'embed-vote-' + pageId;
							var $embedDiv = $compile('<div id=\'' + divId +
								'\' class=\'md-whiteframe-2dp\' arb-vote-bar page-id=\'' + pageId +
								'\' is-embedded=\'true\' show-meta-info=\'true\'></div>')(scope);
							$link.replaceWith($embedDiv);
						},
						error: function(data, status) {
							console.error('Couldn\'t load embedded votes: ' + pageAlias);
						}
					});
				});
			});
		},
	};
});
