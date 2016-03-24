'use strict';
// jscs:disable requireCamelCaseOrUpperCaseIdentifiers

// Directive for the actual DOM elements which allows the user to edit a page.
app.directive('arbEditPage', function($location, $filter, $timeout, $interval, $http, $mdDialog, $mdMedia, pageService, userService, autocompleteService, markdownService) {
	return {
		templateUrl: 'static/html/editPage.html',
		scope: {
			pageId: '@',
			// Whether or not this edit page is embedded in some column, and should be
			// sized accordingly
			isEmbedded: '=',
			// True iff this is embedded inside a dialog
			insideDialog: '=',
			// Called when the user is done with the edit.
			doneFn: '&',
		},
		controller: function($scope) {
			$scope.userService = userService;
			$scope.pageService = pageService;
			$scope.page = pageService.editMap[$scope.pageId];
			$scope.fullView = !$scope.isEmbedded && $mdMedia('gt-md');
			$scope.gtXSmallScreen = $mdMedia('gt-xs');
			$scope.gtSmallScreen = $mdMedia('gt-sm');
			$scope.gtMediumScreen = $mdMedia('gt-md');
			$scope.liveEditUrl = pageService.getEditPageUrl($scope.page.pageId) + $scope.page.currentEditNum;
			$scope.isPageDirty = $scope.page.isAutosave;

			// If the alias contains a subdomain, then remove it
			var periodIndex = $scope.page.alias.indexOf('.');
			if (periodIndex > 0) {
				$scope.page.alias = $scope.page.alias.substring(periodIndex + 1);
			}

			// Return true if we should be using a table layout (so we can stack right
			// and left columns vertically)
			$scope.useTableLayout = function() {
				return !$scope.fullView && !$scope.inPreview && !$scope.otherDiff;
			};

			// Select correct tab
			$scope.selectedTab = ($scope.page.wasPublished || $scope.page.title.length > 0) ? 1 : 0;
			if ($location.search().tab) {
				$scope.selectedTab = $location.search().tab;
			}
			if ($scope.page.isComment()) {
				$scope.selectedTab = 1;
			}

			// Set up markdown
			$timeout(function() {
				var $wmdPreview = $('#wmd-preview' + $scope.page.pageId);
				// Initialize pagedown
				markdownService.createEditConverter($scope.page.pageId, function(refreshFunc) {
					$timeout(function() {
						markdownService.processLinks($scope, $wmdPreview, refreshFunc);
					});
				});
			});

			// Called when user selects a page from insert link input
			$scope.insertLinkSelect = function(result) {
				if (!$scope.insertLinkCallback) return;
				var result = result;
				$scope.showInsertLink = false;
				$timeout(function() {
					$scope.insertLinkCallback(result ? result.alias : undefined);
					$scope.insertLinkCallback = undefined;
				});
			};

			// Lens sort listeners (using ng-sortable library)
			$scope.page.lensIds.sort(function(a, b) {
				return pageService.pageMap[a].lensIndex - pageService.pageMap[b].lensIndex;
			});
			$scope.lensSortListeners = {
				orderChanged: function(event) {
					var data = {pageId: $scope.page.pageId, orderMap: {}};
					for (var n = 0; n < $scope.page.lensIds.length; n++) {
						var pageId = $scope.page.lensIds[n];
						pageService.pageMap[pageId].lensIndex = n + 1;
						data.orderMap[pageId] = n + 1;
					}
					$http({method: 'POST', url: '/updateLensOrder/', data: JSON.stringify(data)})
					.error(function(data) {
						$scope.addMessage('lensOrder', 'Error updating lens order: ' + data, 'error');
					});
				},
			};

			// Toggle in and out of preview when not in fullView
			$scope.inPreview = false;
			$scope.togglePreview = function(show) {
				$scope.inPreview = show;
			};

			// Setup all the settings
			$scope.isWiki = $scope.page.isWiki();
			$scope.isQuestion = $scope.page.isQuestion();
			$scope.isAnswer = $scope.page.isAnswer();
			$scope.isComment = $scope.page.isComment();
			$scope.isLens = $scope.page.isLens();
			$scope.isGroup = $scope.page.isGroup() || $scope.page.isDomain();
			$scope.forceExpandSimilarPagesCount = $scope.isQuestion ? 5 : 10;

			// Set up page types.
			if ($scope.isQuestion) {
				$scope.pageTypes = {question: 'Question'};
			} else if ($scope.isAnswer) {
				$scope.pageTypes = {answer: 'Answer'};
			} else if ($scope.isComment) {
				$scope.pageTypes = {comment: 'Comment'};
			} else if ($scope.isWiki) {
				$scope.pageTypes = {wiki: 'Wiki Page'};
			} else if ($scope.isLens) {
				$scope.pageTypes = {lens: 'Lens Page'};
			}

			// Set up group names.
			var groupIds = userService.user.groupIds;
			$scope.groupOptions = {'': '-'};
			if (groupIds) {
				for (var i in groupIds) {
					var groupId = groupIds[i];
					var groupName = pageService.pageMap[groupId].title;
					$scope.groupOptions[groupId] = groupName;
				}
			}

			// Set up sort types.
			$scope.sortTypes = {
				likes: 'By Likes',
				recentFirst: 'Recent First',
				oldestFirst: 'Oldest First',
				alphabetical: 'Alphabetically',
			};

			// Set up vote types.
			$scope.voteTypes = {
				'': '-',
				probability: 'Probability',
				approval: 'Approval',
			};

			$scope.lockExists = $scope.page.lockedBy != '' && moment.utc($scope.page.lockedUntil).isAfter(moment.utc());
			$scope.lockedByAnother = $scope.lockExists && $scope.page.lockedBy !== userService.user.id;

			$scope.convertPageIdsToAliases = function(textToConvert) {
				// Convert all links with pageIds to alias links.
				return textToConvert.replace(complexLinkRegexp, function(whole, prefix, text, alias) {
					var page = pageService.pageMap[alias];
					if (page) {
						return prefix + '[' + text + '](' + page.alias + ')';
					}
					return whole;
					/*}).replace(voteEmbedRegexp, function (whole, prefix, alias) {
						var page = pageService.pageMap[alias];
						if (page) {
						return prefix + "[vote: " + page.alias + "]";
						}
						return whole;*/
				}).replace(forwardLinkRegexp, function(whole, prefix, alias, text) {
					var page = pageService.pageMap[alias];
					if (page) {
						return prefix + '[' + page.alias + ' ' + text + ']';
					}
					return whole;
				}).replace(simpleLinkRegexp, function(whole, prefix, alias) {
					if (alias.substring(0,1) == '-') {
						var page = pageService.pageMap[alias.substring(1)];
						if (page) {
							return prefix + '[-' + page.alias + ']';
						}
					} else {
						var page = pageService.pageMap[alias];
						if (page) {
							return prefix + '[' + page.alias + ']';
						}
					}
					return whole;
				}).replace(atAliasRegexp, function(whole, prefix, alias) {
					var page = pageService.pageMap[alias];
					if (page) {
						return prefix + '[@' + page.alias + ']';
					}
					return whole;
				});
			};

			$scope.page.text = $scope.convertPageIdsToAliases($scope.page.text);
			if ($scope.isLens) {
				$scope.page.title = $scope.page.lensTitle();
			}

			// User reverts to an edit
			$scope.revertToEdit = function(editNum) {
				var data = {
					pageId: $scope.page.pageId,
					editNum: editNum,
				};
				$http({method: 'POST', url: '/revertPage/', data: JSON.stringify(data)})
				.success(function(data) {
					$location.url(pageService.getPageUrl($scope.page.pageId));
				})
				.error(function(data) {
					$scope.addMessage('revert', 'Error reverting: ' + data, 'error');
				});
			};

			// =========== Error, warning, and info management system ==============
			$scope.messages = {};
			$scope.addMessage = function(key, text, type, permanent) {
				$scope.messages[key] = {text: text, type: type, permanent: permanent};
			};
			$scope.deleteMessage = function(key) {
				delete $scope.messages[key];
			};

			$scope.hideMessage = function(event) {
				$(event.currentTarget).closest('md-list-item').hide();
			};

			// Check if the user can edit this page
			if ($scope.page.wasPublished) {
				var editLevel = $scope.page.getEditLevel();
				if (editLevel === 'admin') {
					$scope.addMessage('editLevel', 'Enforcing admin priviledges', 'warning');
				} else if (editLevel === 'comment') {
					$scope.addMessage('editLevel', 'Can\'t edit a comment you didn\'t create', 'error', true);
				} else if (editLevel === '') {
				} else {
					$scope.addMessage('editLevel', 'You don\'t have enough karma to edit this page. Required: ' +
						$scope.page.getEditLevel(), 'error', true);
				}
			}
			// Check group permissions
			if ($scope.page.editGroupId !== '' && !($scope.page.editGroupId in $scope.groupOptions)) {
				$scope.addMessage('editGroup', 'You need to be part of ' +
					pageService.pageMap[$scope.page.editGroupId].title + ' group to edit this page', 'error', true);
			}
			// Check if you've loaded an edit that's not currently live
			if ($scope.page.edit !== $scope.page.currentEditNum && !$scope.page.isAutosave && !$scope.page.isSnapshot) {
				$scope.addMessage('nonLiveEdit', 'Currently looking at a non-live edit', 'warning');
			}
			if ($scope.page.wasPublished && $scope.page.isAutosave) {
				$scope.addMessage('nonLiveEdit', 'Restored an autosave which was last updated ' +
					$filter('relativeDateTime')(pageService.primaryPage.createdAt), 'warning');
			}
			if ($scope.page.wasPublished && $scope.page.isSnapshot) {
				$scope.addMessage('nonLiveEdit', 'Restored a snapshot which was last updated ' +
					$filter('relativeDateTime')(pageService.primaryPage.createdAt), 'warning');
			}

			// =========== Autosaving / snapshotting / publishing stuff ==============
			$scope.autosaving = false;
			$scope.publishing = false;
			$scope.snapshotting = false;

			var prevEditPageData = undefined;
			$timeout(function() {
				// Compute prevEditPageData, so we don't fire off autosave when there were
				// no changes made.
				prevEditPageData = computeAutosaveData();
			});

			// Helper function for savePage. Computes the data to submit via AJAX.
			var computeAutosaveData = function() {
				// We have to pull the text from textarea directly, because if it's changed by
				// Markdown library, AngularJS doesn't notice it and page.text isn't updated.
				$scope.page.text = $('#wmd-input' + $scope.pageId)[0].value;
				var data = {
					pageId: $scope.pageId,
					prevEdit: $scope.page.prevEdit,
					title: $scope.page.title,
					clickbait: $scope.page.clickbait,
					text: $scope.page.text,
				};
				if ($scope.page.anchorContext) {
					data.anchorContext = $scope.page.anchorContext;
					data.anchorText = $scope.page.anchorText;
					data.anchorOffset = $scope.page.anchorOffset;
				}
				return data;
			};
			// Save the current page.
			// callback is called with the error (or undefined on success)
			var savePage = function(isAutosave, isSnapshot, callback, autosaveNotPerformedCallback) {
				var data = computeAutosaveData();
				if (!isAutosave || JSON.stringify(data) !== JSON.stringify(prevEditPageData)) {
					shouldFindSimilar = true;
					prevEditPageData = $.extend({}, data);
					data.isAutosave = isAutosave;
					data.isSnapshot = isSnapshot;
					data.isEditorComment = $scope.page.isEditorComment;
					// Send the data to the server.
					// TODO: if the call takes too long, we should show a warning.
					$http({method: 'POST', url: '/editPage/', data: JSON.stringify(data)})
					.success(function(data) {
						if (isAutosave) {
							// Refresh the lock
							$scope.page.lockedUntil = moment.utc().add(30, 'm').format('YYYY-MM-DD HH:mm:ss');
						}
						$scope.isPageDirty = isAutosave;
						callback();
					})
					.error(function(data) {
						callback(data);
					});
				} else {
					if (autosaveNotPerformedCallback) autosaveNotPerformedCallback();
				}
			};

			// Called when user clicks Publish button
			$scope.publishPage = function() {
				$scope.publishing = true;
				$scope.savePageInfo(function(error) {
					if (error) {
						$scope.publishing = false;
						$scope.addMessage('publish', 'Publishing failed: ' + error, 'error');
					} else {
						savePage(false, false, function(error) {
							$scope.publishing = false;
							if (error) {
								$scope.addMessage('publish', 'Publishing failed: ' + error, 'error');
							} else {
								$scope.doneFn({result: {
									pageId: $scope.page.pageId,
									alias: $scope.page.alias
								}});
							}
						});
					}
				});
			};

			// Process Snapshot button click
			$scope.snapshotPage = function() {
				$scope.snapshotting = true;
				$scope.successfulSnapshot = false;
				savePage(false, true, function(error) {
					$scope.snapshotting = false;
					if (error) {
						$scope.addMessage('snapshot', 'Snapshot failed: ' + error, 'error');
					} else {
						$scope.addMessage('snapshot', 'Snapshot saved!', 'info');
					}
				});
			};

			// Process Discard button click.
			$scope.discardPage = function(continueEditing) {
				var cont = function() {
					if (continueEditing) {
						window.location.href = $scope.liveEditUrl;
					} else if ($scope.doneFn) {
						$scope.doneFn({result: {
							pageId: $scope.page.pageId,
							alias: $scope.page.alias,
							discard: true,
						}});
					}
				};
				pageService.discardPage($scope.page.pageId, cont, cont);
			};

			// Process Delete button click.
			$scope.deletePage = function() {
				pageService.deletePage($scope.page.pageId, function() {
					if ($scope.doneFn) {
						$scope.doneFn({result: {
							pageId: $scope.page.pageId,
							alias: $scope.page.alias,
							discard: true,
						}});
					}
				}, function(data) {
					$scope.addMessage('delete', 'Error deleting page: ' + data, 'error');
				});
			};

			// Set up autosaving.
			$scope.successfulAutosave = false;
			var autosaveFunc = function() {
				if ($scope.autosaving) return;
				$scope.autosaving = true;
				$scope.successfulAutosave = false;
				savePage(true, false, function(error) {
					$scope.autosaving = false;
					$scope.successfulAutosave = !error;
					if (error) {
						$scope.addMessage('autosave', 'Autosave error: ' + error, 'error', true);
					} else {
						$scope.deleteMessage('autosave');
					}
				}, function() {
					$scope.autosaving = false;
				});
			};
			var autosaveInterval = $interval(autosaveFunc, 5000);
			$scope.$on('$destroy', function() {
				$interval.cancel(autosaveInterval);
				// Autosave just in case.
				savePage(true, false, function() {});
			});

			// =========== Find similar pages ==============
			var shouldFindSimilar = false;
			$scope.expandSimilarPages = false;
			$scope.toggleSimilarPages = function(show) {
				$scope.expandSimilarPages = show;
			};
			$scope.similarPages = [];
			var findSimilarFunc = function() {
				if ($scope.page.wasPublished) return;
				if (!shouldFindSimilar || $scope.isComment) return;
				// Don't search, if we don't have the first tab selected
				if (!$scope.isQuestion && $scope.selectedTab != 0) return;
				shouldFindSimilar = false;
				var data = {
					title: $scope.page.title,
					// Cutting off text at the last (arbitrary) 4k characters, so Elastic doesn't choke
					text: $scope.page.text.length > 4000 ? $scope.page.text.slice(-4000) : $scope.page.text,
					clickbait: $scope.page.clickbait,
					pageType: $scope.page.type,
				};
				autocompleteService.findSimilarPages(data, function(data) {
					$scope.similarPages.length = 0;
					for (var n = 0; n < data.length; n++) {
						var pageId = data[n].pageId;
						if (pageId === $scope.page.pageId) continue;
						$scope.similarPages.push({pageId: pageId, score: data[n].score});
					}
				});
			};
			var similarInterval = $interval(findSimilarFunc, $scope.isQuestion ? 10000 : 3000);
			$scope.$on('$destroy', function() {
				$interval.cancel(similarInterval);
			});

			// =========== Show diff between edits ==============
			// otherDiff stores the edit we load for diffing.
			$scope.otherDiff = undefined;
			$scope.diffHtml = undefined;
			// Refresh the diff edit text.
			$scope.refreshDiff = function() {
				var dmp = new diff_match_patch(); // jscs:ignore requireCapitalizedConstructors
				var diffs = dmp.diff_main($scope.otherDiff.text, $scope.page.text);
				dmp.diff_cleanupSemantic(diffs);
				$scope.diffHtml = dmp.diff_prettyHtml(diffs).replace(/&para;/g, '');
			};
			// Process click event for diffing edits.
			$scope.showDiff = function(editNum) {
				// Load the edit from the server.
				pageService.loadEdit({
					pageAlias: $scope.page.pageId,
					specificEdit: editNum,
					skipProcessDataStep: true,
					success: function(data, status) {
						$scope.otherDiff = data[$scope.page.pageId];
						$scope.otherDiff.text = $scope.convertPageIdsToAliases($scope.otherDiff.text);
						$scope.refreshDiff();
						$scope.selectedTab = 1;
					},
				});
			};
			$scope.hideDiff = function() {
				$scope.otherDiff = undefined;
				$scope.diffHtml = undefined;
			};

			// Save the page info.
			// callback is called with a potential error message when the server replies
			$scope.savePageInfo = function(callback) {
				pageService.savePageInfo($scope.page, callback);
			};

			// Focus on the default element in the tab
			$scope.focusDefaultTabElement = function() {
				var activeTab = $scope.selectedTab;
				var defaultTabItem = $('[default-tab-item|=' + activeTab + ']');
				if (0 in defaultTabItem) {
					defaultTabItem[0].focus();
				}
			};

			// Change selected tab manually
			$scope.changeTab = function(activeTab) {
				if (activeTab < 0) activeTab = 4;
				if (activeTab >= 5) activeTab = 0;
				$scope.selectedTab = activeTab;
				var tabList = $('md-tab-item');
				if (activeTab in tabList) {
					tabList[activeTab].focus();
				}
			};

			$scope.handleKeyPress = function(event) {
				if (event.ctrlKey && event.keyCode == 38) {
					// Ctrl + up
					$scope.changeTab($scope.selectedTab - 1);
					setTimeout($scope.focusDefaultTabElement, 1000);
				} else if (event.ctrlKey && event.keyCode == 40) {
					// Ctrl + down
					$scope.changeTab($scope.selectedTab + 1);
					setTimeout($scope.focusDefaultTabElement, 1000);
				}
			};
		},
		link: function(scope, element, attrs) {
			$timeout(function() {
				// Synchronize scrolling between the textarea and the preview.
				var $wmdInput = element.find('.wmd-input');
				var $divs = $wmdInput.add(element.find('.preview-area'));
				var syncScroll = function(event) {
					var $other = $divs.not(this).off('scroll');
					var other = $other.get(0);
					var percentage = this.scrollTop / (this.scrollHeight - this.offsetHeight);
					other.scrollTop = Math.round(percentage * (other.scrollHeight - other.offsetHeight));
					// Firefox workaround. Rebinding without delay isn't enough.
					/*setTimeout(function() {
						$other.on("scroll", syncScroll);
					}, 10);*/
				};
				if (scope.fullView) {
					$divs.on('scroll', syncScroll);
				}

				if (scope.isComment) {
					// Scroll to show the entire edit page element and focus on the input.
					var editorTop = element.offset().top + element.height() - $(window).height() + 80;
					$('html, body').animate({
						scrollTop: Math.max($('body').scrollTop(), editorTop),
					}, 1000);
					$wmdInput.focus();
				}

				// Listen to events from Markdown.Editor
				var $markdownToolbar = element.find('.wmd-button-bar');
				var blurHooked = false;
				// Show autocomplete for inserting an intrasite link
				$markdownToolbar.on('showInsertLink', function(event, callback) {
					scope.showInsertLink = true;
					scope.insertLinkCallback = callback;
					// NOTE: not sure why, but we need two timeouts here
					$timeout(function() { $timeout(function() {
						var $input = element.find('.insert-autocomplete').find('input').focus();
					}); });
				});

				// Create a dialog for (resuming) editing a new page
				var resumePageId = undefined;
				$markdownToolbar.on('showNewPageDialog', function(event, callback, newPageType) {
					var parentIds = [];
					if (newPageType === 'child') {
						parentIds = [scope.page.pageId];
					} else if (newPageType === 'sibling') {
						parentIds = scope.page.parentIds;
					}
					$mdDialog.show({
						templateUrl: 'static/html/editPageDialog.html',
						controller: 'EditPageDialogController',
						autoWrap: false,
						targetEvent: event,
						locals: {
							resumePageId: resumePageId,
							parentIds: parentIds,
						},
					})
					.then(function(result) {
						resumePageId = undefined;
						if (result.hidden) {
							resumePageId = result.pageId;
						} else if (result.discard) {
							callback(undefined);
						} else {
							callback(result.alias);
						}
					});
					return false;
				});
			});
		},
	};
});
