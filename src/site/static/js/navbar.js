"use strict";

// navbar directive displays the navbar at the top of each page
app.directive("arbNavbar", function(pageService, userService, autocompleteService, $http, $location, $compile) {
	return {
		templateUrl: "/static/html/navbar.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			// Get a domain url (with optional subdomain)
			scope.getDomainUrl = function(subdomain) {
				if (subdomain) {
					subdomain += ".";
				} else {
					subdomain = "";
				}
				if (/localhost/.exec($location.host())) {
					return "http://" + subdomain + "localhost:8012";
				} else {
					return "http://" + subdomain + "arbital.com"
				}
			};

			$("#logout").click(function() {
				$.removeCookie("zanaduu", {path: "/"});
			});

			// Setup search via navbar.
			var $navSearch = element.find("#nav-search");
			if ($navSearch.length > 0) {
				$navSearch.autocomplete({
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
						window.location.href = "/pages/" + ui.item.value;
						return false;
					},
				});
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
