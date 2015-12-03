"use strict";

// Directive for the actual DOM elements which allows the user to edit a page.
app.directive("arbEditPage", function($location, $timeout, $interval, $http, pageService, userService, autocompleteService, markdownService) {
	return {
		templateUrl: "/static/html/editPage.html",
		scope: {
			pageId: "@",
			// If true, we'll show the preview on the right half of the screen
			useFullView: "@",
			// Called when the user is done with the edit.
			doneFn: "&",
		},
		controller: function($scope) {
			$scope.userService = userService;
			$scope.pageService = pageService;
			$scope.page = pageService.editMap[$scope.pageId];
			$scope.selectedTab = ($scope.page.wasPublished || $scope.page.title.length > 0) ? 1 : 0;
			$scope.fullView = false; // $scope.fullView
			$scope.revealAfterRender = false;

			// Select correct tab
			if ($location.search().tab) {
				$scope.selectedTab = $location.search().tab;
			}

			// Set up markdown
			$timeout(function() {
				var $wmdPreview = $("#wmd-preview" + $scope.page.pageId);
				// Initialize pagedown
				markdownService.createEditConverter($scope.page.pageId, function(refreshFunc) {
					$timeout(function() {
						markdownService.processLinks($wmdPreview, refreshFunc);
					});
				});
			});

			// Called when user selects a page from insert link input
			$scope.insertLinkSelect = function(result) {
				if (!$scope.insertLinkCallback) return;
				var result = result;
				$scope.showInsertLink = false;
				$timeout(function() {
					$scope.insertLinkCallback(result.alias);
					$scope.insertLinkCallback = undefined;
				});
			};

			// Toggle in and out of preview when not in fullView
			$scope.inPreview = false;
			$scope.togglePreview = function(show) {
				$scope.inPreview = show;
			};

			// Setup all the settings
			$scope.isWiki = $scope.page.type === "wiki";
			$scope.isQuestion = $scope.page.type === "question";
			$scope.isAnswer = $scope.page.type === "answer";
			$scope.isComment = $scope.page.type === "comment";
			$scope.isLens = $scope.page.type === "lens";
			$scope.isGroup = $scope.page.type === "group" || $scope.page.type === "domain";

			// Set up page types.
			if ($scope.isQuestion) {
				$scope.pageTypes = {question: "Question"};
			} else if($scope.isAnswer) {
				$scope.pageTypes = {answer: "Answer"};
			} else if($scope.isComment) {
				$scope.pageTypes = {comment: "Comment"};
			} else if($scope.isWiki) {
				$scope.pageTypes = {wiki: "Wiki Page"};
			} else if($scope.isLens) {
				$scope.pageTypes = {lens: "Lens Page"};
			}

			// Set up group names.
			var groupIds = userService.user.groupIds;
			$scope.groupOptions = {"0": "-"};
			if (groupIds) {
				for (var i in groupIds) {
					var groupId = groupIds[i];
					var groupName = pageService.pageMap[groupId].title;
					$scope.groupOptions[groupId] = groupName;
				}
			}

			// Set up sort types.
			$scope.sortTypes = {
				likes: "By Likes",
				recentFirst: "Recent First",
				oldestFirst: "Oldest First",
				alphabetical: "Alphabetically",
			};

			// Set up vote types.
			$scope.voteTypes = {
				"": "",
				probability: "Probability",
				approval: "Approval",
			};

			$scope.lockExists = $scope.page.lockedBy != "0" && moment.utc($scope.page.lockedUntil).isAfter(moment.utc());
			$scope.lockedByAnother = $scope.lockExists && $scope.page.lockedBy !== userService.user.id;
			$scope.isMemberOfGroup = $scope.page.editGroupId === '0' || ($scope.page.editGroupId in $scope.groupOptions);

			// Convert all links with pageIds to alias links.
			$scope.page.text = $scope.page.text.replace(complexLinkRegexp, function(whole, prefix, text, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					return prefix + "[" + text + "](" + page.alias + ")";
				}
				return whole;
			/*}).replace(voteEmbedRegexp, function (whole, prefix, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					return prefix + "[vote: " + page.alias + "]";
				}
				return whole;*/
			}).replace(forwardLinkRegexp, function (whole, prefix, alias, text) {
				var page = pageService.pageMap[alias];
				if (page) {
					return prefix + "[" + page.alias + " " + text + "]";
				}
				return whole;
			}).replace(simpleLinkRegexp, function (whole, prefix, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					return prefix + "[" + page.alias + "]";
				}
				return whole;
			}).replace(urlLinkRegexp, function(whole, prefix, text, url, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					return prefix + "[" + text + "](" + url + page.alias + ")";
				}
				return whole;
			}).replace(atAliasRegexp, function(whole, prefix, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					return prefix + "[@" + page.alias + "]";
				}
				return whole;
			});

			// User reverts to an edit
			$scope.revertToEdit = function(editNum) {
				var data = {
					pageId: $scope.page.pageId,
					editNum: editNum,
				};
				$http({method: "POST", url: "/revertPage/", data: JSON.stringify(data)})
				.success(function(data) {
					$location.path(pageService.getPageUrl($scope.page.pageId));
				})
				.error(function(data) {
					$scope.addMessage("revert", "Error reverting: " + data, "error");
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
				$(event.currentTarget).closest("md-list-item").hide();
			};

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
				var data = {
					pageId: $scope.pageId,
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
					// Send the data to the server.
					// TODO: if the call takes too long, we should show a warning.
					$http({method: "POST", url: "/editPage/", data: JSON.stringify(data)})
					.success(function(data) {
						if (isAutosave) {
							// Refresh the lock
							$scope.page.lockedUntil = moment.utc().add(30, "m").format("YYYY-MM-DD HH:mm:ss");
						}
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
				savePage(false, false, function(error) {
					$scope.publishing = false;
					if (error) {
						$scope.addMessage("publish", "Publishing failed: " + error, "error");
					} else {
						$scope.doneFn({result: {pageId: $scope.page.pageId}});
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
						$scope.addMessage("snapshot", "Snapshot failed: " + error, "error");
					} else {
						$scope.addMessage("snapshot", "Snapshot saved!", "info");
					}
				});
			};

			// Process Discard button click.
			$scope.discardPage = function() {
				var cont = function() {
					if ($scope.doneFn) {
						$scope.doneFn({result: {pageId: $scope.page.pageId, discard: true}});
					}
				};
				pageService.discardPage($scope.page.pageId, cont, cont);
			};

			// Process Delete button click.
			$scope.deleting = false;
			$scope.toggleDeleteConfirm = function(show) {
				$scope.deleting = show;
			};
			$scope.deletePage = function() {
				pageService.deletePage($scope.page.pageId, function() {
					if ($scope.doneFn) {
						$scope.doneFn({result: {pageId: $scope.page.pageId, discard: true}});
					}
				}, function(data) {
					$scope.addMessage("delete", "Error deleting page: " + data, "error");
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
					$scope.autosaveError = error;
				}, function() {
					$scope.autosaving = false;
				});
			};
			var autosaveInterval = $interval(autosaveFunc, 5000);
			$scope.$on("$destroy", function() {
				$interval.cancel(autosaveInterval);
				// Autosave just in case.
				savePage(true, false, function() {});
			});

			$scope.saveSettingsClick = function() {
				$scope.savePageInfo(function(error) {
					if (error) {
						$scope.addMessage("pageInfo", "Error saving settings: " + error, "error");
					} else {
						$scope.addMessage("pageInfo", "Settings saved!", "info");
					}
				});
			};

			// =========== Find similar pages ==============
			var shouldFindSimilar = false;
			$scope.forceExpandSimilarPages = $scope.isQuestion;
			$scope.expandSimilarPages = $scope.forceExpandSimilarPages;
			$scope.toggleSimilarPages = function(show) {
				$scope.expandSimilarPages = show;
			};
			$scope.similarPages = [];
			var findSimilarFunc = function() {
				if (!shouldFindSimilar) return;
				shouldFindSimilar = false;
				var data = {
					title: $scope.page.title,
					// Cutting off text at the last (arbitrary) 4k characters, so Elastic doesn't choke
					text: $scope.page.text.length > 4000 ? $scope.page.text.slice(-4000) : $scope.page.text,
					clickbait: $scope.page.clickbait,
					pageType: $scope.page.type,
				};
				autocompleteService.findSimilarPages(data, function(data){
					$scope.similarPages.length = 0;
					for (var n = 0; n < data.length; n++) {
						var pageId = data[n].label;
						if (pageId === $scope.page.pageId) continue;
						$scope.similarPages.push({pageId: pageId, score: data[n].score});
					}
				});
			};
			var similarInterval = $interval(findSimilarFunc, 10000);
			$scope.$on("$destroy", function() {
				$interval.cancel(similarInterval);
			});

			// =========== Show diff between edits ==============
			// otherDiff stores the edit we load for diffing.
			$scope.otherDiff = undefined;
			$scope.diffHtml = undefined;
			// Refresh the diff edit text.
			$scope.refreshDiff = function() {
				var dmp = new diff_match_patch();
				var diffs = dmp.diff_main($scope.otherDiff.text, $scope.page.text);
				dmp.diff_cleanupSemantic(diffs);
				$scope.diffHtml = dmp.diff_prettyHtml(diffs).replace(/&para;/g, "");
			}
			// Process click event for diffing edits.
			$scope.showDiff = function(editNum) {
				// Load the edit from the server.
				pageService.loadEdit({
					pageAlias: $scope.page.pageId,
					specificEdit: editNum,
					skipProcessDataStep: true,
					success: function(data, status) {
						$scope.otherDiff = data[$scope.page.pageId];
						$scope.refreshDiff();
						$scope.selectedTab = 1;
					},
				});
			};
			$scope.hideDiff = function() {
				$scope.otherDiff = undefined;
				$scope.diffHtml = undefined;
			};
		},
		link: function(scope, element, attrs) {
			$timeout(function() {
				// Synchronize scrolling between the textarea and the preview.
				var $divs = element.find(".wmd-input").add(element.find(".preview-area"));
				var syncScroll = function(event) {
					var $other = $divs.not(this).off("scroll"), other = $other.get(0);
					var percentage = this.scrollTop / (this.scrollHeight - this.offsetHeight);
					other.scrollTop = Math.round(percentage * (other.scrollHeight - other.offsetHeight));
					// Firefox workaround. Rebinding without delay isn't enough.
					//setTimeout(function() {
						$other.on("scroll", syncScroll);
					//}, 10);
				};
				$divs.on("scroll", syncScroll);

				// Listen to events from Markdown.Editor
				var $markdownToolbar = element.find(".wmd-button-bar");
				$markdownToolbar.on("showInsertLink", function(event, callback) {
					scope.showInsertLink = true;
					scope.insertLinkCallback = callback;
					// NOTE: not sure why, but we need two timeouts here
					$timeout(function() { $timeout(function() {
						element.find(".insert-autocomplete").find("input").focus();
					}); });
				});

				// Save the page info.
				// callback is called with a potential error message when the server replies
				scope.savePageInfo = function(callback){
					var data = {
						pageId: scope.page.pageId,
						type: scope.page.type,
						editGroupId: scope.page.editGroupId,
						hasVote: scope.page.hasVote,
						voteType: scope.page.voteType,
						editKarmaLock: scope.page.editKarmaLock,
						alias: scope.page.alias,
						sortChildrenBy: scope.page.sortChildrenBy,
					};
					$http({method: "POST", url: "/editPageInfo/", data: JSON.stringify(data)})
					.success(function(data) {
						if(callback) callback();
					})
					.error(function(data) {
						console.log(data);
						if(callback) callback(data);
					});
				};

				$timeout(function() {
					scope.revealAfterRender = true;
				});
			});
		},
	};
});
