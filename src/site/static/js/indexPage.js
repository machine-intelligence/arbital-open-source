'use strict';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/indexPage.html',
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});
