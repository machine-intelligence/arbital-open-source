"use strict";

// Directive for listing masteries and allowing the user to claim them.
app.directive("arbMasteryList", function($timeout, $http, pageService, userService) {
	return {
		templateUrl: "/static/html/masteryList.html",
		scope: {
			idsSource: "=",
			// If true, show the requisites the user has first
			showHasFirst: "=",
			// If true, allow the user to toggle through want states
			allowWants: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Sort requirements
			$scope.idsSource.sort(function(a, b) {
				var hasA = pageService.hasMastery(a);
				var hasB = pageService.hasMastery(b);
				if (hasA !== hasB) {
					var result = (hasA ? 1 : 0) - (hasB ? 1 : 0);
					if ($scope.showHasFirst) result = -result;
					return result;
				}
				var result = (pageService.wantsMastery(a) ? 1 : 0) - (pageService.wantsMastery(b) ? 1 : 0);
				if ($scope.showHasFirst) result = -result;
				return result;
			});

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function(masteryId) {
				if (pageService.hasMastery(masteryId)) {
					pageService.updateMasteries([], [masteryId], []);
				} else {
					if ($scope.allowWants && !pageService.wantsMastery(masteryId)) {
						pageService.updateMasteries([], [], [masteryId]);
					} else {
						pageService.updateMasteries([masteryId], [], []);
					}
				}
			};
		},
	};
});

