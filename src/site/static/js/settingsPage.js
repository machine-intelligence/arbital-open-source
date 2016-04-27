'use strict';

// Directive for the Settings page.
app.directive('arbSettingsPage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/settingsPage.html',
		scope: {
			domains: '=',
			invitesSent: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.newInviteCodeClaimed = '';

			// Set up frequency types.
			$scope.frequencyTypes = {
				never: 'Never',
				weekly: 'Weekly',
				daily: 'Daily',
				immediately: 'Immediately',
			};

			// Whether the user has claimed any invites; whether to show the list of claimed codes/domains
			$scope.currentUserHasClaimedInvites = function() {
				return Object.keys(userService.user.invitesClaimed).length > 0;
			};

			// Process settings form submission.
			$scope.submitForm = function(event) {
				var data = {
					emailFrequency: userService.user.emailFrequency,
					emailThreshold: userService.user.emailThreshold,
					newInviteCodeClaimed: $scope.newInviteCodeClaimed.toUpperCase(),
					ignoreMathjax: userService.user.ignoreMathjax,
				};
				submitForm($(event.currentTarget), '/updateSettings/', data, function(r) {
					if (!!r.result && !!r.result.invite) {
						var invite = r.result.invite;
						// Add claimed code to invitesClaimed model and UI table
						userService.user.invitesClaimed[invite.code] = invite;
						$scope.newInviteCodeClaimed = '';
					}
					$scope.submitted = true;
					$scope.$apply();
				}, function(err) {
					console.error('ERROR while updating settings:', err);
				});
			};
		},
	};
});
