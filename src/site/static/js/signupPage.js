"use strict";

// Directive for the Signup page.
app.directive("arbSignup", function($location, pageService, userService) {
	return {
		templateUrl: "/static/html/signupPage.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.formData = {
				email: userService.user.email,
			};

			$scope.formSubmit = function(event) {
				submitForm($(event.currentTarget), "/signup/", $scope.formData, function(r) {
					window.location.href = $location.search().continueUrl;
				}, function() {
				});
			};
		},
	};
});
