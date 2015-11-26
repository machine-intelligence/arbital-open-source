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

			// Set up search
			scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				autocompleteService.performSearch({term: text}, function(results) {
					deferred.resolve(results);
				});
        return deferred.promise;
			};
			scope.searchResultSelected = function(result) {
				/*if (event.ctrlKey) {
					return false;
				}*/
				window.location.href = pageService.getPageUrl(result.label);
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
				var listString = "";
				var listArray = [];
				for (var key in pageService.primaryPage.parents) {
					listArray.push(pageService.primaryPage.parents[key].parentId);
				}
				listString = listArray.join(",");
				window.location.href = "/edit?newParentId=" + listString;
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
