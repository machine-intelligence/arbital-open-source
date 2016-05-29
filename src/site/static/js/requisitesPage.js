'use strict';

// Directive for the Requisites page.
app.directive('arbRequisitesPage', function(arb) {
	return {
		templateUrl: 'static/html/requisitesPage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.masteryList = [];
			for (var id in pageService.masteryMap) {
				$scope.masteryList.push(id);
			}

			// Set all requisites to "not known"
			$scope.resetAll = function() {
				pageService.updateMasteryMap({delete: $scope.masteryList});
			};
		},
	};
});
