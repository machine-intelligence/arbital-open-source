"use strict";

// Directive for the Updates page.
app.directive("arbUpdates", function(pageService, userService, $location) {
	return {
		templateUrl: "/static/html/updatesDir.html",
		scope: {
			updateGroups: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			$(".group-subscribe-to-user-link").on("click", function(event) {
				userService.subscribeToUser($(event.target));
				// Need to update other links to the same user. Can remove when we start using AngularJS here.
				$(".group-subscribe-to-user-link[user-id='" + $target.attr("user-id") + "']").toggleClass("on", $target.hasClass("on"));
				return false;
			});
			$(".group-subscribe-to-page-link").on("click", function(event) {
				var $target = $(event.target);
				userService.subscribeToPage($target);
				// Need to update other links to the same page. Can remove when we start using AngularJS here.
				$(".group-subscribe-to-page-link[page-id='" + $target.attr("page-id") + "']").toggleClass("on", $target.hasClass("on"));
				return false;
			});
		},
	};
});
