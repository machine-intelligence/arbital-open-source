'use strict';

// Directive for the Requisites page.
app.directive('arbRequisitesPage', function(arb) {
	return {
		templateUrl: 'static/html/requisitesPage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.masteryList = [];
			for (var id in arb.pageService.masteryMap) {
				$scope.masteryList.push(id);
			}

			// Set all requisites to "not known"
			$scope.resetAll = function() {
				arb.pageService.updateMasteryMap({delete: $scope.masteryList});
			};
		},
	};
});
