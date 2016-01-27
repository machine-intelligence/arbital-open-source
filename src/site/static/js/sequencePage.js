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

			// Figure our the order of pages through which to take the user
			var computeSequenceIds = function(part, idsList) {
				idsList = idsList || [];
				if (part.requirements) {
					for (var n = 0; n < part.requirements.length; n++) {
						if (pageService.hasMastery(part.requirements[n].pageId)) continue;
						idsList = computeSequenceIds(part.requirements[n], idsList);
					}
				}
				if (part.taughtById !== "0" && idsList.indexOf(part.taughtById) < 0) {
					idsList.push(part.taughtById);
				}
				return idsList;
			};
			
			$scope.startReading = function() {
				var sequenceIds = computeSequenceIds($scope.sequence);
				sequenceIds.push($scope.sequence.pageId);
				var ids = sequenceIds.join(",");
				$location.url(pageService.getPageUrl(sequenceIds[0]) + "?sequence=" + ids);
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

