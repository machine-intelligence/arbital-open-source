'use strict';

// arb-group-index directive displays a set of links to pages
app.directive('arbGroupIndex', function(arb) {
	return {
		templateUrl: 'static/html/groupIndexPage.html',
		scope: {
			groupId: '@',
			idsMap: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.groupId];
			$scope.showingText = $scope.page.isNewPage() || !userService.user.id;
			$scope.showText = function() {
				$scope.showingText = true;
			};
		},
	};
});
