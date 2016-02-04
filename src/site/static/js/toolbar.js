"use strict";

// toolbar directive displays the toolbar at the top of each page
app.directive("arbToolbar", function($mdSidenav, $http, $location, $compile, $rootScope, $timeout, $q, $mdMedia, pageService, userService, autocompleteService) {
	return {
		templateUrl: "static/html/toolbar.html",
		scope: {
			loadingBarValue: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.isTinyScreen = !$mdMedia("gt-xs");
			$scope.doAutofocus = !userService.isTouchDevice;

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

			$scope.logout = function() {
			};

			// Hide toolbar in the edit screen
			$scope.$on("$locationChangeSuccess", function () {
				$scope.hide = $location.path().indexOf("/e") === 0;
			});
			$scope.hide = $location.path().indexOf("/e") === 0;
		},
	};
});
