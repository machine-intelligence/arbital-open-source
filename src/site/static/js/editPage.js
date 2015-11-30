"use strict";

// Directive for the actual DOM elements which allows the user to edit a page.
app.directive("arbEditPage", function($location, $timeout, $interval, $http, pageService, userService, markdownFactory) {
	return {
		templateUrl: "/static/html/editPage.html",
		scope: {
			pageId: "@",
			showPreview: "@",
			isModal: "@",
			// Page this page will "belong" to (e.g. answer belongs to a question,
			// comment belongs to the page it's on)
			primaryPageId: "@",
			// Context, test, and offset are set for editing inline comments
			context: "@",
			text: "@",
			offset: "@",
			// Called when the user is done with the edit.
			doneFn: "&",
		},
		controller: function($scope) {
			$scope.userService = userService;
			$scope.pageService = pageService;
			$scope.page = pageService.editMap[$scope.pageId];

			// Select correct tab
			if ($location.search().tab) {
				$scope.selectedTab = $location.search().tab;
			}

			// Called when user selects a page from insert link input
			$scope.insertLinkSelect = function(result) {
				$scope.showInsertLink = false;
				$timeout(function() {
					$scope.insertLinkCallback(result.alias);
				});
			};

			// Wait until the DOM is created
			$timeout(function() {
				// Initialize pagedown
				new markdownFactory($scope.page.pageId);
			});

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
			} else {
				$scope.pageTypes = {wiki: "Wiki Page", lens: "Lens Page"};
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

			// =========== Error, warning, and info management system ==============
			var messages = {};
			var addMessage = function(key, text, type) {
				messages[key] = {text: text, type: type};
			};
			var deleteMessage = function(key) {
				delete messages[key];
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
					prevEditPageData = $.extend({}, data);
					data.isAutosave = isAutosave;
					data.isSnapshot = isSnapshot;
					console.log("HTTP");
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
						addMessage("publish", "Publishing failed: " + error, "error");
					} else {
						doneFn({alias: $scope.page.pageId});
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
						addMessage("snapshot", "Snapshot failed: " + error, "error");
					} else {
						addMessage("snapshot", "Snapshot saved!", "info");
					}
				});
			};

			// Process Discard button click.
			$scope.discardPage = function() {
				if (doneFn) {
					doneFn({alias: $scope.page.pageId, abandon: true});
				}
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
		},
		link: function(scope, element, attrs) {
			$timeout(function() {
				/*var $textarea = ;
				var textarea = $textarea[0];
				$textarea.on("scroll", function(event) {
					console.log(textarea.scrollTop);
					console.log(textarea.scrollTop / (textarea.scrollHeight - textarea.offsetHeight));
				});*/

				var $divs = $(element).find(".wmd-input").add($(element).find(".preview-area"));
				var syncScroll = function(event) {
					var $other = $divs.not(this).off("scroll"), other = $other.get(0);
					var percentage = this.scrollTop / (this.scrollHeight - this.offsetHeight);
					other.scrollTop = percentage * (other.scrollHeight - other.offsetHeight);
					// Firefox workaround. Rebinding without delay isn't enough.
					//setTimeout(function() {
						$other.on("scroll", syncScroll);
					//}, 10);
				};
				var mouseEnterScrollArea = function(event) {
					var $other = $divs.not(this).off("scroll");
					console.log(this);
					console.log($other);
					$(this).on("scroll", syncScroll);
				};
				$divs.on("mouseenter", mouseEnterScrollArea);
			});
		},
	};
});

// Directive for showing page's change log.
app.directive("arbChangelog", function(pageService, userService) {
	return {
		templateUrl: "/static/html/changelog.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.pageService = pageService;
			scope.page = pageService.editMap[scope.pageId];
		},
	};
});
