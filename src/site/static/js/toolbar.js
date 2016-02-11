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
				Cookies.remove("arbital");
				window.location.reload();
			};
		},
		link: function(scope, element, attrs) {
			// Hide toolbar in the edit screen
			function showHideToolBar() {
				if ($location.path().indexOf("/edit") === 0)
					element.hide();
				else
					element.show();
			}

			scope.$on("$locationChangeSuccess", showHideToolBar);
			showHideToolBar();
		},
	};
});
