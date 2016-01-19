"use strict";

// Directive for the Knowledge page.
app.directive("arbKnowledgePage", function(pageService, userService) {
	return {
		templateUrl: "/static/html/knowledgePage.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.masteryList = [];
			for (var id in pageService.masteryMap) {
				$scope.masteryList.push(id);
			}
		},
	};
});
