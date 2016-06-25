'use strict';

// Directive for the Signup page.
app.directive('arbSignup', function($location, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/signupPage.html'),
		controller: function($scope) {
			$scope.arb = arb;
			$scope.formData = {};

			var onSignupSuccess = function() {
				arb.signupService.closeSignupDialog();
			};

			$scope.formSubmit = function(event) {
				arb.analyticsService.reportSignupAction('submit signup with email', arb.signupService.attemptedAction);
				submitForm($(event.currentTarget), '/signup/', $scope.formData, function(r) {
					arb.analyticsService.reportSignupAction('success signup with email', arb.signupService.attemptedAction);
					onSignupSuccess();
					$scope.$apply(function() {
						arb.urlService.goToUrl($location.search().continueUrl || '/');
					});
				}, function() {
					$scope.$apply(function() {
						$scope.normalError = '(Check if your password meets the requirements.)';
					});
				});
			};

			$scope.signupWithFb = function() {
				arb.analyticsService.reportSignupAction('click signup with fb', arb.signupService.attemptedAction);
				arb.userService.fbLogin(function(response) {
					if (response.status === 'connected') {
						var data = {
							fbAccessToken: response.authResponse.accessToken,
							fbUserId: response.authResponse.userID,
						};
						$http({method: 'POST', url: '/signup/', data: JSON.stringify(data)})
							.success(function(data, status) {
								arb.urlService.goToUrl($location.search().continueUrl || '/');
								arb.analyticsService.reportSignupAction('success signup with fb', arb.signupService.attemptedAction);
								onSignupSuccess();
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
