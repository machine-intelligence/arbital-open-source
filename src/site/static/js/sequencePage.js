"use strict";

// Directive for the sequence page.
app.directive("arbSequencePage", function($location, pageService, userService) {
	return {
		templateUrl: "static/html/sequencePage.html",
		scope: {
			sequence: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			// Ordered list of page ids in the generated sequence
			$scope.readIds = [];
			// If a requisite can't be learned (probably because there is no page that
			// currently teaches it), we add it to this map.
			// requirement id -> [list of page ids that require it]
			$scope.unlearnableIds = {};
			$scope.hasUnlernableIds = false;

			// Figure our the order of pages through which to take the user
			var computeSequenceIds = function() {
				$scope.readIds = [];
				$scope.missingTaughtPartIds = {};
				var processPart = function(part, parentPageId) {
					if (part.requirements) {
						for (var n = 0; n < part.requirements.length; n++) {
							if (pageService.hasMastery(part.requirements[n].pageId)) continue;
							processPart(part.requirements[n], part.pageId);
						}
					}
					if (part.taughtById !== "") {
						if ($scope.readIds.indexOf(part.taughtById) < 0) {
							$scope.readIds.push(part.taughtById);
						}
					} else {
						if (!(part.pageId in $scope.unlearnableIds)) {
							$scope.unlearnableIds[part.pageId] = [];
						}
						if (parentPageId && $scope.unlearnableIds[part.pageId].indexOf(parentPageId) < 0) {
							$scope.unlearnableIds[part.pageId].push(parentPageId);
						}
					}
				};
				processPart($scope.sequence, undefined);
				if ($scope.readIds.indexOf($scope.sequence.pageId) < 0) {
					$scope.readIds.push($scope.sequence.pageId);
				}
				$scope.hasUnlearnableIds = Object.keys($scope.unlearnableIds).length > 0;
			};

			// Get the url for the given page (optional) with sequence support
			$scope.getSequenceUrl = function(startingPageId) {
				startingPageId = startingPageId || $scope.readIds[0];
				return pageService.getPageUrl(startingPageId) + "?sequence=" + $scope.readIds.join(",");
			};

			// Called when the user clicks to start reading the sequence
			$scope.startReading = function() {
				computeSequenceIds();
				// Start them off with the first page
				$location.url($scope.getSequenceUrl());
			};

			// Track whether we show tree or list view
			$scope.showTreeView = true;
			$scope.toggleView = function() {
				$scope.showTreeView = !$scope.showTreeView;
				if (!$scope.showTreeView) {
					// User might have changed their requisites, so let's recompute everything
					computeSequenceIds();
				}
			};
			$scope.toggleView();
		},
	};
});

// Directive for a recursive part of a sequence.
app.directive("arbSequencePart", function(pageService, userService, RecursionHelper) {
	return {
		templateUrl: "static/html/sequencePart.html",
		scope: {
			part: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});

