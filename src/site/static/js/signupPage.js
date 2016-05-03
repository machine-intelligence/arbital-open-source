'use strict';

// Directive for the Signup page.
app.directive('arbSignup', function($location, $http, pageService, userService) {
	return {
		templateUrl: 'static/html/signupPage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.formData = {};

			$scope.formSubmit = function(event) {
				submitForm($(event.currentTarget), '/signup/', $scope.formData, function(r) {
					$scope.$apply(function() {
						$scope.signupSuccess = true;
						$scope.normalError = undefined;
					});
				}, function() {
					$scope.$apply(function() {
						$scope.normalError = '(Check if your password meets the requirements.)';
					});
				});
			};

			$scope.signupWithFb = function() {
				userService.fbLogin(function(response) {
					if (response.status === 'connected') {
						var data = {
							fbAccessToken: response.authResponse.accessToken,
							fbUserId: response.authResponse.userID,
						};
						$http({method: 'POST', url: '/signup/', data: JSON.stringify(data)})
						.success(function(data, status) {
							window.location.href = $location.search().continueUrl || '/';
						})
						.error(function(data, status) {
							console.error('Error FB signup:'); console.log(data); console.log(status);
						});
					} else {
						$scope.socialError = 'Error: ' + response.status;
					}
				});
			};

			// Allow access to global isLive().
			$scope.isLive = isLive;
		},
	};
});
