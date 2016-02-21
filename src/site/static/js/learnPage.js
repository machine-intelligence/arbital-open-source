"use strict";

// Directive for the learn page.
app.directive("arbLearnPage", function($location, pageService, userService) {
	return {
		templateUrl: "static/html/learnPage.html",
		scope: {
			pageId: "@",
			learnMap: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			// Ordered list of page ids in the generated learn
			$scope.readIds = [];
			// If a requisite can't be learned (probably because there is no page that
			// currently teaches it), we add it to this map.
			// requirement id -> [list of page ids that require it]
			$scope.unlearnableIds = {};
			$scope.hasUnlernableIds = false;

			// Figure our the order of pages through which to take the user
			var computeLearnIds = function() {
				$scope.readIds = [];
				$scope.missingTaughtPartIds = {};
				var processNode = function(pageId, parentPageId) {
					var part = $scope.learnMap[pageId];
					for (var n = 0; n < part.requirementIds.length; n++) {
						//if (pageService.hasMastery(part.requirements[n].pageId)) continue;
						processNode(part.requirementIds[n], part.pageId);
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
				processNode($scope.pageId, undefined);
				if ($scope.readIds.indexOf($scope.pageId) < 0) {
					$scope.readIds.push($scope.pageId);
				}
				$scope.hasUnlearnableIds = Object.keys($scope.unlearnableIds).length > 0;
			};

			// Get the url for the given page (optional) with learn support
			$scope.getLearnUrl = function(startingPageId) {
				startingPageId = startingPageId || $scope.readIds[0];
				return pageService.getPageUrl(startingPageId) + "?learn=" + $scope.readIds.join(",");
			};

			// Called when the user clicks to start reading the learn
			$scope.startReading = function() {
				computeLearnIds();
				// Start them off with the first page
				$location.url($scope.getLearnUrl());
			};

			// Track whether we show tree or list view
			$scope.showTreeView = true;
			$scope.toggleView = function() {
				$scope.showTreeView = !$scope.showTreeView;
				if (!$scope.showTreeView) {
					// User might have changed their requisites, so let's recompute everything
					computeLearnIds();
				}
			};
			$scope.toggleView();
		},
	};
});

// Directive for a recursive part of a learn.
app.directive("arbLearnPart", function(pageService, userService, RecursionHelper) {
	return {
		templateUrl: "static/html/learnPart.html",
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.part = $scope.learnMap[$scope.pageId];
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});

