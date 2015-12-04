"use strict";

// Directive for the User page.
app.directive("arbUserPage", function(pageService, userService) {
	return {
		templateUrl: "/static/html/userPage.html",
		scope: {
			userId: "@",
			idsMap: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
		},
	};
});

// Directive for the User page panel
app.directive("arbUserPagePanel", function(pageService, userService) {
	return {
		templateUrl: "/static/html/userPagePanel.html",
		scope: {
			pageIds: "=",
			panelTitle: "@",
			// Whether to show a public/private icon, pass in "yes"/"no" respectively.
			isPublic: "@",
			hideLikes: "@",
			showLastEdit: "@",
			showQuickEdit: "@",
			showRedLinkCount: "@",
			showCommentCount: "@",
			// If set, we'll pull the page from the editMap instead of pageMap
			useEditMap: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			scope.getPage = function(pageId) {
				if (scope.useEditMap) {
					return pageService.editMap[pageId];
				} 
				return pageService.pageMap[pageId];
			};
		},
	};
});
