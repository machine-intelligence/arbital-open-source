'use strict';

import app from './angular.ts';

// Directive for the Settings page.
app.directive('arbSettingsPage', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/settingsPage.html'),
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
				arb.userService.updateSettings(function successFn() {
					$scope.submitted = true;
				});
			};
		},
	};
});
