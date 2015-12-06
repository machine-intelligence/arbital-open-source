"use strict";

// toolbar directive displays the toolbar at the top of each page
app.directive("arbToolbar", function($mdSidenav, $http, $location, $compile, $rootScope, $timeout, $q, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/toolbar.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Keep the current url updated
			$scope.currentUrl = encodeURIComponent($location.absUrl());
			$rootScope.$on("$routeChangeSuccess", function() {
				$scope.currentUrl = encodeURIComponent($location.absUrl());
			});

			// Called when a search result is selected
			$scope.searchResultSelected = function(result) {
				if (result) {
					window.location.href = pageService.getPageUrl(result.pageId);
				}
			}

			// Open RHS menu
			$scope.toggleRightMenu = function() {
		    $mdSidenav("right").toggle();
		  };
		},
	};
});
