'use strict';
// jscs:disable requireCamelCaseOrUpperCaseIdentifiers

import app from './angular.ts';
import {isIntIdValid} from './util.ts';

// Directive for the actual DOM elements which allows the user to edit a page.
app.directive('arbEditPage', function($location, $filter, $timeout, $interval, $http, $mdDialog, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/editPage.html'),
		scope: {
			pageId: '@',
			// Whether or not this edit page is embedded in some column, and should be
			// sized accordingly
			isEmbedded: '=',
			// True iff this is embedded inside a dialog
			insideDialog: '=',
			// Called when the user is done with the edit.
			doneFn: '&',
			// True if this is just a paragraph edit.
			paragraphEditMode: '=',
			paragraphIndex: '=',
			// True if we are just editing a summary.
			summaryEditMode: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.editMap[$scope.pageId];
			$scope.fullView = !$scope.isEmbedded && $mdMedia('gt-md');
			$scope.gtXSmallScreen = $mdMedia('gt-xs');
			$scope.gtSmallScreen = $mdMedia('gt-sm');
			$scope.gtMediumScreen = $mdMedia('gt-md');
			$scope.votesStartedAnonymous = $scope.page.votesAnonymous;
			// Set to true if the page has been changed (and therefore can be discarded)
			$scope.isPageDirty = $scope.page.isAutosave;
			// Set to true when the user can enter custom snapshot text.
			$scope.showSnapshotText = false;
			// Set when we can't allow the user to edit the text anymore (e.g. when
			// we notice that a new edit has been published.)
			$scope.freezeEdit = false;
			$scope.publishOptions = {
				isProposal: false,
			};
			$scope.isReviewingProposal = $scope.page.edit == $scope.page.proposalEditNum;

			$scope.justTextMode = $scope.paragraphEditMode || $scope.summaryEditMode;

			// Set up stuff for when we're just editing one paragraph
			if ($scope.paragraphEditMode) {
				$scope.originalFullText = $scope.page.text;
				var endIndex = $scope.page.text.indexOf('\n\n', $scope.paragraphIndex);
				if (endIndex == -1) {
					endIndex = undefined;
				}
				$scope.originalParagraphToEdit = $scope.page.text.substring($scope.paragraphIndex, endIndex);
				$scope.page.text = $scope.originalParagraphToEdit;
			}

			// Extract parentId from URL
			$scope.quickParentId = $location.search().parentId;
			$location.replace().search('parentId', undefined);

			// Modify the page
			$scope.page.text = arb.pageService.convertPageIdsToAliases($scope.page.text);

			$scope.getPublishText = function() {
				if ($scope.page.isDeleted) return 'Republish';
				if ($scope.isReviewingProposal) return 'Approve edit';
				if ($scope.publishOptions.isProposal) return 'Propose';
				if ($scope.page.permissions.edit.has) {
					if (!$scope.page.wasPublished && $scope.page.submitToDomainId != '0') {
						return 'Publish & Submit';
					}
					return 'Publish';
				}
				if ($scope.page.permissions.proposeEdit.has) return 'Propose';
				return '';
			};
			$scope.getPublishTooltipText = function() {
				if ($scope.page.isDeleted) return 'Republish this page';
				if ($scope.isReviewingProposal) return 'Approve the edit and publish this page';
				if ($scope.publishOptions.isProposal) return 'Propose an edit to this page';
				if ($scope.page.permissions.edit.has) {
					if (!$scope.page.wasPublished && $scope.page.submitToDomainId != '0') {
						return 'Publish the page and submit it to ' +
							arb.stateService.domainMap[$scope.page.submitToDomainId].alias;
					}
					return 'Make this version live';
				}
				if ($scope.page.permissions.proposeEdit.has) return 'Propose an edit to this page';
				return '';
			};

			// Whether the editor should be in advanced mode.
			$scope.showAdvancedMode = function() {
				return arb.userService.user.maxTrustLevel > 0 || arb.userService.user.showAdvancedEditorMode;
			};

			// Called when the advanced mode view is toggled
			$scope.advancedModeToggled = function() {
				arb.userService.updateSettings();
				if ($scope.showAdvancedMode()) {
					// Keep the user on the settings tab
					$scope.selectedTab = 3;
				}
			};

			$scope.getRelationshipTabIndex = function() {
				return $scope.showAdvancedMode() ? 2 : -1;
			};

			// Return true if we should show the publish button.
			$scope.isPublishButtonVisible = function() {
				return ($scope.selectedTab != $scope.getRelationshipTabIndex() || !$scope.page.wasPublished) && !$scope.freezeEdit;
			};

			// If the alias contains a subdomain, then remove it
			var periodIndex = $scope.page.alias.indexOf('.');
			if (periodIndex > 0) {
				$scope.page.alias = $scope.page.alias.substring(periodIndex + 1);
			}

			// Select correct tab
			$scope.selectedTab = ($scope.page.wasPublished || $scope.page.title.length > 0) ? 1 : 0;
			if ($location.search().tab) {
				$scope.selectedTab = $location.search().tab;
			}
			if ($scope.page.isComment()) {
				$scope.selectedTab = 1;
			}

			// See if the user is creating a claim
			var newClaimDomainId = $location.search().newClaimDomainId;
			if (newClaimDomainId) {
				$scope.page.hasVote = true;
				$scope.page.voteType = 'approval';
				if (arb.userService.user.canSubmitLinks(newClaimDomainId)) {
					$scope.page.editDomainId = newClaimDomainId;
				}
				$location.search('newClaimDomainId', undefined);
			}

			$scope.onParentChange = function() {
				var parentIds = arb.stateService.editMap[$scope.pageId].parentIds;
				if (parentIds.length != 1) {return;}

				var currentEditDomainId = arb.stateService.editMap[$scope.pageId].editDomainId;
				var currentEditDomain = arb.stateService.domainMap[currentEditDomainId];
				var parentDomainId = arb.stateService.pageMap[parentIds[0]].editDomainId;
				var parentDomain = arb.stateService.domainMap[parentDomainId];
				if (parentDomain.pageId == currentEditDomain.pageId) {return;}

				var confirmDialog = $mdDialog.confirm()
					.textContent('Would you like to change this page\'s domain (currently ' + currentEditDomain.alias + ') to match the parent\'s domain (' + parentDomain.alias +')?')
					.ok('Yes')
					.cancel('Cancel');

				$mdDialog.show(confirmDialog).then(function() {
					// If they clicked yes, then do the thing.
					$scope.page.editDomainId = parentDomainId;
					$scope.selectedTab = 3;
					$mdDialog.show(
						$mdDialog.alert()
							.textContent('This page\'s domain has been set to ' + parentDomain.alias + '. You will need to publish the page for this change to be saved.')
							.ok('ok')
					);
				});
			}

			// If this is a new page, and it had custom alias set, we can now set the title,
			// since this won't affect tab selection.
			if (!$scope.page.wasPublished && $scope.page.alias !== $scope.page.pageId) {
				$scope.page.title = $scope.page.alias.substring(0, 1).toUpperCase() + $scope.page.alias.substring(1);
				$scope.page.title = $scope.page.title.replace(/_/g, ' ');
			}

			// Set up markdown
			$timeout(function() {
				arb.markdownService.createEditConverter($scope, $scope.page.pageId);
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

			// Toggle in and out of preview when not in fullView
			$scope.inPreview = false;
			$scope.togglePreview = function(show) {
				$scope.inPreview = show;
			};

			$scope.shouldShowLivePreview = function() {
				if ($scope.isLivePreviewForceHidden) {
					return false;
				}
				if ($scope.isLivePreviewForceShown) {
					return true;
				}
				return $scope.fullView;
			}

			$scope.isLivePreviewForceHidden = false;
			$scope.isLivePreviewForceShown = false;
			$scope.forceHideLivePreview = function() {
				$scope.isLivePreviewForceHidden = true;
				$scope.isLivePreviewForceShown = false;
			};
			$scope.forceShowLivePreview = function() {
				$scope.isLivePreviewForceHidden = false;
				$scope.isLivePreviewForceShown = true;
			};
			$scope.toggleForceLivePreview = function() {
				if ($scope.shouldShowLivePreview()) {
					$scope.forceHideLivePreview();
				} else {
					$scope.forceShowLivePreview();
				}
			};

			// Return true if the preview should be shown
			$scope.isPreviewVisible = function() {
				if ($scope.otherDiff || $scope.sideEdit) {
					return false;
				}
				return $scope.shouldShowLivePreview() || $scope.inPreview;
			};

			// Setup all the settings
			$scope.isNormalEdit = !($scope.page.isSnapshot || $scope.page.isAutosave);

			// if this is a work in progress, load the saved edit summary, otherwise edit summary should be blank
			// (that is, do *not* load the edit summary for the previous edit that we're working off of)
			$scope.page.newEditSummary = $scope.isNormalEdit ? '' : $scope.page.editSummary;

			// Set up page types.
			if ($scope.page.isComment()) {
				$scope.pageTypes = {comment: 'Comment'};
			} else if ($scope.page.isWiki() || $scope.page.isQuestion()) {
				$scope.pageTypes = {wiki: 'Wiki', question: 'Question'};
			}

			// Set up domain options.
			$scope.domainOptions = arb.userService.getDomainOptions($scope.page);
			$scope.submitToDomainOptions = angular.extend({'0': '-'}, $scope.domainOptions);

			$scope.lockExists = $scope.page.lockedBy != '' && moment.utc($scope.page.lockedUntil).isAfter(moment.utc());
			$scope.lockedByAnother = $scope.lockExists && $scope.page.lockedBy !== arb.userService.user.id;

			// User reverts to an edit
			$scope.revertToEdit = function(editNum) {
				var data = {
					pageId: $scope.page.pageId,
					editNum: editNum,
				};
				$http({method: 'POST', url: '/revertPage/', data: JSON.stringify(data)})
				.success(function(data) {
					arb.urlService.goToUrl(arb.urlService.getPageUrl($scope.page.pageId));
				})
				.error(function(data) {
					$scope.addMessage('revert', 'Error reverting: ' + data, 'error');
				});
			};

			// Called when a user selects a question to merge into.
			$scope.mergeCandidate = undefined;
			$scope.selectedMergeQuestion = function(result) {
				if (result.pageId == $scope.page.pageId) return;
				$scope.mergeCandidate = arb.stateService.pageMap[result.pageId];
			};

			// Called when the user wants to merge this question.
			$scope.mergeQuestion = function() {
				var data = {
					questionId: $scope.page.pageId,
					intoQuestionId: $scope.mergeCandidate.pageId,
				};
				$http({method: 'POST', url: '/mergeQuestions/', data: JSON.stringify(data)})
				.success(function(data) {
					arb.urlService.goToUrl(arb.urlService.getPageUrl($scope.mergeCandidate.pageId));
				})
				.error(function(data) {
					$scope.addMessage('merge', 'Error merging: ' + data, 'error');
				});
			};

			$scope.moreRelationshipIds = undefined;
			$scope.loadMoreRelationships = function() {
				var data = {pageAlias: $scope.page.pageId};
				arb.stateService.postData('/json/moreRelationships/', data,
					function success(data) {
						$scope.moreRelationshipIds = data.result.moreRelationshipIds;
					}, function error(data) {
						$scope.addMessage('moreRelationships', 'Error loading more relationships: ' + data, 'error');
					}
				);
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
			if (!$scope.page.permissions.proposeEdit.has && !$scope.page.permissions.edit.has) {
				$scope.addMessage('editLevel', $scope.page.permissions.proposeEdit.reason, 'error', true);
			} else if (!$scope.page.permissions.edit.has) {
				$scope.addMessage('editLevel', $scope.page.permissions.edit.reason, 'warning');
			}
			// Check if you've loaded an edit that's not currently live
			if ($scope.page.edit !== $scope.page.currentEdit && $scope.isNormalEdit) {
				$scope.addMessage('nonLiveEdit', 'Currently looking at a non-live edit', 'warning');
			}
			if ($scope.page.wasPublished && $scope.page.isAutosave) {
				$scope.addMessage('nonLiveEdit', 'Loaded an autosave which was last updated ' +
					$filter('smartDateTime')(arb.stateService.primaryPage.editCreatedAt), 'warning');
			}
			if ($scope.page.wasPublished && $scope.page.isSnapshot) {
				$scope.addMessage('nonLiveEdit', 'Loaded a snapshot which was last updated ' +
					$filter('smartDateTime')(arb.stateService.primaryPage.editCreatedAt), 'warning');
			}
			// Check if we loaded a live edit, but the user has a draft
			if ($scope.page.wasPublished && $scope.page.hasDraft && $scope.isNormalEdit) {
				//$scope.addMessage('hasDraft', 'You have an unpublished draft that\'s more recent than this edit', 'error');
			}
			if ($scope.page.isDeleted) {
				$scope.addMessage('deletedPage', 'This page was deleted. Republishing will restore it.', 'warning');
			}

			// =========== Autosaving / snapshotting / publishing stuff ==============
			$scope.autosaving = false;
			$scope.publishing = false;
			$scope.snapshotting = false;

			var prevEditPageData = undefined;
			$timeout(function() {
				// If we loaded an edit that's not based off of the current edit, freeze!
				// Note: do this before we compute prevEditPageData, to make sure autosave goes through.
				if ($scope.page.wasPublished) {
					var isThisEditLive = $scope.page.edit == $scope.page.currentEdit;
					if (!isThisEditLive && $scope.page.prevEdit != $scope.page.currentEdit) {
						$scope.freezeEdit = true;
						$scope.savePage(false, true);
						var message = 'A new version was published. To prevent edit conflicts, please refresh the page to see it. (A snapshot of your current state has been saved.)';
						$scope.addMessage('obsoleteEdit', message, 'error');
					}
				}
				// Compute prevEditPageData, so we don't fire off autosave when there were
				// no changes made.
				prevEditPageData = computeAutosaveData();
			});

			// Helper function for savePage. Computes the data to submit via AJAX.
			var computeAutosaveData = function() {
				// We have to pull the text from textarea directly, because if it's changed by
				// Markdown library, AngularJS doesn't notice it and page.text isn't updated.
				var enteredText = ($('#wmd-input' + $scope.pageId)[0] as HTMLTextAreaElement).value;

				if ($scope.paragraphEditMode) {
					$scope.page.text = $scope.originalFullText.replace($scope.originalParagraphToEdit, enteredText);
				} else {
					$scope.page.text = enteredText;
				}
				var saveData = arb.editService.computeSavePageData($scope.page);

				$scope.page.text = enteredText;

				return saveData;
			};
			// Save the current page.
			// callback is called with the error (or undefined on success)
			$scope.savePage = function(isAutosave, isSnapshot, callback, autosaveNotPerformedCallback) {
				var data = computeAutosaveData();
				if (!isAutosave || JSON.stringify(data) !== JSON.stringify(prevEditPageData)) {
					prevEditPageData = $.extend({}, data);
					data.isAutosave = isAutosave;
					data.isSnapshot = isSnapshot;
					data.isProposal = !$scope.page.permissions.edit.has || $scope.publishOptions.isProposal;
					// Send the data to the server.
					// TODO: if the call takes too long, we should show a warning.
					$http({method: 'POST', url: '/editPage/', data: JSON.stringify(data)})
						.success(function(returnedData) {
							var newEdit = returnedData.result.obsoleteEdit;
							if (newEdit) {
								// A new edit has been published while the user has been editing.
								$scope.freezeEdit = true;
								var message = 'User (id=' + newEdit.editCreatorId + ') published a new version. To prevent edit conflicts, please refresh the page to see it. (A snapshot of your current state has been saved.)';
								$scope.addMessage('obsoleteEdit', message, 'error');
							}
							if (isAutosave) {
								// Refresh the lock
								$scope.page.lockedUntil = moment.utc().add(30, 'm').format('YYYY-MM-DD HH:mm:ss');
							}
							$scope.isPageDirty = isAutosave;
							$scope.isReviewingProposal = false;
							arb.analyticsService.reportEventToHeapAndMixpanel('Save', {
								pageId: data.pageId,
								textLength: data.text.length,
								isAutoave: data.isAutosave,
								isSnapshot: data.isSnapshot,
								isProposal: data.isProposal,
							});
	
							if (callback) callback();
						})
						.error(function(returnedData) {
							if (callback) callback(returnedData);
						});
				} else {
					if (autosaveNotPerformedCallback) autosaveNotPerformedCallback();
				}
			};

			// Call the doneFn callback after the page has been fully published.
			var publishPageDone = function() {
				if (!$scope.page.permissions.edit.has || $scope.publishOptions.isProposal) {
					arb.popupService.showToast({text: 'Your proposal has been submitted.'});
				}
				// Report to analytics
				if (!$scope.page.isComment()) {
					if (!$scope.page.permissions.edit.has) {
						arb.analyticsService.reportPublishAction('propose edit', $scope.pageId, $scope.page.text.length);
					} else if ($scope.page.wasPublished) {
						arb.analyticsService.reportPublishAction('edit', $scope.pageId, $scope.page.text.length);
					} else {
						arb.analyticsService.reportPublishAction('new', $scope.pageId, $scope.page.text.length);
					}
				}
				$scope.doneFn({result: {
					pageId: $scope.page.pageId,
					alias: $scope.page.alias,
				}});
			};
			// Called when user clicks Publish button
			$scope.publishPage = function() {
				if ($scope.isReviewingProposal) {
					$scope.processEditProposal(false);
					return;
				}

				$scope.publishing = true;
				$scope.savePageInfo(function(error) {
					if (error) {
						$scope.publishing = false;
						$scope.addMessage('publish', 'Publishing failed: ' + error, 'error');
						return;
					}
					$scope.savePage(false, false, function(error) {
						$scope.publishing = false;
						if (error) {
							$scope.addMessage('publish', 'Publishing failed: ' + error, 'error');
						} else if ($location.search().markId) {
							// Update the mark as resolved
							arb.markService.resolveMark({
								markId: $location.search().markId,
								resolvedPageId: $scope.pageId,
							}, function success() {
								arb.popupService.showToast({text: 'You resolved the query mark.'});
								publishPageDone();
							}, function error() {
								publishPageDone();
							});
						} else {
							publishPageDone();
						}
					});
				});
			};

			// Process Snapshot button click
			$scope.snapshotPage = function() {
				$scope.showSnapshotText = false;
				$scope.snapshotting = true;
				$scope.successfulSnapshot = false;
				$scope.savePage(false, true, function(error) {
					$scope.snapshotting = false;
					if (error) {
						$scope.addMessage('snapshot', 'Snapshot failed: ' + error, 'error');
					} else {
						$scope.addMessage('snapshot', 'Snapshot saved!', 'info');
					}
				});
			};

			// Process Discard button click.
			$scope.discardPage = function() {
				var cont = function() {
					if (!$scope.doneFn) return;
					$scope.doneFn({result: {
						pageId: $scope.page.pageId,
						alias: $scope.page.alias,
						discard: true,
					}});
				};
				arb.pageService.discardPage($scope.page.pageId, cont, cont);
			};

			// Return true iff we should skip asking whether the user wants to discard the page.
			$scope.shouldSkipDiscard = function() {
				return !($scope.page.title.length > 0 || $scope.page.clickbait.length > 0 || $scope.page.text.length > 0);
			};

			// Process Delete button click.
			$scope.deletePage = function() {
				arb.pageService.deletePage($scope.page.pageId, function() {
					if ($scope.doneFn) {
						$scope.doneFn({result: {
							pageId: $scope.page.pageId,
							alias: $scope.page.alias,
							discard: true,
							deletedPage: true,
						}});
					}
				}, function(data) {
					$scope.addMessage('delete', 'Error deleting page: ' + data, 'error');
				});
			};

			// Set up autosaving.
			$scope.successfulAutosave = false;
			$scope.autosaveFunc = function() {
				if ($scope.autosaving) return;
				$scope.autosaving = true;
				$scope.successfulAutosave = false;
				$scope.savePage(true, false, function(error) {
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

			// =========== Find similar pages ==============
			var searchingForSimilarPages = false;
			var prevSimilarData = {};
			$scope.similarPages = [];
			var findSimilarFunc = function() {
				if (searchingForSimilarPages) return;
				if ($scope.selectedTab != 0) return;
				if ($scope.page.wasPublished) return;
				if ($scope.page.isComment()) return;

				var data = {
					title: $scope.page.title,
					// Cutting off text at the last 4k characters, so Elastic doesn't choke
					text: $scope.page.text.length > 4000 ? $scope.page.text.slice(-4000) : $scope.page.text,
					clickbait: $scope.page.clickbait,
				};
				if (JSON.stringify(data) == JSON.stringify(prevSimilarData)) {
					return;
				}

				searchingForSimilarPages = true;
				prevSimilarData = data;
				arb.autocompleteService.findSimilarPages(data, function(data) {
					searchingForSimilarPages = false;
					$scope.similarPages.length = 0;
					for (var n = 0; n < data.length; n++) {
						var pageId = data[n].pageId;
						if (pageId === $scope.page.pageId) continue;
						$scope.similarPages.push({pageId: pageId, score: data[n].score});
					}
				});
			};
			var similarInterval = $interval(findSimilarFunc, 500);
			$scope.$on('$destroy', function() {
				$interval.cancel(similarInterval);
			});

			// =========== Show diff between edits ==============
			// otherDiff stores the edit we load for diffing.
			$scope.otherDiff = undefined;
			$scope.diffHtml = undefined;
			$scope.diffExpanded = false;
			// Refresh the diff edit text.
			$scope.refreshDiff = function() {
				$scope.diffHtml = arb.diffService.getDiffHtml($scope.otherDiff, $scope.page, $scope.diffExpanded);
			};
			$scope.toggleExpandDiff = function() {
				$scope.diffExpanded = !$scope.diffExpanded;
				$scope.refreshDiff();
			};
			// Process click event for diffing edits
			$scope.showDiff = function(editNum) {
				// Load the edit from the server
				arb.pageService.loadEdit({
					pageAlias: $scope.page.pageId,
					specificEdit: editNum,
					skipProcessDataStep: true,
					convertPageIdsToAliases: true,
					success: function(data, status) {
						$scope.otherDiff = data.edits[$scope.page.pageId];
						$scope.refreshDiff();
						$scope.selectedTab = 1;
					},
				});
			};
			$scope.hideDiff = function() {
				$scope.otherDiff = undefined;
				$scope.diffHtml = undefined;
			};

			if ($scope.page.proposalEditNum > 0) {
				$timeout(function() {
					$scope.showDiff($scope.page.prevEdit);
					arb.popupService.showToast({text: 'There is a pending edit proposal. Please review.'});
				});
			}

			// =========== Side by side edit ==============
			// sideEdit stores the edit we loaded to show on the right-hand side
			$scope.sideEdit = undefined;
			// Process click event for showing a side edit
			$scope.loadSideEdit = function(editNum) {
				// Load the edit from the server
				arb.pageService.loadEdit({
					pageAlias: $scope.page.pageId,
					specificEdit: editNum,
					skipProcessDataStep: true,
					convertPageIdsToAliases: true,
					success: function(data, status) {
						$scope.sideEdit = data.edits[$scope.page.pageId];
						$scope.selectedTab = 1;
					},
				});
			};
			$scope.hideSideEdit = function() {
				$scope.sideEdit = undefined;
			};

			// =========== Search strings ==============
			$scope.addSearchStringData = {
				pageId: $scope.pageId,
				text: '',
			};
			$scope.addSearchString = function() {
				console.log($scope.addSearchStringData);
				$http({method: 'POST', url: '/newSearchString/', data: JSON.stringify($scope.addSearchStringData)})
				.success(function(data) {
					$scope.page.searchStrings[data.result.searchStringId] = $scope.addSearchStringData.text;
					$scope.addSearchStringData.text = '';
				})
				.error(function(data) {
					$scope.addMessage('addSearchString', 'Error adding a search string: ' + data, 'error');
				});
			};

			$scope.deleteSearchString = function(id) {
				var postData = {
					id: id,
				};
				$http({method: 'POST', url: '/deleteSearchString/', data: JSON.stringify(postData)})
				.error(function(data) {
					$scope.addMessage('deleteSearchString', 'Error deleting a search string: ' + data, 'error');
				});
				delete $scope.page.searchStrings[id];
			};

			// Save the page info.
			// callback is called with a potential error message when the server replies
			$scope.savePageInfo = function(callback) {
				arb.pageService.savePageInfo($scope.page, callback);
			};

			// Return true iff any of the pageInfo values changed.
			$scope.pageInfoChanged = function() {
				if (!$scope.page.wasPublished) return false;
				// TODO: the page won't be in the pageMap if it's deleted. Ideally we have a better
				// workaround for this.
				if (!($scope.pageId in arb.stateService.pageMap)) return true;
				var originalPageInfo = arb.stateService.pageMap[$scope.pageId].getPageInfo();
				var newPageInfo = $scope.page.getPageInfo();
				var changed = !angular.equals(originalPageInfo, newPageInfo);
				$scope.isReviewingProposal = $scope.isReviewingProposal && !changed;
				return changed;
			};

			$scope.$watch(function() {
				return $scope.isReviewingProposal;
			}, function() {
				if (!$scope.isReviewingProposal) {
					$location.search('edit', undefined);
				}
			});

			// Called when the user decided to accept or dismiss the edit proposal.
			$scope.processEditProposal = function(dismiss) {
				for (var n = 0; n < $scope.page.changeLogs.length; n++) {
					var changeLog = $scope.page.changeLogs[n];
					if (changeLog.type == 'newEditProposal' && changeLog.edit == $scope.page.proposalEditNum) {
						arb.pageService.approveEditProposal(changeLog, dismiss);
						if (dismiss) {
							$scope.isReviewingProposal = false;
						} else {
							publishPageDone();
						}
						break;
					}
				}
			};

			$scope.showPublishingOptionsPanel = false;
			$scope.togglePublishingOptionsPanel = function() {
				$scope.showPublishingOptionsPanel = !$scope.showPublishingOptionsPanel;
			};
		},
		link: function(scope: any, element, attrs) {
			// Do autosave every so often.
			var autosaveInterval = $interval(scope.autosaveFunc, 5000);
			// When this element is destroyed, do one last autosave just in case.
			element.on('$destroy', function() {
				$interval.cancel(autosaveInterval);
				scope.savePage(true, false, function() {});
			});

			scope.toggleSnapshotting = function() {
				scope.showSnapshotText = !scope.showSnapshotText;
				if (scope.showSnapshotText) {
					scope.page.snapshotText = 'My snapshot (' + moment().format('YYYY-MM-DD h:mm a') + ')';
					$timeout(function() {
						element.find('.snapshot-text').focus();
					});
				}
			};

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
						$other.on('scroll', syncScroll);
					}, 10);*/
				};
				$divs.on('scroll', syncScroll);

				if (scope.page.isComment()) {
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
				$markdownToolbar.on('showInsertLink', function(event, callback, isAtMention) {
					scope.showInsertLink = true;
					scope.searchGroups = isAtMention;
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
						templateUrl: versionUrl('static/html/editPageDialog.html'),
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

				// Create a dialog for (resuming) editing a new claim
				var resumeClaimPageId = undefined;
				$markdownToolbar.on('showNewClaimDialog', function(event, callback, title) {
					$mdDialog.show({
						templateUrl: versionUrl('static/html/editClaimDialog.html'),
						controller: 'EditClaimDialogController',
						autoWrap: false,
						targetEvent: event,
						locals: {
							resumePageId: resumeClaimPageId,
							title: title,
							originalPage: scope.page,
						},
					})
					.then(function(result) {
						resumeClaimPageId = undefined;
						if (result.hidden) {
							resumeClaimPageId = result.pageId;
							callback(undefined);
						} else if (result.discard) {
							callback(undefined);
						} else {
							callback(result.pageId);
						}
					}, function() {
						callback(undefined);
					});
					return false;
				});
			});
		},
	};
});
