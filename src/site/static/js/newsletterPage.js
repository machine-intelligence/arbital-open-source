'use strict';

// arbNewsletter directive displays a way for the user to edit their newsletter preferences
app.directive('arbNewsletter', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/newsletterPage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.alreadySubscribed = true;

			var interestMap = userService.user.mailchimpInterests;
			if (Object.keys(interestMap).length <= 0) {
				interestMap = {
					'7ec5d431b0': true,
					'7b38bc3921': false,
				};
				$scope.alreadySubscribed = false;
			}

			$scope.subscribeData = {
				email: userService.user.email,
				interests: interestMap,
			};
			$scope.subscribed = false;
			$scope.subscribeToList = function() {
				$http({method: 'POST', url: '/mailchimpSignup/', data: JSON.stringify($scope.subscribeData)})
				.success(function(data) {
					$scope.subscribed = true;
					$scope.subscribeError = undefined;
					$scope.$digest();
				})
				.error(function(data) {
					$scope.subscribeError = data;
					$scope.$digest();
				});
			};
		},
	};
});