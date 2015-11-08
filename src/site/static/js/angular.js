"use strict";

// Set up angular module.
var app = angular.module("arbital", ["ngResource", "ui.bootstrap", "RecursionHelper"]);
app.config(function($interpolateProvider, $locationProvider, $provide){
	$interpolateProvider.startSymbol("{[{").endSymbol("}]}");

	$locationProvider.html5Mode({
		enabled: true,
		requireBase: false,
		rewriteLinks: false
	});
});

// ArbitalCtrl is used across all pages.
app.controller("ArbitalCtrl", function ($scope, $location, $timeout, $http, $compile, userService, pageService) {
	$scope.pageService = pageService;
	$scope.userService = userService;

	// Process last visit url parameter
	var lastVisit = $location.search().lastVisit;
	if (lastVisit) {
		$("body").attr("last-visit", lastVisit);
		$location.search("lastVisit", null);
	}

	// Refresh all the fields that need to be updated every so often.
	var refreshAutoupdates = function() {
		$(".autoupdate").each(function(index, element) {
			$compile($(element))($scope);
		});
		$timeout(refreshAutoupdates, 30000);
	};
	refreshAutoupdates();

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

	// Handle intrasite link clicks.
	/*$("body").on("click", ".intrasite-link", function(event) {
		var $link = $(event.target);
		if ($link.attr("href") !== "#") return true;
		$(".dynamic-body").empty();
		pageService.pageMap = {};
		pageService.userMap = {};
		loadPrimaryPage($link.attr("page-id"));
	});*/

	// ========== Smart loading ==============
	// Here we check the url to dynamically load the necessary data.
	
	// Get subdomain if any
	var subdomain = undefined;
	var subdomainMatch = /^([A-Za-z0-9]+)\.(localhost|arbital\.com)\/?$/.exec($location.host());
	if (subdomainMatch) {
		subdomain = subdomainMatch[1];
	}

	// Returns a function we can use as success handler for POST requests for dynamic data.
	// callback - returns {
	//   title: title to set for the window
	//   element: optional jQuery element to add dynamically to the body
	//   error: optional error message to print
	// }
	var getSuccessFunc = function(callback) {
		return function(data) {
			// Sometimes we don't get data.
			if (data) {
				console.log("Dynamic request data:"); console.log(data);
				userService.processServerData(data);
				pageService.processServerData(data);
			}

			// Because the subdomain could have any case, we need to find the alias
			// in the loaded map so we can get the alias with correct case
			if (subdomain) {
				for (var pageAlias in pageService.pageMap) {
					if (subdomain.toUpperCase() === pageAlias.toUpperCase()) {
						subdomain = pageAlias;
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
				$(".dynamic-body").append(result.element);
				$compile($(".dynamic-body"))($scope);
			}
			if (result.title) {
				document.title = result.title + " - Arbital";
			}
		};
	};

	// Returns a function we can use as error handler for POST requests for dynamic data.
	var getErrorFunc = function(urlPageType) {
		return function(data, status){
			console.log("Error /json/" + urlPageType + "/:"); console.log(data); console.log(status);
			$(".global-error").text(data).show();
			document.title = "Error - Arbital";
		};
	}

	// Primary page
	var loadPrimaryPage = function(pageId) {
		// Get the primary page data
		var postData = {
			pageAlias: pageId,
			forcedLastVisit: lastVisit,
		};
		$http({method: "POST", url: "/json/primaryPage/", data: JSON.stringify(postData)})
		.success(getSuccessFunc(function(data){
			var page = pageService.pageMap[pageId];
			if (page) {
				$scope.page = page;
				pageService.setPrimaryPage(page);

				return {
					title: $scope.page.title,
					element: $("<arb-primary-page></arb-primary-page>")
				};
			}
			return {
				title: "Not Found",
				error: "Page doesn't exist, was deleted, or you don't have permission to view it.",
			};
		}))
		.error(getErrorFunc("primaryPage"));
	};
	var pagesPath = /^\/pages\/([0-9]+)\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		loadPrimaryPage(match[1]);
	}
	
	// Edit page
	var editPagePath = /^\/edit\/?([0-9]*)\/?$/;
	var match = editPagePath.exec($location.path());
	if (match) {
		var pageId = match[1];

		// Call this when pageId is determined and page is loaded.
		var createEditPage = getSuccessFunc(function() {
			var page = pageService.editMap[pageId];
			pageService.setPrimaryPage(page);

			// Called when the user is done editing the page.
			$scope.doneFn = function(result) {
				if (pageService.primaryPage.wasPublished || !result.abandon) {
					window.location.href = pageService.primaryPage.url();
				} else {
					window.location.href = "/edit/";
				}
			};
			return {
				title: "Edit " + (page.title ? page.title : "New Page"),
				element: $("<arb-edit-page page-id='" + pageId + "' done-fn='doneFn(result)'></arb-edit-page>"),
			};
		});

		$http({method: "POST", url: "/json/default/"})
		.success(getSuccessFunc(function(data){
			if (pageId) {
				// Load the last edit
				pageService.loadEdit({
					pageAlias: pageId,
					specificEdit: $location.search().edit,
					success: createEditPage,
					error: getErrorFunc("loadEdit"),
				});
			} else {
				// Create a new page to edit
				pageService.getNewPage({
					success: function(newPageId) {
						pageId = newPageId;
						var aliasParam = $location.search().alias;
						if (aliasParam) {
							pageService.editMap[pageId].alias = aliasParam;
						}
						$location.path("/edit/" + pageId);
						createEditPage();
					},
				});
			}
			return {
				title: "Edit Page",
			};
		}))
		.error(getErrorFunc("Edit Page"));
	}

	// Domain page
	var pagesPath = /^\/domains\/([A-Za-z0-9]+)\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		pageService.domainAlias = match[1];
		var postData = {
			domainAlias: pageService.domainAlias,
		};
		// Get the domain index page data
		$http({method: "POST", url: "/json/domainPage/", data: JSON.stringify(postData)})
		.success(getSuccessFunc(function(data){
			$scope.indexPageIdsMap = data.result;
			return {
				title: pageService.pageMap[pageService.domainAlias].title,
				element: $("<arb-group-index ids-map='indexPageIdsMap'></arb-group-index>"),
			};
		}))
		.error(getErrorFunc("privateIndex"));
	}

	// Explore page
	var pagesPath = /^\/explore\/?([A-Za-z0-9]*)\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		var postData = {
			groupAlias: subdomain ? subdomain : match[1],
		};
		// Get the explore data
		$http({method: "POST", url: "/json/explore/", data: JSON.stringify(postData)})
		.success(getSuccessFunc(function(data){
			// Decide on the domain alias
			var title;
			if (subdomain) {
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
				element: $("<arb-page-tree init-map='rootPages' additional-map='childPages'" +
					"supersize-roots='true'></arb-page-tree>"),
			};
		}))
		.error(getErrorFunc("explore"));
	}

	// User page
	var userPagePath = /^\/user\/?([A-Za-z0-9]*)\/?$/;
	var match = userPagePath.exec($location.path());
	if (match) {
		var userId = match[1];
		if (!userId) {
			userId = userService.user.id;
		}
		var postData = {
			userId: userId,
		};
		// Get the data
		$http({method: "POST", url: "/json/userPage/", data: JSON.stringify(postData)})
		.success(getSuccessFunc(function(data){
			$scope.userPageIdsMap = data.result;
			return {
				title: userService.userMap[userId].firstName + " " + userService.userMap[userId].lastName,
				element: $("<arb-user-page ids-map='userPageIdsMap'></arb-user-page>"),
			};
		}))
		.error(getErrorFunc("Explore"));
	}

	// Updates page
	var pagesPath = /^\/updates\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		var postData = { };
		// Get the explore data
		$http({method: "POST", url: "/json/updates/", data: JSON.stringify(postData)})
		.success(getSuccessFunc(function(data){
			$scope.updateGroups = data.result.updateGroups;
			return {
				title: "Updates",
				element: $("<arb-updates update-groups='updateGroups'></arb-updates>"),
			};
		}))
		.error(getErrorFunc("Updates"));
	}

	// Groups page
	var pagesPath = /^\/groups\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		$http({method: "POST", url: "/json/groups/"})
		.success(getSuccessFunc(function(data){
			return {
				title: "Groups",
				element: $("<arb-groups-page></arb-groups-page>"),
			};
		}))
		.error(getErrorFunc("groups"));
	}

	// Signup page
	var pagesPath = /^\/signup\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		$http({method: "POST", url: "/json/default/"})
		.success(getSuccessFunc(function(data){
			return {
				title: "Sign Up",
				element: $("<arb-signup></arb-signup>"),
			};
		}))
		.error(getErrorFunc("Settings"));
	}

	// Settings page
	var pagesPath = /^\/settings\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		$http({method: "POST", url: "/json/default/"})
		.success(getSuccessFunc(function(data){
			return {
				title: "Settings",
				element: $("<arb-settings-page></arb-settings-page>"),
			};
		}))
		.error(getErrorFunc("Settings"));
	}

	// Index page
	var pagesPath = /^\/$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		if (subdomain) {
			// Get the private group index page data
			$http({method: "POST", url: "/json/privateIndex/"})
			.success(getSuccessFunc(function(data){
				$scope.indexPageIdsMap = data.result;
				return {
					title: pageService.pageMap[subdomain].title + " - Private Group",
					element: $("<arb-group-index ids-map='indexPageIdsMap'></arb-group-index>"),
				};
			}))
			.error(getErrorFunc("privateIndex"));
		} else {
			// Get the index page data
			$http({method: "POST", url: "/json/index/"})
			.success(getSuccessFunc(function(data){
				$scope.featuredDomains = data.result.featuredDomains;
				return {
					element: $("<arb-index featured-domains='featuredDomains'></arb-index>"),
				};
			}))
			.error(getErrorFunc("index"));
		}
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

