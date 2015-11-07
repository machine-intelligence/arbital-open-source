// arb-group-index directive displays a set of links to pages
app.directive("arbGroupIndex", function(pageService, userService, autocompleteService, $http) {
	return {
		templateUrl: "/static/html/groupIndex.html",
		scope: {
			idsMap: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.user = userService.user;
		},
	};
});
