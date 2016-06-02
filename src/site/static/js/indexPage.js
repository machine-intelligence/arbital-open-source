'use strict';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, arb) {
	return {
		templateUrl: 'static/html/indexPage.html',
		controller: function($scope) {
			$scope.arb = arb;
			$scope.readTab = 0;
			$scope.writeTab = 0;

			$scope.selectReadTab = function(tab) {
				$scope.readTab = tab;
			};
		},
	};
});
