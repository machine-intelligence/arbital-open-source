"use strict";

// Directive for the Login page.
app.directive("arbLogin", function($location, pageService, userService) {
	return {
		templateUrl: "static/html/loginPage.html",
		scope: {
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
		},
	};
});
