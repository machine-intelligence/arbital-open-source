'use strict';

// Directive for the Settings page.
app.directive('arbSettingsPage', function($http, arb) {
	return {
		templateUrl: 'static/html/settingsPage.html',
		scope: {
			domains: '=',
			invitesSent: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Set up frequency types.
			$scope.frequencyTypes = {
				never: 'Never',
				weekly: 'Weekly',
				daily: 'Daily',
				immediately: 'Immediately',
			};

			// Process settings form submission.
			$scope.submitForm = function(event) {
				var data = {
					emailFrequency: userService.user.emailFrequency,
					emailThreshold: userService.user.emailThreshold,
					ignoreMathjax: userService.user.ignoreMathjax,
				};
				submitForm($(event.currentTarget), '/updateSettings/', data, function(r) {
					$scope.submitted = true;
					$scope.$apply();
				}, function(err) {
					console.error('ERROR while updating settings:', err);
				});
			};
		},
	};
});
