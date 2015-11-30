"use strict";

// Directive for the actual DOM elements which allows the user to edit a page.
app.directive("arbEditPage", function($location, $timeout, $compile, pageService, userService, autocompleteService, markdownFactory) {
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
				console.log(result);
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


			// Helper function for savePage. Computes the data to submit via AJAX.
			var computeAutosaveData = function(isAutosave, isSnapshot) {
				var data = {
					pageId: $scope.pageId,
					title: $scope.page.title,
					clickbait: $scope.page.clickbait,
					text: $scope.page.text,
					isAutosave: isAutosave,
					isSnapshot: isSnapshot,
					__invisibleSubmit: isAutosave,
				};
				if (page.anchorContext) {
					data.anchorContext = page.anchorContext;
					data.anchorText = page.anchorText;
					data.anchorOffset = page.anchorOffset;
				}
				return data;
			};
			$scope.autosaving = false;
			$scope.publishing = false;
			var prevEditPageData = undefined;
			// Save the current page.
			// callback is called when the server replies with success. If it's an autosave
			// and the same data has already been submitted, the callback is called with "".
			var savePage = function(isAutosave, isSnapshot, callback) {
				// Prevent stacking up saves without them returning.
				if ($scope.publishing) return;
				$scope.publishing = !isAutosave;
				if (isAutosave && $scope.autosaving) return;
				$scope.autosaving = isAutosave;

				// Submit the form.
				var data = computeAutosaveData(isAutosave, isSnapshot);
				var $form = $topParent.find(".new-page-form");
				if (!isAutosave || JSON.stringify(data) !== JSON.stringify(prevEditPageData)) {
					// TODO: if the call takes too long, we should show a warning.
					submitForm($form, "/editPage/", data, function(r) {
						if (isAutosave) {
							$scope.autosaving = false;
							// Refresh the lock
							page.lockedUntil = moment.utc().add(30, "m").format("YYYY-MM-DD HH:mm:ss");
						}
						if (isSnapshot) {
							// Prevent an autosave from triggering right after a successful snapshot
							prevEditPageData.isSnapshot = false;
							prevEditPageData.isAutosave = true;
							data.__invisibleSubmit = true; 
						}

						// Process warnings
						var warningsLength = r.result.aliasWarnings.length;
						var $aliasWarning = $topParent.find(".alias-warning");
						$aliasWarning.text("").toggle(warningsLength > 0);
						$topParent.find(".alias-form-group").toggleClass("has-warning", warningsLength > 0);
						for(var n = 0; n < warningsLength; n++){
							$aliasWarning.text($aliasWarning.text() + r.result.aliasWarnings[n] + "\n");
						}
						callback(true);
					}, function() {
						if (isAutosave) $scope.autosaving = false;
						if ($scope.publishing) {
							$scope.publishing = false;
							// Pretend it was a failed autosave
							data.__invisibleSubmit = true; 
							data.isAutosave = true;
						}
					});
					prevEditPageData = data;
				} else {
					callback(false);
					$scope.autosaving = false;
				}
			}
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
