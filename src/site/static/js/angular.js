"use strict";

// Set up angular module.
var app = angular.module("arbital", ["ngMaterial", "ngResource", "ngRoute",
		"ngMessages", "ngSanitize", "RecursionHelper"]);

app.config(function($interpolateProvider, $locationProvider, $provide, $routeProvider, $mdIconProvider){
	$mdIconProvider.icon("thumb_up_outline", "static/icons/thumb-up-outline.svg")
		.icon("thumb_down_outline", "static/icons/thumb-down-outline.svg")
		.icon("link_variant", "static/icons/link-variant.svg")
		.icon("format_header_pound", "static/icons/format-header-pound.svg");

	$locationProvider.html5Mode(true);

	$routeProvider
	.when("/", {
		template: "",
		controller: "IndexPageController",
	})
	.when("/domains/:alias", {
		template: "",
		controller: "DomainPageController",
	})
	.when("/pages/:alias", {
		template: "",
		controller: "PrimaryPageController",
		reloadOnSearch: false,
	})
	.when("/edit/:alias?/:edit?", {
		template: "",
		controller: "EditPageController",
		reloadOnSearch: false,
	})
	.when("/user/:id?", {
		template: "",
		controller: "UserPageController",
	})
	.when("/updates/", {
		template: "",
		controller: "UpdatesPageController",
	})
	.when("/groups/", {
		template: "",
		controller: "GroupsPageController",
	})
	.when("/signup/", {
		template: "",
		controller: "SignupPageController",
	})
	.when("/settings/", {
		template: "",
		controller: "SettingsPageController",
	});
});

// ArbitalCtrl is used across all pages.
app.controller("ArbitalCtrl", function ($scope, $location, $timeout, $interval, $http, $compile, $anchorScroll, $mdDialog, userService, pageService, popoverService) {
	$scope.pageService = pageService;
	$scope.userService = userService;

	// Get subdomain if any
	$scope.subdomain = undefined;
	var subdomainMatch = /^([A-Za-z0-9]+)\.(localhost|arbital\.com)\/?$/.exec($location.host());
	if (subdomainMatch) {
		$scope.subdomain = subdomainMatch[1];
	}

	// Refresh all the fields that need to be updated every so often.
	var refreshAutoupdates = function() {
		$(".autoupdate").each(function(index, element) {
			$compile($(element))($scope);
		});
		$timeout(refreshAutoupdates, 30000);
	};
	refreshAutoupdates();

	// Returns a function we can use as success handler for POST requests for dynamic data.
	// callback - returns {
	//   title: title to set for the window
	//   element: optional jQuery element to add dynamically to the body
	//   error: optional error message to print
	// }
	$scope.getSuccessFunc = function(callback) {
		return function(data) {
			// Sometimes we don't get data.
			if (data) {
				console.log("Dynamic request data:"); console.log(data);
				userService.processServerData(data);
				pageService.processServerData(data);
			}

			// Because the subdomain could have any case, we need to find the alias
			// in the loaded map so we can get the alias with correct case
			if ($scope.subdomain) {
				for (var pageAlias in pageService.pageMap) {
					if ($scope.subdomain.toUpperCase() === pageAlias.toUpperCase()) {
						$scope.subdomain = pageAlias;
						pageService.privateGroupId = pageService.pageMap[pageAlias].pageId;
						break;
					}
				}
			}

			// Get the results from page-specific callback
			$(".global-error").hide();
			var result = callback(data);
			if (result.error) {
				$(".global-error").text(result.error).show();
				document.title = "Error - Arbital";
			}
			if (result.element) {
				// Only show the element after it and all the children have been fully compiled and linked
				result.element.addClass("reveal-after-render-parent");
				$("#loading-bar").show();
				var revealInterval = $interval(function() {
					var hiddenChildren = result.element.find(".reveal-after-render");
					if (hiddenChildren.length > 0) {
						hiddenChildren.each(function() {
							if ($(this).children().length > 0) {
								$(this).removeClass("reveal-after-render");
							}
						});
						return;
					}
					$interval.cancel(revealInterval);
					// Do short timeout to prevent some rendering bugs that occur on edit page
					$timeout(function() {
						result.element.removeClass("reveal-after-render-parent");
						$("#loading-bar").hide();
						$anchorScroll();
					}, 50);
				}, 50);

				$("[ng-view]").append(result.element);
			}
			if (result.title) {
				document.title = result.title + " - Arbital";
			}
		};
	};

	// Returns a function we can use as error handler for POST requests for dynamic data.
	$scope.getErrorFunc = function(urlPageType) {
		return function(data, status){
			console.error("Error /json/" + urlPageType + "/:"); console.log(data); console.log(status);
			$(".global-error").text(data).show();
			document.title = "Error - Arbital";
		};
	}
});

// simpleDateTime filter converts our typical date&time string into local time.
app.filter("simpleDateTime", function() {
	return function(input) {
		return moment.utc(input).format("LT, l");
	};
});

// relativeDateTime converts date&time into a relative string, e.g. "5 days ago"
app.filter("relativeDateTime", function() {
	return function(input) {
		return moment.utc(input).fromNow();
	};
});
app.filter("relativeDateTimeNoSuffix", function() {
	return function(input) {
		return moment.utc(input).fromNow(true);
	};
});


app.controller("IndexPageController", function ($scope, $routeParams, $http, $compile, pageService, userService) {
	if ($scope.subdomain) {
		// Get the private group index page data
		$http({method: "POST", url: "/json/privateIndex/"})
		.success($scope.getSuccessFunc(function(data){
			$scope.indexPageIdsMap = data.result;
			return {
				title: pageService.pageMap[$scope.subdomain].title + " - Private Group",
				element: $compile("<arb-group-index ids-map='indexPageIdsMap'></arb-group-index>")($scope),
			};
		}))
		.error($scope.getErrorFunc("privateIndex"));
	} else {
		// Get the index page data
		$http({method: "POST", url: "/json/index/"})
		.success($scope.getSuccessFunc(function(data){
			$scope.featuredDomains = data.result.featuredDomains;
			return {
				element: $compile("<arb-index featured-domains='featuredDomains'></arb-index>")($scope),
			};
		}))
		.error($scope.getErrorFunc("index"));
	}
});

app.controller("DomainPageController", function ($scope, $routeParams, $http, $compile, pageService, userService) {
	pageService.domainAlias = $routeParams.alias;
	var postData = {
		domainAlias: pageService.domainAlias,
	};
	// Get the domain index page data
	$http({method: "POST", url: "/json/domainPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		$scope.indexPageIdsMap = data.result;
		var groupId = pageService.pageMap[pageService.domainAlias].pageId;
		return {
			title: pageService.pageMap[groupId].title,
			element: $compile("<arb-group-index group-id='" + groupId +
				"' ids-map='indexPageIdsMap'></arb-group-index>")($scope),
		};
	}))
	.error($scope.getErrorFunc("domainPage"));
});

app.controller("PrimaryPageController", function ($scope, $routeParams, $http, $compile, pageService, userService) {
	// Get the primary page data
	var postData = {
		pageAlias: $routeParams.alias,
	};
	$http({method: "POST", url: "/json/primaryPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		var page = pageService.pageMap[postData.pageAlias];
		if (!page) {
			return {
				title: "Not Found",
				error: "Page doesn't exist, was deleted, or you don't have permission to view it.",
			};
		}
		pageService.primaryPage = page;
		return {
			title: page.title,
			element: $compile("<arb-primary-page></arb-primary-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("primaryPage"));
});


app.controller("EditPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var pageId = $routeParams.alias;

	// Need to call /default/ in case we are creating a new page
	// TODO(alexei): have /newPage/ return /default/ data long with /edit/ data
	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		if (+pageId) {
			// Load the last edit
			pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: $routeParams.edit ? +$routeParams.edit : 0,
				success: $scope.getSuccessFunc(function() {
					var page = pageService.editMap[pageId];
					if ($location.search().alias) {
						// Set page's alias
						page.alias = $location.search().alias;
						$location.replace().search("alias", undefined);
					}

					pageService.primaryPage = page;
			
					// Called when the user is done editing the page.
					$scope.doneFn = function(result) {
						var page = pageService.editMap[result.pageId];
						if (!page.wasPublished && result.discard) {
							$location.path("/edit/");
						} else {
							$location.path(pageService.getPageUrl(page.pageId));
						}
					};
					return {
						title: "Edit " + (page.title ? page.title : "New Page"),
						element: $compile("<div arb-edit-page class='full-height' page-id='" + pageId +
							"' done-fn='doneFn(result)'></div>")($scope),
					};
				}),
				error: $scope.getErrorFunc("loadEdit"),
			});
		} else {
			var type = $location.search().type;
			$location.replace().search("type", undefined);
			var newParentIdString = $location.search().newParentId;
			$location.replace().search("newParentId", undefined);
			// Create a new page to edit
			pageService.getNewPage({
				type: type,
				parentIds: newParentIdString ? newParentIdString.split(",") : [],
				success: function(newPageId) {
					$location.replace().path(pageService.getEditPageUrl(newPageId));
				},
			});
		}
		return {
			title: "Edit Page",
		};
	}))
	.error($scope.getErrorFunc("Edit Page"));
});

app.controller("UserPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var userId = $routeParams.id;
	if (!userId) {
		userId = userService.user.id;
	}
	var postData = {
		userId: userId,
	};
	// Get the data
	$http({method: "POST", url: "/json/userPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		$scope.userPageIdsMap = data.result;
		return {
			title: userService.userMap[userId].firstName + " " + userService.userMap[userId].lastName,
			element: $compile("<arb-user-page user-id='" + userId + "' ids-map='userPageIdsMap'></arb-user-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("User"));
});

app.controller("UpdatesPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var postData = { };
	// Get the explore data
	$http({method: "POST", url: "/json/updates/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		$scope.updateGroups = data.result.updateGroups;
		return {
			title: "Updates",
			element: $compile("<arb-updates update-groups='updateGroups'></arb-updates>")($scope),
		};
	}))
	.error($scope.getErrorFunc("Updates"));
});

app.controller("GroupsPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/groups/"})
	.success($scope.getSuccessFunc(function(data){
		return {
			title: "Groups",
			element: $compile("<arb-groups-page></arb-groups-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("Groups"));
});

app.controller("SignupPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		if (!userService.user || userService.user.id === "0") {
			window.location.href = "/login/?continueUrl=" + encodeURIComponent($location.search().continueUrl);
			return {};
		}
		return {
			title: "Sign Up",
			element: $compile("<arb-signup></arb-signup>")($scope),
		};
	}))
	.error($scope.getErrorFunc("Signup"));
});

app.controller("SettingsPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
		$http({method: "POST", url: "/json/default/"})
		.success($scope.getSuccessFunc(function(data){
			return {
				title: "Settings",
				element: $compile("<arb-settings-page></arb-settings-page>")($scope),
			};
		}))
		.error($scope.getErrorFunc("Settings"));
});
