"use strict";

// Directive for listing masteries and allowing the user to claim them.
app.directive("arbMasteryList", function($timeout, $http, pageService, userService) {
	return {
		templateUrl: "/static/html/masteryList.html",
		scope: {
			idsSource: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Sort requirements
			$scope.idsSource.sort(function(a, b) {
				var hasA = pageService.hasMastery(a);
				var hasB = pageService.hasMastery(b);
				if (hasA !== hasB) {
					return (hasA ? 1 : 0) - (hasB ? 1 : 0);
				}
				return (pageService.wantsMastery(a) ? 1 : 0) - (pageService.wantsMastery(b) ? 1 : 0);
			});

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function(masteryId) {
				if (pageService.hasMastery(masteryId)) {
					pageService.updateMasteries([], [masteryId], []);
				} else {
					pageService.updateMasteries([masteryId], [], []);
				}
			};
		},
	};
});

