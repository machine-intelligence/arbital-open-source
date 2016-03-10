"use strict";

// Directive for the learn page.
app.directive("arbLearnPage", function($location, $compile, pageService, userService) {
	return {
		templateUrl: "static/html/learnPage.html",
		scope: {
			pageIds: "=",
			optionsMap: "=",
			tutorMap: "=",
			requirementMap: "=",
			continueLearning: "=",
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
				$scope.unlearnableIds = {};
				// Function for recursively processing a pageId that needs to be learned
				var processRequirement = function(pageId, parentPageId) {
					var requirement = $scope.requirementMap[pageId];
					var tutor = requirement.bestTutorId ? $scope.tutorMap[requirement.bestTutorId] : undefined;
					if (tutor) {
						for (var n = 0; n < tutor.requirementIds.length; n++) {
							processRequirement(tutor.requirementIds[n], pageId);
						}
						if ($scope.readIds.indexOf(tutor.pageId) < 0) {
							$scope.readIds.push(tutor.pageId);
						}
					} else {
						if (!(pageId in $scope.unlearnableIds)) {
							$scope.unlearnableIds[pageId] = [];
						}
						if (parentPageId && $scope.unlearnableIds[pageId].indexOf(parentPageId) < 0) {
							$scope.unlearnableIds[pageId].push(parentPageId);
						}
					}
				};
				// Process the requirements tree into a linear path
				for (var n = 0; n < $scope.pageIds.length; n++) {
					var pageId = $scope.pageIds[n];
					if (pageId in $scope.requirementMap) {
						processRequirement(pageId, undefined);
						if ($scope.readIds.indexOf(pageId) < 0 && $scope.optionsMap[pageId].appendToPath) {
							$scope.readIds.push(pageId);
						}
					}
				}
				$scope.hasUnlearnableIds = Object.keys($scope.unlearnableIds).length > 0;
			};

			// Called when the user clicks to start reading the learn
			$scope.startReading = function(redirect) {
				computeLearnIds();
				var path = {
					pageIds: $scope.pageIds,
					readIds: $scope.readIds,
					unlearnableIds: $scope.unlearnableIds,
				};
				Cookies.set("path", path);
				if (redirect) {
					// Start them off with the first page
					$location.url(pageService.getPageUrl($scope.readIds[0]));
				}
			};

			// Track whether we show tree or list view
			$scope.showTreeView = !$location.search().showTree;
			$scope.toggleView = function() {
				$scope.showTreeView = !$scope.showTreeView;
				if (!$scope.showTreeView) {
					// User might have changed their requisites, so let's recompute everything
					computeLearnIds();
				}
			};
			$scope.toggleView();
		},
		link: function(scope, element, attrs) {
			// Change which tutor to use for learning the given requisite.
			scope.changeTutor = function(reqId, newTutorId) {
				scope.requirementMap[reqId].bestTutorId = newTutorId;
				$compile(element.find(".root-learn-part"))(scope);
			};
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
			$scope.requirement = $scope.requirementMap[$scope.pageId];
			$scope.tutor = $scope.requirement.bestTutorId ? $scope.tutorMap[$scope.requirement.bestTutorId] : undefined;
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});
