"use strict";

// toolbar directive displays the toolbar at the top of each page
app.directive("arbToolbar", function($mdSidenav, $http, $location, $compile, $rootScope, pageService, userService, autocompleteService) {
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

			scope.openRightMenu = function() {
		    $mdSidenav("right").toggle();
		  };

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

			// Setup search via toolbar.
			var $navSearch = element.find("#nav-search");
			if ($navSearch.length > 0) {
				/*$navSearch.autocomplete({
					source: function(request, callback) {
						autocompleteService.performSearch({term: request.term}, callback);
					},
					minLength: 3,
					delay: 400,
					focus: function (event, ui) {
						return false;
					},
					select: function (event, ui) {
						if (event.ctrlKey) {
							return false;
						}
						window.location.href = pageService.getPageUrl(ui.item.label);
						return false;
					},
				});*/
				autocompleteService.setAutocompleteRendering($navSearch, scope, true);
			}
		},
	};
});

// footer directive displays the page's footer
app.directive("arbFooter", function() {
	return {
		templateUrl: "/static/html/footer.html",
	};
});
