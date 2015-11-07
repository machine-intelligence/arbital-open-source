/* angular.tmpl.js is a .tmpl file that is inserted as a <script> into the
	<header> portion of html pages that use angular. It defines the arbital module
	and ArbitalCtrl, which are used on every page. */
{{define "angular"}}
<script>

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

// User service.
app.service("userService", function(){
	var that = this;

	// Logged in user.
	this.user = {{GetCurrentUserJson}};
	this.userMap = {
		{{if .UserMap}}
			{{range $k,$v := .UserMap}}
				"{{$k}}": {{GetUserJson $v}},
			{{end}}
		{{end}}
	};
	console.log("Initial user map:"); console.dir(this.userMap);

	// Check if we can let this user do stuff.
	this.userIsCool = function() {
		return this.user.karma >= 200;
	};

	// Return url to the user page.
	this.getUserUrl = function(userId) {
		return "/user/" + userId;
	};

	// (Un)subscribe a user to a thing.
	var subscribeTo = function(doSubscribe, data, done) {
		$.ajax({
			type: "POST",
			url: doSubscribe ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(done);
	};
	// (Un)subscribe a user to another user.
	this.subscribeToUser = function($target) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			userId: $target.attr("user-id"),
		};
		subscribeTo($target.hasClass("on"), data, function(r) {});
	}
	this.subscribeToPage = function($target) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			pageId: $target.attr("page-id"),
		};
		subscribeTo($target.hasClass("on"), data, function(r) {});
	};

	// Call this to process data we received from the server.
	this.processServerData = function(data) {
		$.extend(that.userMap, data["users"]);
	}
});

// pages stores all the loaded pages and provides multiple helper functions for
// working with pages.
app.service("pageService", function(userService, $http){
	var that = this;

	// All loaded pages.
	this.pageMap = {
		{{range $k,$v := .PageMap}}
			"{{$k}}": {{GetPageJson $v}},
		{{end}}
	};
	
	// All loaded edits. (These are the pages we will be editing.)
	this.editMap = {
		{{range $k,$v := .EditMap}}
			"{{$k}}": {{GetPageJson $v}},
		{{end}}
	};

	// All loaded masteries.
	this.masteryMap = {
		{{range $k,$v := .MasteryMap}}
			"{{$k}}": {{GetMasteryJson $v}},
		{{end}}
	};

	// Update whether on not the user has a mastery.
	this.updateMastery = function(scope, masteryId, has) {
		var mastery = that.masteryMap[masteryId];
		if (!mastery) {
			mastery = {pageId: masteryId};
			that.masteryMap[masteryId] = mastery;
		}
		mastery.has = has;
		mastery.isManuallySet = true;
		scope.$apply();

		// Send POST request.
		var data = {
			masteryId: masteryId,
			has: has,
		};
		$.ajax({
			type: "POST",
			url: "/updateMastery/",
			data: JSON.stringify(data),
		}).fail(function(r) {
			console.log("Failed to claim mastery:"); console.log(r);
		});
	};

	// Primary page is the one that's displayed front and center.
	this.primaryPage = undefined;
	// List of callbacks to notify when primary page changes.
	this.primaryPageCallbacks = [];
	// Set the primary page, triggering the callbacks.
	this.setPrimaryPage = function(newPrimaryPage) {
		var oldPrimaryPage = this.primaryPage;
		this.primaryPage = newPrimaryPage;
		for (var n = 0; n < this.primaryPageCallbacks.length; n++) {
			this.primaryPageCallbacks[n](oldPrimaryPage);
		}
		$("body").attr("last-visit", moment.utc(this.primaryPage.lastVisit).format("YYYY-MM-DD HH:mm:ss"));
	};
	
	// Call this to process data we received from the server.
	this.processServerData = function(data) {
		$.extend(this.masteryMap, data["masteries"]);

		var pageData = data["pages"];
		for (var id in pageData) {
			var page = pageData[id];
			if (page.isCurrentEdit) {
				this.addPageToMap(pageData[id]);
			} else {
				this.addPageToEditMap(pageData[id]);
			}
		}

		var editData = data["edits"];
		for (var id in editData) {
			this.addPageToEditMap(editData[id]);
		}
	}

	this.getPageUrl = function(pageId){
		return "/pages/" + pageId;
	};
	this.getEditPageUrl = function(pageId){
		return "/edit/" + pageId;
	};

	// These functions will be added to each page object.
	var pageFuncs = {
		// Check if the user has never visited this page before.
		isNewPage: function() {
			if (userService.user.id === "0") return false;
			return this.creatorId != userService.user.id &&
				(this.lastVisit === "" || this.originalCreatedAt >= this.lastVisit);
		},
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
			if (userService.user.id === "0") return false;
			return this.creatorId != userService.user.id &&
				this.lastVisit !== "" && this.createdAt >= this.lastVisit && this.lastVisit > this.originalCreatedAt;
		},
		// Return empty string if the user can edit this page. Otherwise a reason for
		// why they can't.
		getEditLevel: function() {
			var karmaReq = 200; // TODO: fix this
			if (userService.user.karma < karmaReq) {
				if (userService.user.isAdmin) {
					// Can edit but only because user is an admin.
					return "admin";
				}
				return "" + karmaReq;
			}
			return "";
		},
		// Return empty string if the user can delete this page. Otherwise a reason
		// for why they can't.
		getDeleteLevel: function() {
			var karmaReq = 200; // TODO: fix this
			if (userService.user.karma < karmaReq) {
				if (userService.user.isAdmin) {
					return "admin";
				}
				return "" + karmaReq;
			}
			return "";
		},
		// Return true iff the page is deleted.
		isDeleted: function() {
			return this.type === "deleted";
		},
		// Get page's url
		url: function() {
			return that.getPageUrl(this.pageId);
		},
		// Get url to edit the page
		editUrl: function() {
			return that.getEditPageUrl(this.pageId);
		},
	};
	
	// Massage page's variables to be easier to deal with.
	var setUpPage = function(page, pageMap) {
		if (page.children == null) page.children = [];
		if (page.parents == null) page.parents = [];
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		// Add page's alias to the map as well
		if (pageMap && page.pageId !== page.alias) {
			pageMap[page.alias] = page;
		}
		return page;
	};
	// Add the given page to the global pageMap. If the page with the same id
	// already exists, we do a clever merge.
	var isPageValueTruthy = function(v) {
		// "0" is falsy
		if (v === "0") {
			return false;
		}
		// Empty array is falsy.
		if ($.isArray(v) && v.length == 0) {
			return false;
		}
		// Empty object is falsy.
		if ($.isEmptyObject(v)) {
			return false;
		}
		return true;
	};
	this.addPageToMap = function(newPage) {
		var oldPage = this.pageMap[newPage.pageId];
		if (newPage === oldPage) return;
		if (oldPage === undefined) {
			this.pageMap[newPage.pageId] = setUpPage(newPage, this.pageMap);
			return;
		}
		// Merge each variable.
		for (var k in oldPage) {
			var oldV = isPageValueTruthy(oldPage[k]);
			var newV = isPageValueTruthy(newPage[k]);
			if (!newV) {
				// No new value.
				continue;
			}
			if (!oldV) {
				// No old value, so use the new one.
				oldPage[k] = newPage[k];
			}
			// Both new and old values are legit. Overwrite with new.
			oldPage[k] = newPage[k];
		}
	};

	// Remove page with the given pageId from the global pageMap.
	this.removePageFromMap = function(pageId) {
		delete this.pageMap[pageId];
	};

	// Add the given page to the global editMap.
	this.addPageToEditMap = function(page) {
		this.editMap[page.pageId] = setUpPage(page);
	}

	// Remove page with the given pageId from the global editMap;
	this.removePageFromEditMap = function(pageId) {
		delete this.editMap[pageId];
	};

	// Load children for the given page. Success/error callbacks are called only
	// if request was actually made to the server.
	this.loadChildren = function(parent, success, error) {
		var that = this;
		if (parent.hasLoadedChildren) {
			success(parent.loadChildrenData, 200);
			return;
		} else if (parent.isLoadingChildren) {
			return;
		}
		parent.isLoadingChildren = true;
		console.log("Issuing POST request to /json/children/?parentId=" + parent.pageId);
		$http({method: "POST", url: "/json/children/", data: JSON.stringify({parentId: parent.pageId})}).
			success(function(data, status){
				parent.isLoadingChildren = false;
				parent.hasLoadedChildren = true;
				userService.processServerData(data);
				that.processServerData(data);
				parent.loadChildrenData = data["pages"];
				success(data["pages"], status);
			}).error(function(data, status){
				parent.isLoadingChildren = false;
				console.log("Error loading children:"); console.log(data); console.log(status);
				error(data, status);
			});
	};

	// Return function for sorting children ids.
	this.getChildSortFunc = function(sortChildrenBy) {
		var pageMap = this.pageMap;
		if(sortChildrenBy === "alphabetical") {
			return function(aId, bId) {
				var aTitle = pageMap[aId].title;
				var bTitle = pageMap[bId].title;
				// If title starts with a number, we want to compare those numbers directly,
				// otherwise "2" comes after "10".
				var aNum = parseInt(aTitle);
				if (aNum) {
					var bNum = parseInt(bTitle);
					if (bNum) {
						return aNum - bNum;
					}
				}
				return pageMap[aId].title.localeCompare(pageMap[bId].title);
			};
		} else if (sortChildrenBy === "recentFirst") {
			return function(aId, bId) {
				return pageMap[bId].originalCreatedAt.localeCompare(pageMap[aId].originalCreatedAt);
			};
		} else if (sortChildrenBy === "oldestFirst") {
			return function(aId, bId) {
				return pageMap[aId].originalCreatedAt.localeCompare(pageMap[bId].originalCreatedAt);
			};
		} else {
			if (sortChildrenBy !== "likes") {
				console.error("Unknown sort type: " + sortChildrenBy);
				console.log(page);
			}
			return function(aId, bId) {
				var diff = pageMap[bId].likeCount - pageMap[aId].likeCount;
				if (diff === 0) {
					return pageMap[aId].title.localeCompare(pageMap[bId].title);
				}
				return diff;
			};
		}
	};
	// Sort the given page's children.
	this.sortChildren = function(page) {
		var sortFunc = this.getChildSortFunc(page.sortChildrenBy);
		page.children.sort(function(aChild, bChild) {
			return sortFunc(aChild.childId, bChild.childId);
		});
	};

	// Load parents for the given page. Success/error callbacks are called only
	// if request was actually made to the server.
	this.loadParents = function(child, success, error) {
		var that = this;
		if (child.hasLoadedParents) {
			success(child.loadParentsData, 200);
			return;
		} else if (child.isLoadingParents) {
			return;
		}
		child.isLoadingParents = true;
		console.log("Issuing POST request to /json/parents/?childId=" + child.pageId);
		$http({method: "POST", url: "/json/parents/", data: JSON.stringify({childId: child.pageId})}).
			success(function(data, status){
				child.isLoadingParents = false;
				child.hasLoadedParents = true;
				userService.processServerData(data);
				that.processServerData(data);
				child.loadParentsData = data["pages"];
				success(data["pages"], status);
			}).error(function(data, status){
				child.isLoadingParents = false;
				console.log("Error loading parents:"); console.log(data); console.log(status);
				error(data, status);
			});
	};

	// Load the page with the given pageAlias.
	// options {
	//	 url: url to call
	//   success: callback on success
	//   error: callback on error
	// }
	// Track which pages we are already loading. Map url+pageAlias -> true.
	var loadingPageAliases = {};
	var loadPage = function(pageAlias, options) {
		// Check if the page is already being loaded, and mark it as such if it's not.
		var loadKey = options.url + pageAlias;
		if (loadKey in loadingPageAliases) {
			return;
		}
		loadingPageAliases[loadKey] = true;

		console.log("Issuing a POST request to: " + options.url + "?pageAlias=" + pageAlias);
		$http({method: "POST", url: options.url, data: JSON.stringify({pageAlias: pageAlias})}).
			success(function(data, status){
				console.log("JSON " + options.url + " data:"); console.dir(data);
				userService.processServerData(data);
				that.processServerData(data);
				var pageData = data["pages"];
				for (var id in pageData) {
					delete loadingPageAliases[options.url + id];
					delete loadingPageAliases[options.url + pageData[id].alias];
				}
				if(options.success) options.success();
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(options.error) options.error(data, status);
			}
		);
	};

	// Get data to display a popover for the page with the given alias.
	this.loadIntrasitePopover = function(pageAlias, options) {
		options.url = "/json/intrasitePopover/";
		loadPage(pageAlias, options);
	};

	// Get data to display a lens.
	this.loadLens = function(pageAlias, options) {
		options.url = "/json/lens/";
		loadPage(pageAlias, options);
	};
	
	// Load edit.
	// options {
	//   pageAlias: pageAlias to load
	//   specificEdit: load page with this edit number
	//	 editLimit: only load edits lower than this number
	//	 createdAtLimit: only load edits that were created before this date
	//	 skipProcessDataStep: if true, we don't process the data we get from the server
	//   success: callback on success
	//   error: callback on error
	// }
	this.loadEdit = function(options) {
		// Set up options.
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;
		var skipProcessDataStep = options.skipProcessDataStep; delete options.skipProcessDataStep;

		console.log("Issuing a POST request to: /json/edit/?pageAlias=" + options.pageAlias);
		$http({method: "POST", url: "/json/edit/", data: JSON.stringify(options)}).
			success(function(data, status){
				console.log("JSON /json/edit/ data:"); console.dir(data);
				if (!skipProcessDataStep) {
					userService.processServerData(data);
					that.processServerData(data);
				}
				if(success) success(data["edits"], status);
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

	// Get a new page from the server.
	// options {
	//	success: callback on success
	//}
	this.getNewPage = function(options) {
		$http({method: "POST", url: "/json/newPage/"}).
			success(function(data, status){
				console.log("JSON /json/newPage/ data:"); console.dir(data);
				var pageId = Object.keys(data["pages"])[0];
				that.processServerData(data);
				if(options.success) options.success(pageId);
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(error) error(data, status);
			});
	}

	// Delete the page with the given pageId.
	this.deletePage = function(pageId, success, error) {
		var data = {
			pageId: pageId,
		};
		$http({method: "POST", url: "/deletePage/", data: JSON.stringify(data)})
			.success(function(data, status){
				console.log("Successfully deleted " + pageId);
				if(success) success(data, status);
			})
			.error(function(data, status){
				console.log("Error deleting " + pageId + ":"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

	// Abandon the page with the given id.
	this.abandonPage = function(pageId, success, error) {
		var data = {
			pageId: pageId,
		};
		$http({method: "POST", url: "/abandonPage/", data: JSON.stringify(data)}).
			success(function(data, status){
				console.log("Successfully abandoned " + pageId);
				if(success) success(data, status);
			})
			.error(function(data, status){
				console.log("Error abandoning " + pageId + ":"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

	// Add a new relationship between pages using the given options.
	// options = {
	//	parentId: id of the parent page
	//	childId: id of the child page
	//	type: type of the relationships
	// }
	this.newPagePair = function(options, success) {
		$http({method: "POST", url: "/newPagePair/", data: JSON.stringify(options)})
			.success(function(data, status){
				if(success) success();
			})
			.error(function(data, status){
				console.log("Error creating new page pair:"); console.log(data); console.log(status);
			});
	};
	// Note: you also need to specify the type of the relationship here, sinc we
	// don't want to accidentally delete the wrong type.
	this.deletePagePair = function(options) {
		$http({method: "POST", url: "/deletePagePair/", data: JSON.stringify(options)})
			.error(function(data, status){
				console.log("Error deleting a page pair:"); console.log(data); console.log(status);
			});
	};

	// TODO: make these into page functions?
	// Return true iff we should show that this page is public.
	this.showPublic = function(pageId) {
		/*if (this.privateGroupId !== undefined) {
			return this.privateGroupId !== this.pageMap[pageId].seeGroupId;
		}*/
		var page = this.pageMap[pageId];
		if (!this.primaryPage) return false;
		return this.primaryPage.seeGroupId !== page.seeGroupId && page.seeGroupId === "0";
	};
	// Return true iff we should show that this page belongs to a group.
	this.showLockedGroup = function(pageId) {
		var page = this.pageMap[pageId];
		if (!this.primaryPage) return page.seeGroupId !== "0";
		return this.primaryPage.seeGroupId !== page.seeGroupId && page.seeGroupId !== "0";
	};

	// Setup all initial pages.
	console.log("Initial pageMap: "); console.dir(this.pageMap);
	for (var id in this.pageMap) {
		setUpPage(this.pageMap[id], this.pageMap);
	}
});

// Autocomplete service provides data for autocompletion.
app.service("autocompleteService", function($http, $compile, pageService){
	var that = this;

	// Set how to render search results for the given autocomplete input.
	this.setAutocompleteRendering = function($input, scope, resultsAreLinks) {
		$input.data("ui-autocomplete")._renderItem = function(ul, item) {
			var elementType = "span";
			var elementTypeEnd = "span";
			if (resultsAreLinks) {
				elementType = "a href='" + pageService.getPageUrl(item.label) + "'";
				elementTypeEnd = "a";
			}
			var $el = $compile("<li class='ui-menu-item'><" + elementType +
				" arb-likes-page-title class='search-result' page-id='" + item.value +
				"' show-clickbait='true' is-search-result='true'></" + elementTypeEnd + "></li>")(scope);
			$el.attr("data-value", item.value);
			return $el.appendTo(ul);
		};
	};

	// Take data we get from BE search, and extract the data to forward it to
	// an autocompelete input. Also update the pageMap.
	this.processAutocompleteResults = function(data) {
		if (!data) return [];
		// Add new pages to the pageMap.
		for (var pageId in data.pages) {
			pageService.addPageToMap(data.pages[pageId]);
		}
		// Create list of results we can give to autocomplete.
		var resultList = [];
		var hits = data.result.search.hits;
		for (var n = 0; n < hits.length; n++) {
			var source = hits[n]._source;
			resultList.push({
				value: source.alias,
				label: source.pageId,
				alias: source.alias,
				title: source.title,
				clickbait: source.clickbait,
				seeGroupId: source.seeGroupId,
				score: hits[n]._score,
			});
		}
		return resultList;
	};


	// Do a normal search with the given options.
	// options = {
	//	term: string to search for
	//	pageType: contraint for what type of pages we are looking for
	// }
	// Returns: list of results
	this.performSearch = function(options, callback) {
		$http({method: "POST", url: "/json/search/", data: JSON.stringify(options)})
			.success(function(data, status){
				var results = that.processAutocompleteResults(data);
				if (callback) callback(results);
			})
			.error(function(data, status){
				console.log("Error loading /search/ autocomplete data:"); console.log(data); console.log(status);
				if (callback) callback({});
			});
	}

	// Load data for autocompleting parents search.
	this.parentsSource = function(request, callback) {
		$http({method: "POST", url: "/json/parentsSearch/", data: JSON.stringify(request)})
			.success(function(data, status){
				var results = that.processAutocompleteResults(data);
				if (callback) callback(results);
			})
			.error(function(data, status){
				console.log("Error loading /parentsSearch/ autocomplete data:"); console.log(data); console.log(status);
				callback([]);
			});
	};

	// Set up autocompletion based on parents search for the given input field.
	this.setupParentsAutocomplete = function($input, selectCallback) {
	  $input.autocomplete({
			source: that.parentsSource,
			minLength: 3,
			delay: 300,
			focus: function (event, ui) {
				return false;
			},
			select: function (event, ui) {
				return selectCallback(event, ui);
			}
	  });
	}

	// Find other pages similar to the page with the given data.
	this.findSimilarPages = function(pageData, callback) {
		$http({method: "POST", url: "/json/similarPageSearch/", data: JSON.stringify(pageData)})
		.success(function(data, status){
			var results = that.processAutocompleteResults(data);
			if (callback) callback(results);
		})
		.error(function(data, status){
			console.log("Error doing similar page search:"); console.log(data); console.log(status);
		});
	};
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

// ArbitalCtrl is used across all pages.
app.controller("ArbitalCtrl", function ($scope, $location, $timeout, $http, $compile, userService, pageService, autocompleteService) {
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

		if (pageId) {
			// Load the last edit
			var specificEdit = $location.search().edit;
			pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: specificEdit,
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
	}

	// Domain index page
	var pagesPath = /^\/domains\/([A-Za-z0-9]+)\/?$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		pageService.domainAlias = match[1];
		var postData = {
			domainAlias: pageService.domainAlias,
		};
		// Get the domain index page data
		$http({method: "POST", url: "/json/domainIndex/", data: JSON.stringify(postData)})
		.success(getSuccessFunc(function(data){
			$scope.indexPageIdsMap = data["result"];
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
			var rootPage = pageService.pageMap[data["result"].rootPageId];
			$scope.rootPages = {};
			$scope.rootPages[rootPage.pageId] = rootPage;
			$scope.childPages = {};
			var length = rootPage.children ? rootPage.children.length : 0;
			for (var n = 0; n < length; n++) {
				var childId = rootPage.children[n].childId;
				$scope.childPages[childId] = pageService.pageMap[childId];
			}

			// Add the tree directive
			return {
				title: title,
				element: $("<arb-page-tree init-map='rootPages' additional-map='childPages'" +
					"supersize-roots='true'></arb-page-tree>"),
			};
		}))
		.error(getErrorFunc("explore"));
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
		getSuccessFunc(function(data){
			return {
				title: "Sign Up",
				element: $("<arb-signup></arb-signup>"),
			};
		})();
	}

	// Index page
	var pagesPath = /^\/$/;
	var match = pagesPath.exec($location.path());
	if (match) {
		if (subdomain) {
			// Get the private group index page data
			$http({method: "POST", url: "/json/privateIndex/"})
			.success(getSuccessFunc(function(data){
				$scope.indexPageIdsMap = data["result"];
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
				$scope.featuredDomains = data["result"].featuredDomains;
				return {
					element: $("<arb-index featured-domains='featuredDomains'></arb-index>"),
				};
			}))
			.error(getErrorFunc("index"));
		}
	}
});



// =============================== DIRECTIVES =================================

// navbar directive displays the navbar at the top of each page
app.directive("arbNavbar", function($http, $location, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/navbar.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.user = userService.user;

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

// userName directive displayes a user's name.
app.directive("arbUserName", function(userService) {
	return {
		templateUrl: "/static/html/userName.html",
		scope: {
			userId: "@",
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.user = userService.userMap[scope.userId];
		},
	};
});

// newLinkModal directive is used for storing a modal that creates new links.
app.directive("arbNewLinkModal", function(autocompleteService) {
	return {
		templateUrl: "/static/html/newLinkModal.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			var $input = element.find(".new-link-input");
			// Set up autocomplete
			autocompleteService.setupParentsAutocomplete($input, function(event, ui) {
				element.find(".modal-content").submit();
				return true;
			});
			// Set up search for new link modal
			autocompleteService.setAutocompleteRendering($input, scope);
		},
	};
});

// intrasitePopover containts the popover body html.
app.directive("arbIntrasitePopover", function(pageService, userService) {
	return {
		templateUrl: "/static/html/intrasitePopover.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			arbMarkdown.init(false, scope.pageId, scope.page.summary, element.find(".intrasite-popover-body"), pageService, userService);
		},
	};
});

// pageTitle displays page's title with optional meta info.
app.directive("arbPageTitle", function(pageService, userService) {
	return {
		templateUrl: "/static/html/pageTitle.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
		},
	};
});

// likesPageTitle displays likes span followed by page's title span.
app.directive("arbLikesPageTitle", function(pageService, userService) {
	return {
		templateUrl: "/static/html/likesPageTitle.html",
		scope: {
			pageId: "@",
			showClickbait: "@",
			showRedLinkCount: "@",
			showQuickEditLink: "@",
			showCreatedAt: "@",
			isSearchResult: "@",
			isSupersized: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
		},
	};
});

// pageTree displays pageTreeNodes in a recursive tree structure.
app.directive("arbPageTree", function() {
	return {
		templateUrl: "/static/html/pageTree.html",
		controller: function ($scope, pageService) {
			// Map of pageId -> array of nodes.
			var pageIdToNodesMap = {};
			// Return a new node object corresponding to the given pageId.
			// The pair will also be added to the pageIdToNodesMap.
			var createNode = function(pageId) {
				var node = {
					pageId: pageId,
					showChildren: false,
					children: [],
				};
				var nodes = pageIdToNodesMap[node.pageId];
				if (nodes === undefined) {
					nodes = [];
					pageIdToNodesMap[node.pageId] = nodes;
				}
				nodes.push(node);
				return node;
			};
		
			// Sort node's children based on how the corresponding page sorts its children.
			$scope.sortNodeChildren = function(node) {
				var sortChildrenBy = "alphabetical";
				if (node === $scope.rootNode) {
					if ($scope.primaryPageId) {
						sortChildrenBy = pageService.pageMap[$scope.primaryPageId].sortChildrenBy;
					}
				} else {
					sortChildrenBy = pageService.pageMap[node.pageId].sortChildrenBy;
				}
				var sortFunc = pageService.getChildSortFunc(sortChildrenBy);
				node.children.sort(function(aNode, bNode) {
					return sortFunc(aNode.pageId, bNode.pageId);
				});
			};
		
			// Return true iff the given node has a node child corresponding to the pageId.
			var nodeHasPageChild = function(node, pageId) {
				return node.children.some(function(child) {
					return child.pageId == pageId;
				});
			};
		
			// processPages adds a new node for every page in the given newPagesMap.
			$scope.processPages = function(newPagesMap, topLevel) {
				if (newPagesMap === undefined) return;
				// Process parents and create children nodes.
				for (var pageId in newPagesMap) {
					var page = pageService.pageMap[pageId];
					if (!page) {
						console.warn("Couldn't find child id " + pageId);
						continue;
					}
					var parents = page.parents; // array of pagePairs used to populate children nodes
					if ($scope.isParentTree !== undefined) {
						parents = page.children;
					}
					if (topLevel) {
						if (!nodeHasPageChild($scope.rootNode, pageId)) {
							var node = createNode(pageId);
							node.isTopLevel = true;
							$scope.rootNode.children.push(node);
						}
					} else {
						// For each parent the page has, find all corresponding nodes, and add
						// a new child node to each of them.
						var parentsLen = parents.length;
						for (var i = 0; i < parentsLen; i++){
							var parentId = parents[i].parentId;
							if ($scope.isParentTree !== undefined) {
								parentId = parents[i].childId;
							}
							var parentPage = pageService.pageMap[parentId];
							var parentNodes = parentPage ? (pageIdToNodesMap[parentPage.pageId] || []) : [];
							var parentNodesLen = parentNodes.length;
							for (var ii = 0; ii < parentNodesLen; ii++){
								var parentNode = parentNodes[ii];
								if (!nodeHasPageChild(parentNode, pageId)) {
									parentNode.children.push(createNode(pageId));
								}
							}
						}
					}
				}
			};
		
			// Imaginary root node we use to make the architecture simpler.
			$scope.rootNode = {pageId:"-1", children:[]};
		
			// Populate the tree.
			$scope.processPages($scope.initMap, true);
			$scope.processPages($scope.additionalMap);
		
			if (!$scope.isParentTree) {
				// Sort children.
				$scope.sortNodeChildren($scope.rootNode);
				for (var n = 0; n < $scope.rootNode.children.length; n++) {
					$scope.sortNodeChildren($scope.rootNode.children[n]);
				}
			}
		},
		scope: {
			supersizeRoots: "@", // if defined, the root nodes are displayed bigger
			isParentTree: "@", // if defined, the nodes' children actually represent page's parents, not children
			primaryPageId: "@", // if defined, we'll assume this page is the parent of the roots
			initMap: "=", // if defined, the pageId->page map will be used to seed the tree's roots
			additionalMap: "=", // if defined, the pageId->page map will be used to populate this tree
		},
	};
});

// pageTreeNode displays the corresponding page and it's node children
// recursively, allowing the user to recursively explore the page tree.
app.directive("arbPageTreeNode", function(RecursionHelper) {
	return {
		templateUrl: "/static/html/pageTreeNode.html",
		controller: function ($scope, pageService) {
			$scope.page = pageService.pageMap[$scope.node.pageId];
			$scope.node.showChildren = !!$scope.node.isTopLevel && $scope.additionalMap;
		
			// Toggle the node's children visibility.
			$scope.toggleNode = function(event) {
				// TODO: this recursive expansion is pretty fucked. Need to redo the whole
				// thing probably, without RecursionHelper.
				var recursiveExpand = event.shiftKey || event.shiftKey === undefined;
				$scope.node.showChildren = !$scope.node.showChildren;
				if ($scope.node.showChildren) {
					var loadFunc = pageService.loadChildren;
					if ($scope.isParentTree) {
						loadFunc = pageService.loadParents;
					}
					loadFunc.call(pageService, $scope.page,
						function(data, status) {
							$scope.processPages(data);
							if (recursiveExpand) {
								// Recursively expand children nodes
								window.setTimeout(function() {
									$(event.target).closest("arb-page-tree-node").find(".page-panel-body")
										.find(".collapse-link.glyphicon-triangle-right:visible").trigger("click");
								});
							}
						},
						function(data, status) { }
					);
				}
			};
			// Return true iff the corresponding page is loading children.
			$scope.isLoadingChildren = function() {
				return $scope.page.isLoadingChildren;
			};
		
			// Return true if we should show the collapse arrow button for this page.
			$scope.showCollapseArrow = function() {
				return (!$scope.isParentTree && $scope.page.hasChildren) || ($scope.isParentTree && $scope.page.hasParents);
			};
		
			// Return true iff this node should be displayed larger.
			$scope.isSupersized = function() {
				return $scope.node.isTopLevel && $scope.supersizeRoots;
			};
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		},
	};
});
</script>
{{end}}
