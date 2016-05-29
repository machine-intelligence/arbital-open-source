'use strict';

// arb-explore-page directive displays a set of featured domains
app.directive('arbExplorePage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/explorePage.html',
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});
