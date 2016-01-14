"use strict";

// Directive for multiple choice
app.directive("arbMultipleChoice", function($timeout, $http, pageService, userService) {
	return {
		restrict: "A",
		templateUrl: "/static/html/multipleChoice.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Check if the user has the given mastery.
			$scope.hasMastery = function(masteryId) {
				return pageService.masteryMap[masteryId].has;
			};

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function(masteryId) {
				pageService.updateMastery($scope, masteryId, !$scope.hasMastery(masteryId));
			};
		},
		link: function(scope, element, attrs) {
			$(element).replaceWith("div");
		},
	};
});

