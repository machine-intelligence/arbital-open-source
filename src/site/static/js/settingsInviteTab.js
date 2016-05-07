'use strict';

// Directive for the Settings page.
app.directive('arbSettingsInviteTab', function($http, $filter, pageService, userService) {
	return {
		templateUrl: 'static/html/settingsInviteTab.html',
		scope: {
			domains: '=',
			invitesSent: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.alwaysTrue = true;

			// Controls whether form to create invite is shown
			$scope.creatingInvite = false;
			// If user tries to submit any invalid emails
			$scope.invalidEmails = [];
			$scope.selectedDomains = {};

			$scope.invitesSent.sort(function(a, b) {
				return a.toEmail.localeCompare(b.toEmail);
			});

			// Create new invites for the given emails to the given domains.
			$scope.newInvite = function(emails) {
				// Test email string for validity
				var testedEmails = cleanEmails(emails);
				if (testedEmails.invalid.length > 0) {
					$scope.invalidEmails = testedEmails.invalid;
					return;
				}
				$scope.invalidEmails = [];

				// Get domainIds
				var domainIds = [];
				for (var domainId in $scope.selectedDomains) {
					if ($scope.selectedDomains[domainId]) {
						domainIds.push(domainId);
					}
				}

				for (var n = 0; n < testedEmails.valid.length; n++) {
					var email = testedEmails.valid[n];
					var inviteToPost = {
						toEmail: email,
						domainIds: domainIds,
					};
					// use a self-invoking anonymous function so that the email variable in the callbacks below refers to
					// a different value with each iteration of the for loop, rather than being a fixed reference that
					// takes on whatever value email had when the callback was called
					(function(_email) {
						$http({
							method: 'POST',
							url: '/newInvite/',
							data: JSON.stringify(inviteToPost),
						})
						.success(function(data) {
							$scope.creatingInvite = false;
							for (var domainId in data.result.inviteMap) {
								$scope.invitesSent.unshift(data.result.inviteMap[domainId]);
							}
						})
						.error(function(data, status) {
							console.error('Unable to create invites:'); console.log(data);
							$scope.invalidEmails.push(_email);
						});
					})(email);
				}
			};

			$scope.toggleCreatingInvite = function() {
				$scope.creatingInvite = !$scope.creatingInvite;
			};

			// Get text describing what the current status of the invite is.
			$scope.getInviteStatus = function(invite) {
				if (invite.claimedAt.length > 0 && invite.claimedAt[0] !== '0') {
					return 'claimed ' + $filter('relativeDateTime')(invite.claimedAt);
				} else if (invite.emailSentAt.length > 0 && invite.emailSentAt[0] !== '0') {
					return 'invite sent ' + $filter('relativeDateTime')(invite.emailSentAt);
				}
				return 'not claimed yet';
			};
		},
	};
});
