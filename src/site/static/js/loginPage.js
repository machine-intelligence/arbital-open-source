"use strict";

// Directive for the Login page.
app.directive("arbLogin", function($location, $http, pageService, userService) {
	return {
		templateUrl: "static/html/loginPage.html",
		scope: {
			// True if the login is embe
			isEmbedded: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.formData = {};

			$scope.formSubmit = function(event) {
				submitForm($(event.currentTarget), "/login/", $scope.formData, function(r) {
					window.location.href = $location.search().continueUrl || "/";
				}, function() {
				});
			};

			$scope.loginWithFb = function() {
				userService.fbLogin(function(response){
					if (response.status === "connected") {
						var data = {
							fbAccessToken: response.authResponse.accessToken,
							fbUserId: response.authResponse.userID,
						};
						$http({method: "POST", url: "/signup/", data: JSON.stringify(data)})
						.success(function(data, status){
							window.location.href = $location.search().continueUrl || "/";
						})
						.error(function(data, status){
							console.error("Error FB signup:"); console.log(data); console.log(status);
						});
					} else {
						$scope.socialError = "Error: " + response.status;
					}
				});
			};
		},
	};
});
