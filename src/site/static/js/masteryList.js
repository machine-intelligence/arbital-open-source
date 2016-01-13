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

			// Check if the user has the given mastery.
			$scope.hasMastery = function(masteryId) {
				return pageService.masteryMap[masteryId].has;
			};

			// Sort requirements
			$scope.idsSource.sort(function(a, b) {
				return ($scope.hasMastery(a) ? 1 : 0) - ($scope.hasMastery(b) ? 1 : 0);
			});

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function(masteryId) {
				pageService.updateMastery($scope, masteryId, !$scope.hasMastery(masteryId));
			};
		},
	};
});

