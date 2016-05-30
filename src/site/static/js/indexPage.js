'use strict';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, arb) {
	return {
		templateUrl: 'static/html/indexPage.html',
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
