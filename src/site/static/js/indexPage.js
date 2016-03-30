'use strict';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/indexPage.html',
		scope: {
			featuredDomains: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			// HARDCODED
			$scope.page = pageService.pageMap['1k0'];
			$scope.showingText = $scope.page.isNewPage() || !userService.user.id;

			$scope.subscribeData = {
				interests: {
					'7ec5d431b0': true,
					'7b38bc3921': false,
				},
			};
			$scope.subscribed = false;
			$scope.subscribeToList = function() {
				$http({method: 'POST', url: '/mailchimpSignup/', data: JSON.stringify($scope.subscribeData)})
				.success(function(data) {
					$scope.subscribed = true;
					$scope.subscribeError = undefined;
				})
				.error(function(data) {
					$scope.subscribeError = data;
				});
			};
		},
	};
});
