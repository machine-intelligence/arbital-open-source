'use strict';

// Directive for the Signup page.
app.directive('arbSignup', function($location, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/signupPage.html'),
		scope: {},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.formData = {};

			var onSignupSuccess = function() {
				arb.signupService.closeSignupDialog();
				if ($location.search().continueUrl) {
					arb.urlService.goToUrl($location.search().continueUrl);
				}
			};

			$scope.formSubmit = function(event) {
				arb.analyticsService.reportSignupAction('submit signup with email', arb.signupService.attemptedAction);
				arb.stateService.postData('/signup/', $scope.formData,
					function(r) {
						arb.analyticsService.reportSignupAction('success signup with email', arb.signupService.attemptedAction);
						onSignupSuccess();
					},
					function() {
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
						arb.stateService.postData('/signup/', JSON.stringify(data),
							function(data, status) {
								arb.analyticsService.reportSignupAction('success signup with fb', arb.signupService.attemptedAction);
								onSignupSuccess();
							},
							function(data, status) {
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
