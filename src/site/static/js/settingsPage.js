"use strict";

// Directive for the Settings page.
app.directive("arbSettingsPage", function(pageService, userService, autocompleteService, $timeout, $http) {
	return {
		templateUrl: "/static/html/settingsPage.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			// Set up frequency types.
			scope.frequencyTypes = {never: "Never", weekly: "Weekly", daily: "Daily", immediately: "Immediately"};

			// Get the data.
			scope.emailFrequency = userService.user.emailFrequency;
			scope.emailThreshold = userService.user.emailThreshold;

			// Store the last saved values, to show or hide the "Submitted" text
			scope.savedEmailFrequency = scope.emailFrequency;
			scope.savedEmailThreshold = scope.emailThreshold;

			scope.resultText = "";

			// Process Email Settings form submission.
			var $form = $("#settings-form");
			$form.on("submit", function(event) {
				var data = {
					emailFrequency: $form.attr("emailFrequency"),
					emailThreshold: $form.attr("emailThreshold"),
				};
				submitForm($form, "/updateSettings/", data, function(r) {
					scope.savedEmailFrequency = scope.emailFrequency;
					scope.savedEmailThreshold = scope.emailThreshold;
					scope.resultText = "Submitted";
					scope.$apply();
				});
				return false;
			});

		},
	};
});
