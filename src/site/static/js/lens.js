'use strict';
// jscs:disable requireCamelCaseOrUpperCaseIdentifiers

// Directive to show a lens' content
app.directive('arbLens', function($http, $location, $compile, $timeout, $interval, $mdMedia, $mdBottomSheet, $rootScope,
	arb) {
	return {
		templateUrl: 'static/html/lens.html',
		scope: {
			pageId: '@',
			lensParentId: '@',
			isSimpleEmbed: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.page = arb.pageService.pageMap[$scope.pageId];
			if ($scope.lensParentId) {
				$scope.lensParentPage = arb.pageService.pageMap[$scope.lensParentId];
			}
			$scope.isTinyScreen = !$mdMedia('gt-xs');
			$scope.isSmallScreen = !$mdMedia('gt-sm');

			$scope.mastery = arb.masteryService.masteryMap[$scope.pageId];
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
				if (moment($scope.page.editCreatedAt).isBefore(earliest)) {
					earliest = $scope.page.editCreatedAt;
				}
				// Load the edit from the server.
				arb.pageService.loadEdit({
					pageAlias: $scope.page.pageId,
					createdAtLimit: earliest,
					skipProcessDataStep: true,
					success: function(data) {
						$scope.diffHtml = arb.diffService.getDiffHtml($scope.page.text, data.edits[$scope.page.pageId].text);
					},
				});
			};

			// Process click on "Subscribe as maintainer"
			$scope.toggleMaintainerSub = function() {
				$scope.page.isSubscribedAsMaintainer = !$scope.page.isSubscribedAsMaintainer;
				if ($scope.page.isSubscribedAsMaintainer) {
					$scope.page.isSubscribed = true;
				}

				$http({method: 'POST', url: '/updateSubscription/', data: JSON.stringify({
					toId: $scope.pageId,
					isSubscribed: $scope.page.isSubscribed,
					asMaintainer: $scope.page.isSubscribedAsMaintainer,
				})});
			};

			// Compute how many visible comments there are.
			$scope.visibleCommentCount = function() {
				var count = 0;
				for (var n = 0; n < $scope.page.commentIds.length; n++) {
					var commentId = $scope.page.commentIds[n];
					count += (!arb.pageService.pageMap[commentId].isEditorComment || arb.stateService.getShowEditorComments()) ? 1 : 0;
				}
				return count;
			};

			// Listen for shortcut keys
			$(document).keyup(function(event) {
				if (!event.ctrlKey || !event.altKey) return true;
				$scope.$apply(function() {
					if (event.keyCode == 77) $scope.newInlineComment(); // M
					else if (event.keyCode == 85) $scope.newQueryMark(); // U
				});
			});

			// ============ Masteries ====================

			// Compute subject ids that the user hasn't learned yet.
			$scope.subjectIds = $scope.page.subjectIds.filter(function(id) { return !arb.masteryService.hasMastery(id); });

			// Check if the user meets all requirements
			$scope.meetsAllRequirements = function(pageId) {
				var page = $scope.page;
				if (pageId) {
					page = arb.pageService.pageMap[pageId];
				}
				for (var n = 0; n < page.requirementIds.length; n++) {
					if (!arb.masteryService.hasMastery(page.requirementIds[n])) {
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
					if (!arb.masteryService.hasMastery($scope.subjectIds[n])) {
						return false;
					}
				}
				return true;
			};
			$scope.showLearnedPanel = !$scope.knowsAllSubjects();

			// Toggle all requirements
			$scope.toggleRequirements = function() {
				if ($scope.meetsAllRequirements()) {
					arb.masteryService.updateMasteryMap({delete: $scope.page.requirementIds});
				} else {
					arb.masteryService.updateMasteryMap({knows: $scope.page.requirementIds});
				}
			};

			// Toggle all subjects
			$scope.toggleSubjects = function(continuePath) {
				var callback = $scope.pagesUnlocked;
				if (continuePath) {
					callback = function() {
						$timeout.cancel(callbackPromise);
						if (arb.pathService.path.nextPageId) {
							// Go to the next page.
							arb.urlService.goToUrl(arb.urlService.getPageUrl(arb.pathService.path.nextPageId));
						} else {
							// This is the end of the path.
							arb.pathService.abandonPath();
						}
					};
					// Make sure we execute the callback if we don't hear back from the server.
					var callbackPromise = $timeout(callback, 500);
				}
				if ($scope.knowsAllSubjects()) {
					arb.masteryService.updateMasteryMap({delete: $scope.subjectIds, callback: callback});
				} else {
					arb.masteryService.updateMasteryMap({knows: $scope.subjectIds, callback: callback});
				}
			};

			var primaryPage = arb.pageService.pageMap[$scope.lensParentId];
			var simplestLensId = primaryPage.lensIds[primaryPage.lensIds.length - 1];
			$scope.isSimplestLens = $scope.page.pageId === simplestLensId;

			// Compute simpler lens id if necessary
			if ($scope.showRequirementsPanel) {
				var simplerLensId = undefined;
				for (var n = $scope.page.lensIndex + 1; n < primaryPage.lensIds.length; n++) {
					var lens = arb.pageService.pageMap[primaryPage.lensIds[n]];
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
					$scope.simplerLens = arb.pageService.pageMap[simplerLensId];
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
				return $scope.canQuickContinue && arb.pathService.path && arb.pathService.path.onPath;
			};
			$scope.getQuickContinueText = function() {
				if (arb.pathService.path.nextPageId) {
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

			var processInlineEverything = function() {
				var inlineCommentButtonHeight = 40;
				var inlineIconShiftLeft = inlineCommentButtonHeight * ($mdMedia('gt-md') ? 0.5 : 1.1);

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
					};
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
					var mark = arb.markService.markMap[markId];
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
					var comment = arb.pageService.pageMap[commentId];
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
					try {
						processInlineComment(scope.page.commentIds[n]);
					} catch (err) {
						console.error(err);
					}
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
				var orderRhsButtons = function() {
					orderedInlineButtons.sort(function(a, b) {
						// Create arrays of values which we compare, breaking ties with the next item in the array.
						var arrayA = [a.paragraphIndex, a.anchorOffset, a.markId];
						var arrayB = [b.paragraphIndex, b.anchorOffset, b.markId];
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
				orderRhsButtons();

				// Get the style of an inline comment icon
				scope.getInlineCommentIconStyle = function(commentId) {
					var params = scope.inlineComments[commentId];
					var isVisible = element.closest('.reveal-after-render-parent').length <= 0;
					isVisible = isVisible && (!arb.pageService.pageMap[commentId].isEditorComment || arb.stateService.getShowEditorComments());
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
					if (!params) return;
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
					params.anchorNode.toggleClass('inline-comment-highlight-hover', mouseover || params.visible);
				};

				// Called when the user hovers the mouse over the inline mark icon
				scope.inlineMarkIconMouseover = function(markId, mouseover) {
					var params = scope.inlineMarks[markId];
					if (!params) return;
					params.mouseover = mouseover;
					params.anchorNode.toggleClass('inline-comment-highlight-hover', mouseover || params.visible);
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
					if (!params) return;
					params.visible = !params.visible;
					arb.popupService.hidePopup();
					if (params.visible) {
						if (arb.markService.markMap[markId].type === 'query') {
							showQueryMarkWindow(markId, false);
						} else {
							showEditorMarkWindow(markId);
						}
						scope.inlineMarkIconMouseover(markId, true);
						$location.replace().search('markId', markId);
					}
				};

				// Process creating new inline comments
				var $inlineCommentEditPage = undefined;
				var newInlineCommentButtonTop = 0;
				scope.showRhsButtons = false;

				// Handle text selection.
				var cachedSelection;
				if (arb.isTouchDevice) {
					// On mobile it's very hard to get user's selected text. The best way Alexei found
					// was to just check for selected text every so often.
					$interval(function() {
						if ($inlineCommentEditPage) return;
						arb.stateService.lensTextSelected = !!processSelectedParagraphText();
						if (arb.stateService.lensTextSelected) {
							// Cache the selection we found, because we are pretty much guaranteed to
							// lose it as soon as the user clicks on anything.
							cachedSelection = getStartEndSelection();
						}
					}, 500);

					// Called when the fab is clicked when text is selected.
					scope.$on('fabClicked', function() {
						$mdBottomSheet.show({
							templateUrl: 'static/html/rhsButtons.html',
							controller: 'RhsButtonsController',
							parent: '#fixed-overlay',
						}).then(function(result) {
							scope[result.func].apply(null, result.params);
							arb.stateService.lensTextSelected = false;
						});
					});
				} else {
					var mouseUpFn = function(event) {
						if ($inlineCommentEditPage) return;
						// Do $timeout, because otherwise there is a bug when you double click to
						// select a word/paragraph, then click again and the selection var is still
						// the same (not cleared).
						$timeout(function() {
							scope.showRhsButtons = !!processSelectedParagraphText();
							arb.stateService.lensTextSelected = !!processSelectedParagraphText();
							if (scope.showRhsButtons) {
								newInlineCommentButtonTop = event.pageY;
							}
						});
					};
					$('body').on('mouseup', mouseUpFn);
					scope.$on('$destroy', function() {
						$('body').off('mouseup', mouseUpFn);
					});
				}

				scope.getRhsButtonsStyle = function() {
					return {
						'left': $markdownContainer.offset().left + $markdownContainer.outerWidth() - inlineIconShiftLeft,
						'top': newInlineCommentButtonTop - inlineCommentButtonHeight / 2,
						'zIndex': orderedInlineButtons.length + 2,
					};
				};

				// Create a new inline comment
				scope.newInlineComment = function() {
					var selection = getSelectedParagraphText(cachedSelection);
					if (!selection) return;
					arb.pageService.newComment({
						parentPageId: scope.page.pageId,
						success: function(newCommentId) {
							var comment = arb.pageService.editMap[newCommentId];
							comment.anchorContext = selection.context;
							comment.anchorText = selection.text;
							comment.anchorOffset = selection.offset;
							$inlineCommentEditPage = $compile($('<div arb-edit-page class=\'edit-comment-embed\'' +
								' is-embedded=\'true\' page-id=\'' + newCommentId +
								'\' done-fn=\'newInlineCommentDone(result)\'></div>'))(scope);
							$(selection.paragraphNode).after($inlineCommentEditPage);
							scope.showRhsButtons = false;
						},
					});
				};

				// Called when the user is done with the new inline comment
				scope.newInlineCommentDone = function(result) {
					$inlineCommentEditPage.remove();
					$inlineCommentEditPage = undefined;
					$markdown.find('.inline-comment-highlight').removeClass('inline-comment-highlight');
					if (!result.discard) {
						arb.pageService.newCommentCreated(result.pageId);
						processInlineComment(result.pageId);
						orderRhsButtons();
					}
				};

				// Show all marks on this lens.
				scope.loadedMarks = false;
				scope.loadMarks = function() {
					scope.loadedMarks = true;
					arb.markService.loadMarks({pageId: scope.page.pageId}, function(data) {
						for (var markId in data.marks) {
							processMark(markId);
						}
						orderRhsButtons();
					});
				};

				scope.isEditorFeedbackFabOpen = false;
				scope.toggleEditorFeedbackFab = function(show) {
					scope.isEditorFeedbackFabOpen = show;
				};

				// =========================== Inline questions ===========================

				// Helper to call when a mark window has been closed.
				var markWindowClosed = function(markId, dismiss) {
					if (scope.$$destroyed) return;
					if (dismiss) {
						delete scope.inlineMarks[markId];
						for (var n = 0; n < orderedInlineButtons.length; n++) {
							var button = orderedInlineButtons[n];
							if (button.markId == markId) {
								orderedInlineButtons.splice(n, 1);
								break;
							}
						}
						orderRhsButtons();
					}
					if (markId in scope.inlineMarks) {
						var params = scope.inlineMarks[markId];
						if (params) {
							params.visible = false;
							params.mouseover = false;
						}
					}
					$markdown.find('.inline-comment-highlight').removeClass('inline-comment-highlight');
					$markdown.find('.inline-comment-highlight-hover').removeClass('inline-comment-highlight-hover');
					if ($location.search().markId == markId) {
						// TODO: GAH! We can't erase markId here, because then we'll erase it when the user goes to edit
						// page url with ?markId set. We should fix this with a better URL state management system.
						//$location.replace().search("markId", undefined);
					}
				};

				// Show the window for editing a query mark.
				var showQueryMarkWindow = function(markId, isNew) {
					scope.showRhsButtons = false;
					arb.popupService.showPopup({
						title: isNew ? 'New query mark' : 'Edit query mark',
						$element: $compile('<div arb-query-info mark-id="' + markId +
							'" is-new="::' + isNew +
							'" in-popup="::true' +
							'"></div>')($rootScope),
						persistent: true,
					}, function(result) {
						markWindowClosed(markId, result.dismiss);
					});
				};

				// Show the window for editing an editor mark.
				var showEditorMarkWindow  = function(markId) {
					scope.showRhsButtons = false;
					arb.popupService.showPopup({
						title: 'Edit mark',
						$element: $compile('<div arb-mark-info mark-id="' + markId +
							'" is-new="::false' +
							'"></div>')($rootScope),
					}, function(result) {
						markWindowClosed(markId, result.dismiss);
					});
				};

				// Helper for creating a new mark.
				var newMark = function(type, success) {
					var selection = getSelectedParagraphText(cachedSelection, type != 'query');
					if (!selection && type !== 'query') return;
					arb.markService.newMark({
							pageId: scope.pageId,
							edit: scope.page.edit,
							type: type,
							anchorContext: selection ? selection.context : undefined,
							anchorText: selection ? selection.text : undefined,
							anchorOffset: selection ? selection.offset : undefined,
						},
						function(data) {
							var markId = data.result.markId;
							processMark(markId);
							orderRhsButtons();
							var params = scope.inlineMarks[markId];
							if (params && type == 'query') {
								params.visible = true;
							}
							success(data);
						}
					);
				};

				// Called to create a new query (question/objection) mark.
				scope.newQueryMark = function() {
					newMark('query', function(data) {
						showQueryMarkWindow(data.result.markId, true);
					});
				};

				// Called to create a new editor (confusion/spelling) mark.
				scope.newEditorMark = function(type) {
					newMark(type, function(data) {
						window.getSelection().removeAllRanges();
						scope.showRhsButtons = false;

						scope.toastCallback = function() {
							showEditorMarkWindow(data.result.markId);
						};
						arb.popupService.showToast({
							text: 'Thanks for your feedback!',
							scope: scope,
							normalButton: {
								text: 'Edit',
								callbackText: 'toastCallback()',
							},
						});
					});
				};

				// Scroll down to selected markId
				$timeout(function() {
					var markId = $location.search().markId;
					if (!markId) return;
					scope.inlineMarkIconMouseover(markId, true);
					scope.toggleInlineMark(markId);
					var style = scope.getInlineMarkIconStyle(markId);
					if (style) {
						var top = style.top;
						$('body').scrollTop(top - ($(window).height() / 2));
					}
				});

				// We might get this event from composeFab.
				scope.$on('newQueryMark', function() {
					scope.newQueryMark();
				});

				// Process all embedded votes
				$timeout(function() {
					element.find('[embed-vote-id]').each(function(index) {
						var $link = $(this);
						var pageAlias = $link.attr('embed-vote-id');
						arb.pageService.loadIntrasitePopover(pageAlias, {
							success: function(data, status) {
								var pageId = arb.pageService.pageMap[pageAlias].pageId;
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
			};

			// Everything is on a timeout to let MathJax do its thing
			$timeout(function() {
				MathJax.Hub.Queue(function() {
					// Wrap this in a try block, to make sure that any errors don't mess up MathJax
					try {
						processInlineEverything();
					} catch(err) {
						console.error(err);
					}
				});
			});
		},
	};
});
