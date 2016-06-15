'use strict';

// arb-group-index directive displays a set of links to pages
app.directive('arbGroupIndex', function(arb) {
	return {
		templateUrl: versionUrl('static/html/groupIndexPage.html'),
		scope: {
			groupId: '@',
			idsMap: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.groupId];
			$scope.showingText = $scope.page.isNewPage() || !arb.userService.user.id;
			$scope.showText = function() {
				$scope.showingText = true;
			};
		},
	};
});
