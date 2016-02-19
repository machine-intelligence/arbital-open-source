"use strict";

// Directive for the Groups page.
app.directive("arbGroupsPage", function(pageService, userService, autocompleteService, $timeout, $http) {
	return {
		templateUrl: "static/html/groupsPage.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.newGroupForm = {};

			// Populate the groupMap with the groups the user belongs to
			$scope.groupMap = {};
			for (var n = 0; n < $scope.userService.user.groupIds.length; n++) {
				var groupId = $scope.userService.user.groupIds[n];
				$scope.groupMap[groupId] = pageService.pageMap[groupId];
			}

			// Process removing user from a group.
			$scope.removeFromGroup = function(groupId, userId) {
				var data = {
					userId: userId,
					groupId: groupId,
				};
				$http({method: "POST", url: "/deleteMember/", data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error deleting user:"); console.log(data); console.log(status);
				});

				// Adjust data
				delete $scope.groupMap[data.groupId].members[data.userId];
			};

			// Process updating a member's permissions
			$scope.updateMemberPermissions = function(groupId, userId) {
				var member = $scope.groupMap[groupId].members[userId];
				if (member.canAdmin) {
					member.canAddMembers = true;
				}
				var data = {
					userId : userId,
					groupId: groupId,
					canAddMembers: member.canAddMembers,
					canAdmin: member.canAdmin,
				};
				$http({method: "POST", url: "/updateMember/", data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error updating member:"); console.log(data); console.log(status);
				});
			};

			// Process new member form submission.
			$scope.newMemberFormSubmit = function(event, groupId, userId) {
				var data = {
					groupId: groupId,
					userId: userId,
				};
				submitForm($(event.currentTarget), "/newMember/", data, function(r) {
					location.reload();
				});
			};

			// Process new group form submission.
			$scope.newGroupFormSubmit = function(event) {
				var data = {
					name: $scope.newGroupForm.newGroupName,
					alias: $scope.newGroupForm.newGroupAlias,
				};
				submitForm($(event.currentTarget), "/newGroup/", data, function(r) {
					location.reload();
				});
			};
		},
	};
});
