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

			// Figure out what page to start the user on
			var getStartId = function(part) {
				var startId = "0";
				if (part.requirements) {
					for (var n = 0; n < part.requirements.length; n++) {
						if (pageService.hasMastery(part.requirements[n].pageId)) continue;
						startId = getStartId(part.requirements[n]);
						if (startId !== "0") break;
					}
				}
				if (startId === "0") startId = part.taughtById;
				return startId;
			};
			
			$scope.startReading = function() {
				var startId = getStartId($scope.sequence);
				$location.url(pageService.getPageUrl(startId) + "?sequence=" + $scope.sequence.pageId);
			};
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

