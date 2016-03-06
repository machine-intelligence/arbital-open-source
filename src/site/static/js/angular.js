"use strict";

// Set up angular module.
var app = angular.module("arbital", ["ngMaterial", "ngResource", "ngSilent", "ngRoute",
		"ngMessages", "ngSanitize", "RecursionHelper", "as.sortable"]);

app.config(function($locationProvider, $routeProvider, $mdIconProvider, $mdThemingProvider){
	// Convert "rgb(#,#,#)" color to "#hex"
	var rgb2hex = function(rgb) {
		if (rgb === undefined)
			return '#000000';
		rgb = rgb.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
		function hex(x) {
			return ("0" + parseInt(x).toString(16)).slice(-2);
		}
		return "#" + hex(rgb[1]) + hex(rgb[2]) + hex(rgb[3]);
	}
	// Create themes, by getting the colors from our css files
	$mdThemingProvider.definePalette("arb-primary-theme", $mdThemingProvider.extendPalette("teal", {
		"500": rgb2hex($("#primary-color").css("border-top-color")),
		"300": rgb2hex($("#primary-color").css("border-right-color")),
		"800": rgb2hex($("#primary-color").css("border-bottom-color")),
		"A100": rgb2hex($("#primary-color").css("border-left-color")),
		"contrastDefaultColor": "light",
		"contrastDarkColors": ["300"],
	}));
	$mdThemingProvider.definePalette("arb-accent-theme", $mdThemingProvider.extendPalette("deep-orange", {
		"A200": rgb2hex($("#accent-color").css("border-top-color")),
		"A100": rgb2hex($("#accent-color").css("border-right-color")),
		"A400": rgb2hex($("#accent-color").css("border-bottom-color")),
		"A700": rgb2hex($("#accent-color").css("border-left-color")),
		"contrastDefaultColor": "dark",
		"contrastLightColors": [],
	}));
	$mdThemingProvider.definePalette("arb-warn-theme", $mdThemingProvider.extendPalette("red", {
		"500": rgb2hex($("#warn-color").css("border-top-color")),
		"300": rgb2hex($("#warn-color").css("border-right-color")),
		"800": rgb2hex($("#warn-color").css("border-bottom-color")),
		"A100": rgb2hex($("#warn-color").css("border-left-color")),
		"contrastDefaultColor": "light",
		"contrastDarkColors": ["300"],
	}));
	// Set the theme
	$mdThemingProvider.theme("default")
	.primaryPalette("arb-primary-theme", {
		"default": "500",
		"hue-1": "300",
		"hue-2": "800",
		"hue-3": "A100",
	})
	.accentPalette("arb-accent-theme", {
		"default": "A200",
		"hue-1": "A100",
		"hue-2": "A400",
		"hue-3": "A700",
	})
	.warnPalette("arb-warn-theme", {
		"default": "500",
		"hue-1": "300",
		"hue-2": "800",
		"hue-3": "A100",
	});

	// Set up custom icons
	$mdIconProvider.icon("arbital_logo", "static/icons/arbital-logo.svg", 40)
		.icon("thumb_up_outline", "static/icons/thumb-up-outline.svg")
		.icon("thumb_down_outline", "static/icons/thumb-down-outline.svg")
		.icon("facebook_box", "static/icons/facebook-box.svg")
		.icon("link_variant", "static/icons/link-variant.svg")
		.icon("comment_plus_outline", "static/icons/comment-plus-outline.svg")
		.icon("format_header_pound", "static/icons/format-header-pound.svg");

	$locationProvider.html5Mode(true);
	// Set up mapping from URL path to specific controllers
	$routeProvider
	.when("/", {
		template: "",
		controller: "IndexPageController",
 		reloadOnSearch: false,
	})
	.when("/adminDashboard/", {
		template: "",
		controller: "AdminDashboardPageController",
	})
	.when("/dashboard/", {
		template: "",
		controller: "DashboardPageController",
	})
	.when("/domains/:alias", {
 		template: "",
 		controller: "DomainPageController",
 		reloadOnSearch: false,
 	})
	.when("/edit/:alias?/:editOrAlias?/:edit?", {
 		template: "",
 		controller: "EditPageController",
 		reloadOnSearch: false,
 	})
	.when("/groups/", {
		template: "",
		controller: "GroupsPageController",
	})
	.when("/learn/:pageAlias?/:pageAlias2?", {
 		template: "",
 		controller: "LearnController",
 		reloadOnSearch: false,
 	})
	.when("/login/", {
		template: "",
		controller: "LoginPageController",
	})
	.when("/p/:alias/:alias2?", {
		template: "",
		controller: "PrimaryPageController",
		reloadOnSearch: false,
	})
	.when("/pages/:alias", {
 		template: "",
 		controller: "RedirectToPrimaryPageController",
 		reloadOnSearch: false,
 	})
	.when("/requisites/", {
		template: "",
		controller: "RequisitesPageController",
	})
	.when("/settings/", {
		template: "",
		controller: "SettingsPageController",
	})
	.when("/signup/", {
		template: "",
		controller: "SignupPageController",
	})
	.when("/updates/", {
		template: "",
		controller: "UpdatesPageController",
	})
	.when("/user/:alias/:alias2?", {
 		template: "",
 		controller: "UserPageController",
 		reloadOnSearch: false,
	})
});

// ArbitalCtrl is used across all pages.
// NOTE: we need to include popoverService, so that it can initialize itself
app.controller("ArbitalCtrl", function ($scope, $location, $timeout, $interval, $http, $compile, $anchorScroll, $mdDialog, userService, pageService, popoverService) {
	$scope.pageService = pageService;
	$scope.userService = userService;
	$scope.loadingBarValue = 0;

	// Get subdomain if any
	$scope.subdomain = undefined;
	var subdomainMatch = /^([A-Za-z0-9_]+)\.(localhost|arbital\.com)\/?$/.exec($location.host());
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
				var $loadingBar = $("#loading-bar");
				$loadingBar.show();
				$scope.loadingBarValue = 0;
				var startTime = (new Date()).getTime();

				var showEverything = function() {
					$interval.cancel(revealInterval);
					$timeout.cancel(revealTimeout);
					// Do short timeout to prevent some rendering bugs that occur on edit page
					$timeout(function() {
						result.element.removeClass("reveal-after-render-parent");
						$loadingBar.hide();
						$anchorScroll();
					}, 50);
				};

				var revealInterval = $interval(function() {
					var timePassed = ((new Date()).getTime() - startTime) / 1000;
					$scope.loadingBarValue = Math.min(100, timePassed * 30);
					var hiddenChildren = result.element.find(".reveal-after-render");
					if (hiddenChildren.length > 0) {
						hiddenChildren.each(function() {
							if ($(this).children().length > 0) {
								$(this).removeClass("reveal-after-render");
							}
						});
						return;
					}
					showEverything();
				}, 50);
				// Do a timeout as well, just in case we have a buggy element
				var revealTimeout = $timeout(function() {
					console.error("Forced reveal timeout");
					showEverything();
				}, 1000);

				$("[ng-view]").append(result.element);
			}

			$("body").toggleClass("body-fix", !result.removeBodyFix);

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
	};

	// Watch path changes and update Google Analytics
	$scope.$watch(function() {
		return $location.absUrl();
	}, function() {
		ga("send", "pageview", $location.absUrl());
	});
});

// simpleDateTime filter converts our typical date&time string into local time.
app.filter("simpleDateTime", function() {
	return function(input) {
		return moment.utc(input).local().format("LT, l");
	};
});

// relativeDateTime converts date&time into a relative string, e.g. "5 days ago"
app.filter("relativeDateTime", function() {
	return function(input) {
		if (moment.utc().diff(moment.utc(input), 'days') <= 7) {
			return moment.utc(input).fromNow();
		} else {
			return moment.utc(input).local().format("MMM Do, YYYY [at] LT");
		}
	};
});
app.filter("relativeDateTimeNoSuffix", function() {
	return function(input) {
		return moment.utc(input).fromNow(true);
	};
});

// numSuffix filter converts a number string to a 2 digit number with a suffix, e.g. K, M, G
app.filter("numSuffix", function() {
	return function(input) {
		var num = +input;
		if (num >= 100000) return (Math.round(num / 100000) / 10) + "M";
		if (num >= 100) return (Math.round(num / 100) / 10) + "K";
		return input;
	};
});

// shorten filter shortens a string to the given number of characters
app.filter("shorten", function() {
	return function(input, charCount) {
		if (!input || input.length <= charCount) return input;
		var s = input.substring(0, charCount);
		var lastSpaceIndex = s.lastIndexOf(" ");
		if (lastSpaceIndex < 0) return s + "...";
		return input.substring(0, lastSpaceIndex) + "...";
	};
});


app.controller("IndexPageController", function ($scope, $routeParams, $http, $compile, pageService, userService) {
	if ($scope.subdomain) {
		// Get the private domain index page data
		$http({method: "POST", url: "/json/domainPage/", data: JSON.stringify({})})
		.success($scope.getSuccessFunc(function(data){
			$scope.indexPageIdsMap = data.result;
			return {
				title: pageService.pageMap[$scope.subdomain].title + " - Private Domain",
				element: $compile("<arb-group-index group-id='" + data.result.domainId +
					"' ids-map='::indexPageIdsMap'></arb-group-index>")($scope),
			};
		}))
		.error($scope.getErrorFunc("domainPage"));
	} else {
		// Get the index page data
		$http({method: "POST", url: "/json/index/"})
		.success($scope.getSuccessFunc(function(data){
			$scope.featuredDomains = data.result.featuredDomains;
			return {
				title: "",
				element: $compile("<arb-index featured-domains='::featuredDomains'></arb-index>")($scope),
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
				"' ids-map='::indexPageIdsMap'></arb-group-index>")($scope),
		};
	}))
	.error($scope.getErrorFunc("domainPage"));
});

app.controller("RedirectToPrimaryPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	// Get the primary page data
	var postData = {
		pageAlias: $routeParams.alias,
	};
	$http({method: "POST", url: "/json/redirectToPrimaryPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		var pageId = data;
		if (!pageId) {
			return {
				title: "Not Found",
				error: "Page doesn't exist, was deleted, or you don't have permission to view it.",
			};
		}
		// Redirect to the primary page, but preserve all search variables
		var search = $location.search();
		$location.replace().url(pageService.getPageUrl(pageId));
		for (var k in search) {
			$location.search(k, search[k]);
		}
		return {
		};
	}))
	.error($scope.getErrorFunc("redirectToPrimaryPage"));
});

app.controller("PrimaryPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
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

		if (page.isLens() || page.isComment() || page.isAnswer()) {
			// Redirect to the primary page, but preserve all search variables
			var search = $location.search();
			$location.replace().url(pageService.getPageUrl(page.pageId));
			for (var k in search) {
				$location.search(k, search[k]);
			}
			return {};
		}

		pageService.ensureCanonUrl(pageService.getPageUrl(page.pageId));
		pageService.primaryPage = page;
		return {
			title: page.title,
			element: $compile("<arb-primary-page></arb-primary-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("primaryPage"));
});

app.controller("LearnController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	// Get the primary page data
	var postData = {
		pageAliases: [],
	};
	var continueLearning = false;
	if ($routeParams.pageAlias) {
		postData.pageAliases.push($routeParams.pageAlias);
	} else if ($location.search().path) {
		postData.pageAliases = postData.pageAliases.concat($location.search().path.split(","));
	} else if (pageService.path) {
		postData.pageAliases = pageService.path.pageIds;
		continueLearning = true;
	}

	$http({method: "POST", url: "/json/learn/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		var primaryPage = undefined;
		if ($routeParams.pageAlias) {
			primaryPage = pageService.pageMap[$routeParams.pageAlias];
			pageService.ensureCanonUrl("/learn/" + primaryPage.alias);
		}
		// Convert all aliases to ids
		$scope.pageIds = [];
		for (var n = 0; n < postData.pageAliases.length; n++) {
			var page = pageService.pageMap[postData.pageAliases[n]];
			if (page) {
				$scope.pageIds.push(page.pageId);
			}
		}
		$scope.tutorMap = data.result.tutorMap;
		$scope.requirementMap = data.result.requirementMap;
		return {
			title: "Learn " + (primaryPage ? primaryPage.title : ""),
			element: $compile("<arb-learn-page continue-learning='::" + continueLearning +
				"' page-ids='::pageIds'" +
				" tutor-map='::tutorMap' requirement-map='::requirementMap'" +
				"></arb-learn-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("learn"));
});

app.controller("EditPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var pageId = $routeParams.alias;

	// Need to call /default/ in case we are creating a new page
	// TODO(alexei): have /newPage/ return /default/ data along with /edit/ data
	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		if (pageId && pageId.charAt(0) > '0' && pageId.charAt(0) <= '9') {
			var specificEdit = 0;
			if ($routeParams.edit) {
				specificEdit = +$routeParams.edit ? +$routeParams.edit : 0;
			} else if (+$routeParams.editOrAlias) {
				specificEdit = +$routeParams.editOrAlias ? +$routeParams.editOrAlias : 0;
			}
			// Load the last edit
			pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: specificEdit,
				success: $scope.getSuccessFunc(function() {
					var page = pageService.editMap[pageId];
					if ($location.search().alias) {
						// Set page's alias
						page.alias = $location.search().alias;
						$location.replace().search("alias", undefined);
					}

					pageService.ensureCanonUrl(pageService.getEditPageUrl(pageId));
					pageService.primaryPage = page;

					// Called when the user is done editing the page.
					$scope.doneFn = function(result) {
						var page = pageService.editMap[result.pageId];
						if (!page.wasPublished && result.discard) {
							$location.path("/edit/");
						} else {
							$location.url(pageService.getPageUrl(page.pageId));
						}
					};
					return {
						removeBodyFix: true,
						title: "Edit " + (page.title ? page.title : "New Page"),
						element: $compile("<arb-edit-page class='full-height' page-id='" + pageId +
							"' done-fn='doneFn(result)' layout='column'></arb-edit-page>")($scope),
					};
				}),
				error: $scope.getErrorFunc("edit"),
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
	.error($scope.getErrorFunc("default"));
});

app.controller("UserPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var userAlias = $routeParams.alias;
	var postData = {
		userAlias: userAlias,
	};
	// Get the data
	$http({method: "POST", url: "/json/userPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		var page = pageService.pageMap[postData.userAlias];
		if (!page) {
			return {
				title: "Not Found",
				error: "User doesn't exist.",
			};
		}

		var userId = page.pageId;
		pageService.ensureCanonUrl(pageService.getUserUrl(userId));
		$scope.userPageIdsMap = data.result;
		return {
			title: userService.userMap[userId].firstName + " " + userService.userMap[userId].lastName,
			element: $compile("<arb-user-page user-id='" + userId + "' ids-map='::userPageIdsMap'></arb-user-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("userPage"));
});

app.controller("DashboardPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var postData = { };
	// Get the data
	$http({method: "POST", url: "/json/dashboardPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		$scope.dashboardPageIdsMap = data.result;
		return {
			title: "Your dashboard",
			element: $compile("<arb-dashboard-page ids-map='::dashboardPageIdsMap'></arb-dashboard-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("dashboardPage"));
});

app.controller("AdminDashboardPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var postData = { };
	// Get the data
	$http({method: "POST", url: "/json/adminDashboardPage/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		$scope.adminDashboardData = data.result;
		return {
			title: "Admin dashboard",
			element: $compile("<arb-admin-dashboard-page data='::adminDashboardData'></arb-admin-dashboard-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("adminDashboardPage"));
});

app.controller("UpdatesPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	var postData = { };
	// Get the explore data
	$http({method: "POST", url: "/json/updates/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		$scope.updateGroups = data.result.updateGroups;
		return {
			title: "Updates",
			element: $compile("<arb-updates update-groups='::updateGroups'></arb-updates>")($scope),
		};
	}))
	.error($scope.getErrorFunc("updates"));
});

app.controller("GroupsPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/groups/"})
	.success($scope.getSuccessFunc(function(data){
		return {
			title: "Groups",
			element: $compile("<arb-groups-page></arb-groups-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("groups"));
});

app.controller("SignupPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		if (userService.user.id) {
			window.location.href = "http://" + window.location.host;
		}
		return {
			title: "Sign Up",
			element: $compile("<arb-signup></arb-signup>")($scope),
		};
	}))
	.error($scope.getErrorFunc("default"));
});

app.controller("LoginPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		if (userService.user.id) {
			window.location.href = "http://" + window.location.host;
		}
		return {
			title: "Log In",
			element: $compile("<div class='md-whiteframe-1dp capped-body-width'><arb-login></arb-login></div>")($scope),
		};
	}))
	.error($scope.getErrorFunc("default"));
});

app.controller("RequisitesPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/requisites/"})
	.success($scope.getSuccessFunc(function(data){
		return {
			title: "Requisites",
			element: $compile("<arb-requisites-page></arb-requisites-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("requisites"));
});

app.controller("SettingsPageController", function ($scope, $routeParams, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		return {
			title: "Settings",
			element: $compile("<arb-settings-page></arb-settings-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("default"));
});
