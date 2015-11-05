// arb-index directive displays a set of featured domains
app.directive("arbIndex", function(pageService, userService) {
	return {
		templateUrl: "/static/html/indexDir.html",
		scope: {
			featuredDomains: "=",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			console.log(scope.featuredDomains);
		},
	};
});
