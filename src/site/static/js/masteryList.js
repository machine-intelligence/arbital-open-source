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

			// Check if the user has the given mastery
			$scope.hasMastery = function(masteryId) {
				return pageService.masteryMap[masteryId].has;
			};

			// Check if the user wants the given mastery
			$scope.wantsMastery = function(masteryId) {
				return pageService.masteryMap[masteryId].wants;
			};

			// Sort requirements
			$scope.idsSource.sort(function(a, b) {
				var hasA = $scope.hasMastery(a);
				var hasB = $scope.hasMastery(b);
				if (hasA !== hasB) {
					return (hasA ? 1 : 0) - (hasB ? 1 : 0);
				}
				return ($scope.wantsMastery(a) ? 1 : 0) - ($scope.wantsMastery(b) ? 1 : 0);
			});

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function(masteryId) {
				if ($scope.hasMastery(masteryId)) {
					pageService.updateMasteries([], [], [masteryId]); // wants
				} else if ($scope.wantsMastery(masteryId)) {
					pageService.updateMasteries([], [masteryId], []); // doesn't have
				} else {
					pageService.updateMasteries([masteryId], [], []); // has
				}
			};
		},
	};
});

