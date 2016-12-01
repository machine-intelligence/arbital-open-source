'use strict';

import app from './angular.ts';
import {submitForm, arraysSortFn} from './util.ts';

// Directive for the Domains page.
app.directive('arbDomainsPage', function($timeout, $http, $filter, arb) {
	return {
		templateUrl: versionUrl('static/html/domainsPage.html'),
		scope: {
			invitesSent: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.newGroupForm = {};
			$scope.invitesMap = {}; // domain id -> list of invites

			// Map domain id -> list of user ids that are members of the given domain
			$scope.domainUsersMap = {};
			for (let userId in arb.userService.userMap) {
				for (let domainId in arb.userService.userMap[userId].domainMembershipMap) {
					if (!(domainId in arb.userService.user.domainMembershipMap)) continue;
					if (domainId in $scope.domainUsersMap) {
						if ($scope.domainUsersMap[domainId].indexOf(userId) < 0) {
							$scope.domainUsersMap[domainId].push(userId);
						}
					} else {
						$scope.domainUsersMap[domainId] = [userId];
						$scope.invitesMap[domainId] = [];
					}
				}
			}

			// Split invites by domain
			for (let n = 0; n < $scope.invitesSent.length; n++) {
				let invite = $scope.invitesSent[n];
				$scope.invitesMap[invite.domainId].push(invite);
			}

			// Sort lists of users based on the role and last name
			let roleToInt = function(role) {
				return ['banned', '', 'default', 'trusted', 'reviewer', 'arbiter', 'arbitrator'].indexOf(role);
			};
			for (let domainId in $scope.domainUsersMap) {
				$scope.domainUsersMap[domainId].sort(arraysSortFn(function(userId) {
					var user = arb.userService.userMap[userId];
					return [
						-roleToInt(user.domainMembershipMap[domainId].role),
						user.firstName,
					];
				}));
				$scope.invitesMap[domainId].sort(function(a, b) {
					return a.toEmail.localeCompare(b.toEmail);
				});
			}

			// Get text describing what the current status of the invite is.
			$scope.getInviteStatus = function(invite) {
				if (invite.claimedAt.length > 0 && invite.claimedAt[0] !== '0') {
					return 'claimed ' + $filter('smartDateTime')(invite.claimedAt);
				} else if (invite.emailSentAt.length > 0 && invite.emailSentAt[0] !== '0') {
					return 'invite sent ' + $filter('smartDateTime')(invite.emailSentAt);
				}
				return 'not claimed yet';
			};

			// Process updating a member's permissions
			$scope.updateMemberPermissions = function(domainId, userId) {
				let data = {
					userId: userId,
					domainId: domainId,
					role: arb.userService[userId].domainMembershipMap[domainId].role,
				};
				arb.stateService.postDataWithoutProcessing('/updateDomainRole/', data);
			};

			// Process new member form submission.
			$scope.newMemberFormSubmit = function(domainId, userInput) {
				var data = {
					domainId: domainId,
					userInput: userInput,
				};
				arb.stateService.postDataWithoutProcessing('/newMember/', data, function() {
					arb.popupService.showToast({
						text: 'User added. Refresh the page if you need to change the role from Default.',
					});
				});
			};

			// Process new invite form submission.
			$scope.newInviteFormSubmit = function(domainId, toEmail, domainMembership) {
				var params = {
					domainId: domainId,
					toEmail: toEmail,
					role: domainMembership.role,
				};
				arb.stateService.postDataWithoutProcessing('/newInvite/', params, function(data) {
					$scope.invitesMap[domainId].push(data.result.newInvite);
					arb.popupService.showToast({
						text: 'Invite sent.',
					});
				});
			};
		},
	};
});
