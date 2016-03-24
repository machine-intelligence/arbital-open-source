'use strict';

// toolbar directive displays the toolbar at the top of each page
app.directive('arbToolbar', function($mdSidenav, $http, $location, $compile, $rootScope, $timeout, $q, $mdMedia, pageService, userService, autocompleteService, urlService) {
	return {
		templateUrl: 'static/html/toolbar.html',
		scope: {
			loadingBarValue: '=',
			currentUrl: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.urlService = urlService;
			$scope.isTinyScreen = !$mdMedia('gt-xs');

			$scope.doAutofocus = function() {
				return !userService.isTouchDevice && !urlService.hasLoadedFirstPage;
			};

			// Called when a search result is selected
			$scope.searchResultSelected = function(result) {
				if (result) {
					urlService.goToUrl(pageService.getPageUrl(result.pageId));
				}
			};

			$scope.getSignupUrl = function() {
				return '/signup/?continueUrl=' + encodeURIComponent($location.absUrl());
			};

			// Open RHS menu
			$scope.toggleRightMenu = function() {
				$mdSidenav('right').toggle();
			};

			$scope.logout = function() {
				Cookies.remove('masteryMap');
				Cookies.remove('arbital');
				window.location.reload();
			};

			// Hide toolbar in the edit screen
			$scope.$on('$locationChangeSuccess', function() {
				$scope.hide = $location.path().indexOf('/edit') === 0;
			});
			$scope.hide = $location.path().indexOf('/edit') === 0;
		},
	};
});
