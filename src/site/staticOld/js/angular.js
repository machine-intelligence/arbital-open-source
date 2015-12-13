"use strict";

// Set up angular module.
var app = angular.module("arbital", ["ngResource", "ui.bootstrap", "RecursionHelper", "ngRoute"]);
app.config(function($interpolateProvider, $locationProvider, $provide, $routeProvider){
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
	.when("/explore/:alias?", {
		template: "",
		controller: "ExplorePageController",
	})
	.when("/pages/:alias", {
		template: "",
		controller: "PrimaryPageController",
	})
	.when("/edit/:alias?", {
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
app.controller("ArbitalCtrl", function ($scope, $location, $timeout, $http, $compile, userService, pageService) {
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

	// Process lens tab clicks
	$("body").on("click", ".intrasite-lens-tab", function(event) {
		var $tab = $(event.currentTarget);
		var lensId = $tab.attr("data-target");
		lensId = lensId.substring(lensId.indexOf("-") + 1);
		var lensPage = pageService.pageMap[lensId];
		if (!lensPage) return;
		if (lensPage.summary.length > 0) {
			var page = pageService.pageMap[lensId];
			var lensElement = $("#lens-" + page.pageId);
			lensElement.empty().append('<div class="markdown-text"></div>');
			arbMarkdown.init(false, page.pageId, page.summary, lensElement, pageService, userService);
		}
		return true;
	});

	// Check when user hovers over intrasite links, and show a popover.
	$("body").on("mouseenter", ".intrasite-link", function(event) {
		var $target = $(event.currentTarget);
		if ($target.hasClass("red-link")) return;
		// Don't allow recursive hover in popovers.
		if ($target.closest(".popover-content").length > 0) return;

		// Popover's title.
		var getTitleHtml = function(pageId) {
			return "<arb-likes-page-title is-search-result='true' page-id='" + pageId + "'></arb-likes-page-title>";
		};
		// Create options for the popover.
		var options = {
			html : true,
			placement: "bottom",
			trigger: "manual",
			delay: { "hide": 100 },
			title: function() {
				var pageId = $target.attr("page-id");
				var page = pageService.pageMap[pageId];
				if (page && page.title) {
					return getTitleHtml(pageId);
				}
				return "Loading...";
			},
			content: function() {
				var $link = $target;
				var setPopoverContent = function(page) {
					$timeout(function() {
						var $popover = $("#" + $link.attr("aria-describedby"));
						$popover.find(".popover-title").html(getTitleHtml(page.pageId));
						$compile($popover)($scope);
					});
					var contentHtml = "<arb-intrasite-popover page-id='" + page.pageId + "'></arb-intrasite-popover>";
					return contentHtml;
				};

				// Check if we already have this page cached.
				var pageAlias = $link.attr("page-id");
				var page = pageService.pageMap[pageAlias];
				if (page && page.summary) {
					return setPopoverContent(page);
				}

				// Fetch page data from the server.
				pageService.loadIntrasitePopover(pageAlias, {
					success: function() {
						var page = pageService.pageMap[pageAlias];
						if (!page.summary) {
							page.summary = " "; // to avoid trying to load it again
						}
						var contentHtml = setPopoverContent(page);
						var $popover = $("#" + $link.attr("aria-describedby"));
						$popover.find(".popover-content").html(contentHtml);
					},
				});
				return "<img src='/static/images/loading.gif' class='loading-indicator' style='display:block'/>";
			}
		};
		// Check if this is the first time we hovered.
		var firstTime = $target.attr("first-time");
		if (!firstTime) {
			createHoverablePopover($target, options, {uniqueName: "intrasite-link"});
			$target.attr("first-time", false).trigger("mouseenter");
		}
		return false;
	});

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
			var result = callback(data);
			if (result.error) {
				$(".global-error").text(result.error).show();
				document.title = "Error - Arbital";
			}
			if (result.element) {
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
			console.log("Error /json/" + urlPageType + "/:"); console.log(data); console.log(status);
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
		return {
			title: pageService.pageMap[pageService.domainAlias].title,
			element: $compile("<arb-group-index ids-map='indexPageIdsMap'></arb-group-index>")($scope),
		};
	}))
	.error($scope.getErrorFunc("domainPage"));
});

app.controller("ExplorePageController", function ($scope, $routeParams, $http, $compile, pageService, userService) {
	var postData = {
		groupAlias: $scope.subdomain ? $scope.subdomain : $routeParams.alias,
	};
	// Get the explore data
	$http({method: "POST", url: "/json/explore/", data: JSON.stringify(postData)})
	.success($scope.getSuccessFunc(function(data){
		// Decide on the domain alias
		var title;
		if ($scope.subdomain) {
			title = pageService.pageMap[pageService.privateGroupId].title + " - Explore";
		} else {
			pageService.domainAlias = postData.groupAlias;
			title = pageService.pageMap[pageService.domainAlias].title + " - Explore";
		}

		// Compute root and children maps
		var rootPage = pageService.pageMap[data.result.rootPageId];
		$scope.rootPages = {};
		$scope.rootPages[rootPage.pageId] = rootPage;
		$scope.childPages = {};
		var length = rootPage.children ? rootPage.children.length : 0;
		for (var n = 0; n < length; n++) {
			var childId = rootPage.children[n].childId;
			$scope.childPages[childId] = pageService.pageMap[childId];
		}

		return {
			title: title,
			element: $compile("<arb-page-tree init-map='rootPages' additional-map='childPages'" +
				"supersize-roots='true'></arb-page-tree>")($scope),
		};
	}))
	.error($scope.getErrorFunc("explore"));
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
		pageService.setPrimaryPage(page);
		return {
			title: page.title,
			element: $compile("<arb-primary-page></arb-primary-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("primaryPage"));
});


app.controller("EditPageController", function ($scope, $routeParams, $route, $http, $compile, $location, pageService, userService) {
	var pageId = $routeParams.alias;

	$http({method: "POST", url: "/json/default/"})
	.success($scope.getSuccessFunc(function(data){
		if (+pageId) {
			// Load the last edit
			pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: $location.search().edit,
				success: $scope.getSuccessFunc(function() {
					var page = pageService.editMap[pageId];
					if ($location.search().alias) {
						// Set page's alias
						page.alias = $location.search().alias;
						$location.search("alias", undefined);
					}

					pageService.setPrimaryPage(page);
			
					// Called when the user is done editing the page.
					$scope.doneFn = function(result) {
						if (pageService.primaryPage.wasPublished || !result.abandon) {
							$location.path(pageService.primaryPage.url());
						} else {
							$location.path("/edit/");
						}
						$scope.$apply();
					};
					return {
						title: "Edit " + (page.title ? page.title : "New Page"),
						element: $compile("<arb-edit-page page-id='" + pageId +
							"' done-fn='doneFn(result)'></arb-edit-page>")($scope),
					};
				}),
				error: $scope.getErrorFunc("loadEdit"),
			});
		} else {
			// Create a new page to edit
			pageService.getNewPage({
				success: function(newPageId) {
					// Check if we need to add parents to this new page
					var newParentIdString = $location.search().newParentId;
					$location.search("newParentId", undefined);
					var unfinishedCallbackCount = 0;
					var readyToRedirect = false;
					if (newParentIdString) {
						// Add the parents, and then wait for the server to reply
						var newParentIdList = newParentIdString.split(",");
						for (var key in newParentIdList) {
							var newParentId = newParentIdList[key];
							if (newParentId) {
								unfinishedCallbackCount++;
								// Add a parent for this new page
								pageService.newPagePair({
									parentId: newParentId,
									childId: newPageId,
									type: "parent",
								}, function() {
									unfinishedCallbackCount--;
									if (unfinishedCallbackCount <= 0 && readyToRedirect) {
										$location.replace().path(pageService.getEditPageUrl(newPageId));
									}
								});
							}
						}
						readyToRedirect = true;
						setTimeout(function() {
							if (unfinishedCallbackCount > 0) {
								$location.replace().path(pageService.getEditPageUrl(newPageId));
							}
						}, 1000);
					} else {
						$location.replace().path(pageService.getEditPageUrl(newPageId));
					}
				},
			});
		}
		return {
			title: "Edit Page",
		};
	}))
	.error($scope.getErrorFunc("Edit Page"));
});

app.controller("UserPageController", function ($scope, $routeParams, $route, $http, $compile, $location, pageService, userService) {
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
			element: $compile("<arb-user-page ids-map='userPageIdsMap'></arb-user-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("User"));
});

app.controller("UpdatesPageController", function ($scope, $routeParams, $route, $http, $compile, $location, pageService, userService) {
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

app.controller("GroupsPageController", function ($scope, $routeParams, $route, $http, $compile, $location, pageService, userService) {
	$http({method: "POST", url: "/json/groups/"})
	.success($scope.getSuccessFunc(function(data){
		return {
			title: "Groups",
			element: $compile("<arb-groups-page></arb-groups-page>")($scope),
		};
	}))
	.error($scope.getErrorFunc("Groups"));
});

app.controller("SignupPageController", function ($scope, $routeParams, $route, $http, $compile, $location, pageService, userService) {
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

app.controller("SettingsPageController", function ($scope, $routeParams, $route, $http, $compile, $location, pageService, userService) {
		$http({method: "POST", url: "/json/default/"})
		.success($scope.getSuccessFunc(function(data){
			return {
				title: "Settings",
				element: $compile("<arb-settings-page></arb-settings-page>")($scope),
			};
		}))
		.error($scope.getErrorFunc("Settings"));
});
