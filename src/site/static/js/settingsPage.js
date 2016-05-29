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
			$scope.arb = arb;

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
					emailFrequency: arb.userService.user.emailFrequency,
					emailThreshold: arb.userService.user.emailThreshold,
					ignoreMathjax: arb.userService.user.ignoreMathjax,
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
