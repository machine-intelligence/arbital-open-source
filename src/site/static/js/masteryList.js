"use strict";

// Directive for listing masteries and allowing the user to claim them.
app.directive("arbMasteryList", function($timeout, $http, pageService, userService) {
	return {
		templateUrl: "static/html/masteryList.html",
		scope: {
			idsSource: "=",
			// If true, don't show the checkboxes
			hideCheckboxes: "=",
			// If true, show the requisites the user has first
			showHasFirst: "=",
			// If true, allow the user to toggle through want states
			allowWants: "=",
			// If true, recursively show requirements for each mastery
			showRequirements: "=",
			// If true, show the requirements on one line
			isSpan: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Filter non-existing page ids out
			$scope.idsSource = $scope.idsSource.filter(function(pageId) {
				return (pageId in pageService.pageMap) && !pageService.pageMap[pageId].isDeleted();
			});

			// Sort requirements
			$scope.idsSource.sort(function(a, b) {
				var result = (pageService.hasMastery(a) ? 1 : 0) - (pageService.hasMastery(b) ? 1 : 0);
				if ($scope.showHasFirst) result = -result;
				if (result !== 0) return result;
				result = (pageService.wantsMastery(a) ? 1 : 0) - (pageService.wantsMastery(b) ? 1 : 0);
				if ($scope.showHasFirst) result = -result;
				if (result !== 0) return result;
				return pageService.pageMap[a].title.localeCompare(pageService.pageMap[b].title);
			});
		},
	};
});

