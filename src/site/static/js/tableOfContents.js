"use strict";

// Directive for table of contents
app.directive("arbTableOfContents", function($timeout, $http, $compile, pageService, userService) {
	return {
		templateUrl: "static/html/tableOfContents.html",
		transclude: true,
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.showToc = true;
			$scope.toc = [];
		},
		link: function(scope, element, attrs) {
			var $parent = element.closest("arb-markdown");

			$parent.find("h1,h2,h3").each(function () {
				var $this = $(this);
				var headerType = $this.prop("nodeName");
				if (headerType === "h1") {
				} else if (headerType === "h2") {
				} else if (headerType === "h3") {
				}
				console.log($this.text());
			});
		},
	};
});

