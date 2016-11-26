'use strict';

import app from './angular.ts';
import {submitForm, arraysSortFn} from './util.ts';

// Directive for the Domains page.
app.directive('arbDomainsPage', function($timeout, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/domainsPage.html'),
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.newGroupForm = {};

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
					}
				}
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
			}

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
			$scope.newMemberFormSubmit = function(event, domainId, userInput) {
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
		},
	};
});
