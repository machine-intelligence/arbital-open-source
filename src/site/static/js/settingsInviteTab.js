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

			// Controls whether forms to create invites are shown
			$scope.creatingInvite = false;
			// If user tries to submit any invalid emails in GROUP form
			$scope.invalidEmails = [];
			$scope.domainId = '';

			$scope.invitesSent.sort(function(a, b) {
				return a.toEmail.localeCompare(b.toEmail);
			});

			// Create new invites for the given emails to the given domain.
			$scope.newInvite = function(emails, domainId) {
				// Test email string for validity
				var testedEmails = cleanEmails(emails);
				if (testedEmails.invalid.length > 0) {
					$scope.invalidEmails = testedEmails.invalid;
					return;
				}
				$scope.invalidEmails = [];

				for (var n = 0; n < testedEmails.valid.length; n++) {
					let email = testedEmails.valid[n];
					var inviteToPost = {
						toEmail: email,
						domainId: domainId,
					};
					$http({
						method: 'POST',
						url: '/newInvite/',
						data: JSON.stringify(inviteToPost),
					})
					.success(function(data) {
						$scope.creatingInvite = false;
						$scope.invitesSent.unshift(data.result.invite);
					})
					.error(function(data, status) {
						console.error('Unable to create invites:'); console.log(data);
						$scope.invalidEmails.push(email);
					});
				}
			};

			$scope.toggleCreatingInvite = function() {
				$scope.creatingInvite = !$scope.creatingInvite;
			};

			// Get text describing what the current status of the invite is.
			$scope.getInviteStatus = function(invite) {
				if (invite.claimedAt.length > 0 && invite.claimedAt[0] !== '0') {
					return "claimed " + $filter('relativeDateTime')(invite.claimedAt);
				} else if (invite.emailSentAt.length > 0 && invite.emailSentAt[0] !== '0') {
					return "invite sent " + $filter('relativeDateTime')(invite.emailSentAt);
				} 
				return "not claimed yet";
			};
		},
	};
});
