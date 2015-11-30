"use strict";

// toolbar directive displays the toolbar at the top of each page
app.directive("arbToolbar", function($mdSidenav, $http, $location, $compile, $rootScope, $timeout, $q, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/toolbar.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			// Keep the current url updated
			scope.currentUrl = encodeURIComponent($location.absUrl());
			$rootScope.$on("$routeChangeSuccess", function() {
				scope.currentUrl = encodeURIComponent($location.absUrl());
			});

			// Called when a search result is selected
			scope.searchResultSelected = function(result) {
				if (result) {
					window.location.href = pageService.getPageUrl(result.label);
				}
			}

			// Open RHS menu
			scope.toggleRightMenu = function() {
		    $mdSidenav("right").toggle();
		  };

			// Handle logging out
			$("#logout").click(function() {
				$.removeCookie("zanaduu", {path: "/"});
			});

			$("#newPageButton").click(function() {
				window.location.href = "/edit";
			});

			$("#newSibling").click(function() {
				var parentString = pageService.primaryPage.parentIds.join(",");
				window.location.href = "/edit?newParentId=" + parentString;
			});
		},
	};
});

// footer directive displays the page's footer
app.directive("arbFooter", function() {
	return {
		templateUrl: "/static/html/footer.html",
	};
});
