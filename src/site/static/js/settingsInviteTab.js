'use strict';

// Directive for the Settings page.
app.directive('arbSettingsInviteTab', function($http, pageService, userService) {
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
			$scope.creatingPersonalInvite = false;
			$scope.creatingGroupInvite = false;
			// If user tries to submit any invalid emails in GROUP form
			$scope.invalidEmails = [];
			// To show hints below form inputs or not
			$scope.showHints = true;

			// Send invite to server. Either a new invite or an existing one (e.g. undo delete)
			var postInvite = function(invite, successCallback, errorCallback) {
				// What we send to backend (sanitized and pruned of unneeded properties)
				var inviteToPost = {
					invitees:   invite.invitees,
					senderId:   userService.user.id,
					type: invite.type,
					domainId:   invite.domainId,
					domainName: pageService.getDomainName(invite.domainId),
					// If it's a new code, invite.code will be undefined
					// If it's an old code that is unclaimed,
					//    invite.claimedAt and invite.claimingUserId will be undefined
					oldCode:    invite.code,
					claimingUserId: invite.claimingUserId
				};
				$http({
					method: 'POST',
					url: '/sendInvite/',
					data: JSON.stringify(inviteToPost),
				})
				.success(function(data) {
					// Handle success as needed by caller
					successCallback && successCallback(data);
				})
				.error(function(data, status) {
					console.error('Unable to post invite:'); console.log(data, status);
					errorCallback && errorCallback();
				});
			};

			$scope.domainId = '';
			$scope.newInvite = function(invite) {
				// If there are invalid email addresses, return and show alert
				if (!processNewInvite(invite)) return;

				postInvite(invite, function(data) {
					// If there's a new code from backend, it's a new invite
					if (data.result.newCode) {
						invite.code = data.result.newCode;
						// Add the invite to our local model
						$scope.invitesSent.push(invite);
					}
					// Reset forms to blank and hide
					$scope.email = undefined;
					$scope.emails = undefined;
					$scope.moreGroupEmails = undefined;
					invite.adding = false;
					$scope.domainId = '';
					// NOTE: here is what's not working to hide the forms
					$scope.creatingPersonalInvite = false;
					$scope.creatingGroupInvite = false;
				}, function(data) {
					console.error('Unable to post emails'); console.log(data);
				});
			};

			// If it's a new email, have to do some vetting and cleaning
			var processNewInvite = function(invite) {
				// Test email string for validity
				var testedEmails = cleanEmails(invite.emailStr);

				// If there are invalid emails, show user a warning and don't post
				if (testedEmails.invalid.length > 0) {
					$scope.invalidEmails = testedEmails.invalid;
					return false;
				}
				$scope.invalidEmails = [];

				// If there's no code (it's a new invite) parse invitees into new array of obj
				// If adding to existing invite, just push new to invitees array
				if (!invite.code) {
					invite.invitees = [];
					addInvitees(invite, testedEmails.valid);
				} else if (invite.isUpdate) {
					addInvitees(invite, testedEmails.valid);
				}
				return true;
			};

			// Helper function to turn array of emails into array of invitee objects
			var addInvitees = function(invite, emails) {
				for (var i = 0; i < emails.length; i++) {
					invite.invitees.push({
						'email': emails[i],
						'claimingUserId': '',
					});
				};
			};

			// User can add additional addresses to an existing Group invite
			$scope.updateInvite = function(invite, emailStr) {
				invite.isUpdate = true;
				invite.emailStr = emailStr;
				$scope.newInvite(invite);
			};

			// Delete the invitation from DB, but save locally in case user wants to undo
			$scope.deleteInvite = function(invite) {
				// Mark invite as deleted so that it is hidden
				invite.deleted = true;
				var data = {code: invite.code};
				// Delete from db
				$http({
					method: 'POST',
					url: '/deleteInvite/',
					data: JSON.stringify(data)
				})
				.success(function(data) {
					if (data.result.deletionStatus < 0) {
						invite.deleted = false;
						console.error('Error trying to delete invite');
					}
				})
				.error(function(data, status) {
					console.error('Unable to delete invite.'); console.log(data, status);
				});
			};

			// If a user has clicked delete, we give them a chance to undo
			$scope.undoDelete = function(invite) {
				invite.isUpdate = false;
				postInvite(invite, function(data) {
					// Hide the undo button & show the invite again
					invite.deleted = false;
				}, function(err) {
					console.error('Unable to undo delete', err);
				});
			};

			$scope.toggleCreatingPersonalInvite = function() {
				$scope.creatingPersonalInvite = !$scope.creatingPersonalInvite;
			};
			$scope.toggleCreatingGroupInvite = function() {
				$scope.creatingGroupInvite = !$scope.creatingGroupInvite;
			};

			// Clicking the + button shows form to add more email addresses (Group invite only)
			$scope.toggleAddEmails = function(invite) {
				invite.adding = !invite.adding;
			};
		},
	};
});
