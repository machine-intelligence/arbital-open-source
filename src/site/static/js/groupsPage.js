"use strict";

// Directive for the Groups page.
app.directive("arbGroupsPage", function(pageService, userService, autocompleteService, $timeout, $http) {
	return {
		templateUrl: "/static/html/groupsPage.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			scope.groupMap = {};

			// Get the data.
			$http({method: "POST", url: "/json/groups/"}).
				success(function(data, status){
					console.log("JSON /groups/ data:"); console.log(data);
					userService.processServerData(data);
					pageService.processServerData(data);

					for (var pageId in pageService.pageMap) {
						if (!+pageId) continue; // don't process aliases
						var page = pageService.pageMap[pageId];
						if (page.type === "group") {
							scope.groupMap[pageId] = page;
						}
					}
				}).error(function(data, status){
					console.log("Error groups page:"); console.log(data); console.log(status);
				}
			);

			// Process removing user from a group.
			element.on("click", ".remove-from-group", function(event) {
				var $target = $(event.target);
				var data = {
					userId: $target.closest("[member-id]").attr("member-id"),
					groupId: $target.closest("[group-id]").attr("group-id"),
				};
				$http({method: "POST", url: "/deleteMember/", data: JSON.stringify(data)})
					.error(function(data, status){
						console.log("Error deleting user:"); console.log(data); console.log(status);
					});

				// Adjust data
				delete scope.groupMap[data.groupId].members[data.userId];
			});

			// Process updating a member's permissions
			var updateMemberPermission = function($target, data) {
				data.userId = $target.closest("[member-id]").attr("member-id");
				data.groupId = $target.closest("[group-id]").attr("group-id");
				var member = scope.groupMap[data.groupId].members[data.userId];
				if (!("canAddMembers" in data)) data.canAddMembers = member.canAddMembers;
				if (!("canAdmin" in data)) data.canAdmin = member.canAdmin;
				$http({method: "POST", url: "/updateMember/", data: JSON.stringify(data)})
					.error(function(data, status){
						console.log("Error updatig member:"); console.log(data); console.log(status);
					});
			};
			element.on("click", ".set-can-add-members", function(event) {
				var $target = $(event.target);
				updateMemberPermission($target, {
					canAddMembers: $target.is(":checked"),
				});
			});
			element.on("click", ".set-can-admin", function(event) {
				var $target = $(event.target);
				updateMemberPermission($target, {
					canAdmin: $target.is(":checked"),
				});
			});

			// Process new member form submission.
			element.on("submit", ".new-member-form", function(event) {
				var $target = $(event.target);
				var data = {
					groupId: $target.attr("group-id"),
				};
				submitForm($target, "/newMember/", data, function(r) {
					location.reload();
				});
				return false;
			});
		
			// Process new group form submission.
			var $form = $("#new-group-form");
			$form.on("submit", function(event) {
				var data = {
					name: $form.attr("name"),
					alias: $form.attr("alias"),
				};
				submitForm($form, "/newGroup/", data, function(r) {
					location.reload();
				});
				return false;
			});
		},
	};
});
