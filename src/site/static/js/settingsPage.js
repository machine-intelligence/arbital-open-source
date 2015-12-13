"use strict";

// Directive for the Settings page.
app.directive("arbSettingsPage", function($timeout, $http, pageService, userService) {
	return {
		templateUrl: "/static/html/settingsPage.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Set up frequency types.
			$scope.frequencyTypes = {never: "Never", weekly: "Weekly", daily: "Daily", immediately: "Immediately"};

			// Process Email Settings form submission.
			$scope.submitForm = function(event) {
				var data = {
					emailFrequency: userService.user.emailFrequency,
					emailThreshold: userService.user.emailThreshold,
				};
				submitForm($(event.currentTarget), "/updateSettings/", data, function(r) {
					$scope.submitted = true;
					$scope.$apply();
				});
			};
		},
	};
});
