"use strict";

// Directive for the Signup page.
app.directive("arbSignup", function($location, pageService, userService) {
	return {
		templateUrl: "static/html/signupPage.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.formData = {};

			$scope.formSubmit = function(event) {
				submitForm($(event.currentTarget), "/signup/", $scope.formData, function(r) {
					$scope.$apply(function() {
						$scope.signupSuccess = true;
					});
				}, function() {
				});
			};
		},
	};
});
