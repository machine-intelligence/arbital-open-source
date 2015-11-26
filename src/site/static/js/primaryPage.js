"use strict";

// Directive for the entire primary page.
app.directive("arbPrimaryPage", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/primaryPage.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.primaryPage;
		},
	};
});
