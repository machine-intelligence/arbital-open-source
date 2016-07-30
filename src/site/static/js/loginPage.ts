'use strict';

import app from './angular.ts';
import {submitForm} from './util.ts';

// Directive for the Login page.
app.directive('arbLogin', function($location, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/loginPage.html'),
		scope: {
			// True if the login is embe
			isEmbedded: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.formData = {};
			$scope.forgotPasswordSuccess = false;

			$scope.formSubmit = function(event) {
				submitForm($(event.currentTarget), '/login/', $scope.formData, function(r) {
					window.location.href = $location.search().continueUrl || '/';
				}, function() {
				});
			};

			$scope.loginWithFb = function() {
				arb.userService.fbLogin(function(response) {
					if (response.status === 'connected') {
						var data = {
							fbAccessToken: response.authResponse.accessToken,
							fbUserId: response.authResponse.userID,
						};
						$http({method: 'POST', url: '/signup/', data: JSON.stringify(data)})
						.success(function(data, status) {
							var defaultPath = $location.path();
							if (defaultPath.indexOf('/login/') >= 0 || defaultPath.indexOf('/signup/') >= 0) {
								defaultPath = '/';
							}
							window.location.href = $location.search().continueUrl || defaultPath;
						})
						.error(function(data, status) {
							console.error('Error FB signup:'); console.log(data); console.log(status);
						});
					} else {
						$scope.socialError = 'Error: ' + response.status;
					}
				});
			};

			$scope.forgotPassword = function() {
				arb.stateService.postDataWithoutProcessing('/json/forgotPassword/',
					{email: $scope.formData.email},
					function() {
						$scope.forgotPasswordSuccess = true;
					}
				);
			};
		},
	};
});
