"use strict";

// Directive for the sequence page.
app.directive("arbSequencePage", function(pageService, userService) {
	return {
		templateUrl: "/static/html/sequencePage.html",
		scope: {
			sequence: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Figure out what page to start the user on
			var getStartId = function(part) {
				var startId = "0";
				var found = false;
				if (part.requirements) {
					for (var n = 0; n < part.requirements.length; n++) {
						startId = getStartId(part.requirements[0]);
						if (startId !== "0" && !pageService.hasMastery(startId)) {
							found = true;
							break;
						}
					}
				}
				if (!found) startId = part.taughtById;
				return startId;
			}
			var computeStartUrl = function() {
				$scope.startId = getStartId($scope.sequence);
				$scope.startUrl = pageService.getPageUrl($scope.startId) + "?sequence=" + $scope.sequence.pageId;
			};
			computeStartUrl();
			
			// Watch if any of the masteries change, and update the start URL.
			$scope.$watch(function() {
				var ids = [];
				for (var id in pageService.masteryMap) {
					ids.push(id);
				}
				return ids.join(",");
			}, function() {
				computeStartUrl();
			});
		},
	};
});

// Directive for a recursive part of a sequence.
app.directive("arbSequencePart", function(pageService, userService, RecursionHelper) {
	return {
		templateUrl: "/static/html/sequencePart.html",
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

