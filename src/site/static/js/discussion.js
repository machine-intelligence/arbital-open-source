// Directive to show the discussion section for a page
app.directive("arbDiscussion", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/discussion.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			scope.page.subpageIds = (scope.page.questionIds || []).concat(scope.page.commentIds || []);
			// TODO: sort
		},
	};
});
