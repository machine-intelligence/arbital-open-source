/* angular.tmpl.js is a .tmpl file that is inserted as a <script> into the
	<header> portion of html pages that use angular. It defines the arbital module
	and ArbitalCtrl, which are used on every page. */
{{define "angular"}}
<script>

// Set up angular module.
var app = angular.module("arbital", ["ngResource", "ui.bootstrap", "RecursionHelper"]);
app.config(function($interpolateProvider, $locationProvider){
	$interpolateProvider.startSymbol("{[{").endSymbol("}]}");

	$locationProvider.html5Mode({
		enabled: true,
		requireBase: false,
		rewriteLinks: false
	});
});

// User service.
app.service("userService", function(){
	// Logged in user.
	this.user = {{GetUserJson}};
	this.userMap = {
		{{if .UserMap}}
			{{range .UserMap}}
				"{{.Id}}": {
					id: "{{.Id}}",
					firstName: "{{.FirstName}}",
					lastName: "{{.LastName}}",
					isSubscribed: {{.IsSubscribed}},
				},
			{{end}}
		{{end}}
	};
	console.log("Initial user map:"); console.log(this.userMap);

	this.getUserUrl = function(userId) {
		return "/user/" + userId;
	};

	// Loaded groups.
	this.groupMap = {
		{{if .GroupMap}}
			{{range .GroupMap}}
				"{{.Id}}": {
					id: "{{.Id}}",
					name: "{{.Name}}",
					alias: "{{.Alias}}",
					isVisible: "{{.IsVisible}}",
					rootPageId: "{{.RootPageId}}",
					createdAt: "{{.CreatedAt}}",
				},
			{{end}}
		{{end}}
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
});

// pages stores all the loaded pages and provides multiple helper functions for
// working with pages.
app.service("pageService", function(userService, $http){
	// All loaded pages.
	this.pageMap = {
		{{range $k,$v := .PageMap}}
			"{{$k}}": {{GetPageJson $v}},
		{{end}}
	};

	// Primary page is the one that's displayed front and center.
	this.primaryPage = "{{.PrimaryPageId}}" === "0" ? undefined : this.pageMap["{{.PrimaryPageId}}"];
	// List of callbacks to notify when primary page changes.
	this.primaryPageCallbacks = [];
	// Set the primary page, triggering the callbacks.
	this.setPrimaryPage = function(newPrimaryPage) {
		var oldPrimaryPage = this.primaryPage;
		this.primaryPage = newPrimaryPage;
		for (var n = 0; n < this.primaryPageCallbacks.length; n++) {
			this.primaryPageCallbacks[n](oldPrimaryPage);
		}
	};

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
			if (this.type === "blog" || this.type === "comment") {
				if (this.creatorId == userService.user.id) {
					return "";
				} else {
					return this.type;
				}
			}
			var karmaReq = this.editKarmaLock;
			var editPageKarmaReq = 10; // TODO: fix this
			if (karmaReq < editPageKarmaReq && this.wasPublished) {
				karmaReq = editPageKarmaReq
			}
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
			if (this.type === "blog" || this.type === "comment") {
				if (this.creatorId == userService.user.id) {
					return "";
				} else if (userService.user.isAdmin) {
					return "admin";
				} else {
					return this.type;
				}
			}
			var karmaReq = this.editKarmaLock;
			var deletePageKarmaReq = 200; // TODO: fix this
			if (karmaReq < deletePageKarmaReq) {
				karmaReq = deletePageKarmaReq;
			}
			if (userService.user.karma < karmaReq) {
				if (userService.user.isAdmin) {
					return "admin";
				}
				return "" + karmaReq;
			}
			return "";
		},
		isDeleted: function() {
			return this.type === "deleted";
		},
		// Get page's url
		url: function(forcePageId) {
			if (forcePageId) {
				return "/pages/" + this.pageId;
			}
			return "/pages/" + this.alias;
		},
		// Get url to edit the page
		editUrl: function() {
			return "/edit/" + this.pageId;
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
		if (page.pageId !== page.alias) {
			pageMap[page.alias] = page;
		}
		return page;
	};
	// Add the given page to the global pageMap.
	// overwrite - if true, the given value will overwrite the page data we might already have.
	this.addPageToMap = function(page, overwrite) {
		var existingPage = this.pageMap[page.pageId];
		if (existingPage !== undefined && !overwrite) {
			if (page === existingPage) return false;
			// Merge.
			existingPage.children = existingPage.children.concat(page.children);
			existingPage.parents = existingPage.parents.concat(page.parents);
		} else {
			this.pageMap[page.pageId] = setUpPage(page, this.pageMap);
		}
		return true;
	};
	// Remove page with the given pageId from the global pageMap.
	this.removePageFromMap = function(pageId) {
		delete this.pageMap[pageId];
	};

	// Load children for the given page. Success/error callbacks are called only
	// if request was actually made to the server.
	this.loadChildren = function(parent, success, error) {
		var service = this;
		if (parent.hasLoadedChildren) {
			success(parent.loadChildrenData, 200);
			return;
		} else if (parent.isLoadingChildren) {
			return;
		}
		parent.isLoadingChildren = true;
		console.log("Issuing GET request to /json/children/?parentId=" + parent.pageId);
		$http({method: "GET", url: "/json/children/", params: {parentId: parent.pageId}}).
			success(function(data, status){
				parent.isLoadingChildren = false;
				parent.hasLoadedChildren = true;
				for (id in data) {
					data[id] = service.addPageToMap(data[id]);
				}
				parent.loadChildrenData = data;
				success(data, status);
			}).error(function(data, status){
				parent.isLoadingChildren = false;
				console.log("Error loading children:"); console.log(data); console.log(status);
				error(data, status);
			});
	};

	// Return function for sorting children ids.
	this.getChildSortFunc = function(page) {
		var pageMap = this.pageMap;
		if(page.sortChildrenBy === "alphabetical") {
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
		} else if (page.sortChildrenBy === "chronological") {
			var reverse = page.type === "comment";
			return function(aId, bId) {
				var r = pageMap[bId].originalCreatedAt.localeCompare(pageMap[aId].originalCreatedAt);
				return reverse ? -1*r : r;
			};
		} else {
			if (page.sortChildrenBy !== "likes") {
				console.error("Unknown sort type: " + page.sortChildrenBy);
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
		var sortFunc = this.getChildSortFunc(page);
		page.children.sort(function(aChild, bChild) {
			return sortFunc(aChild.childId, bChild.childId);
		});
	};

	// Load parents for the given page. Success/error callbacks are called only
	// if request was actually made to the server.
	this.loadParents = function(child, success, error) {
		var service = this;
		if (child.hasLoadedParents) {
			success(child.loadParentsData, 200);
			return;
		} else if (child.isLoadingParents) {
			return;
		}
		child.isLoadingParents = true;
		console.log("Issuing GET request to /json/parents/?childId=" + child.pageId);
		$http({method: "GET", url: "/json/parents/", params: {childId: child.pageId}}).
			success(function(data, status){
				child.isLoadingParents = false;
				child.hasLoadedParents = true;
				for (id in data) {
					data[id] = service.addPageToMap(data[id]);
				}
				child.loadParentsData = data;
				success(data, status);
			}).error(function(data, status){
				child.isLoadingParents = false;
				console.log("Error loading parents:"); console.log(data); console.log(status);
				error(data, status);
			});
	};

	// Load the page with the given pageAliases. If it's empty, ask the server for
	// a new page id.
	// options {
	//   includeText: include the full text of the page
	//   includeAuxData: include likes, subscription, etc...
	//   loadComments: whether or not to load comments
	//   loadVotes: whether or not to load votes
	//   allowDraft: allow the server to load an autosave / snapshot, if it's most recent
	//   overwrite: overwrite the existing pages with loaded data
	//   success: callback on success
	//   error: callback on error
	// }
	// Track which pages we are already loading. Map pageAlias -> true.
	var loadingPageAliases = {};
	var count = 0;
	this.loadPages = function(pageAliases, options) {
		var service = this;
		options.pageAliases = [];
		// Add pages to the global map as necessary. Set pages as loading.
		// Compute pageAliasesStr for page ids that are not being loaded already.
		for (var n = 0; n < pageAliases.length; n++) {
			var pageAlias = pageAliases[n];
			if (!(pageAlias in loadingPageAliases)) {
				loadingPageAliases[pageAlias] = true;
				options.pageAliases.push(pageAlias);
			}
		}
		if (pageAliases.length > 0 && options.pageAliases.length == 0) {
			return;  // we are loading all the pages already
		}

		// Set up options.
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;
		var overwrite = options.overwrite; delete options.overwrite;

		console.log("Issuing a GET request to: /json/pages/?pageAliases=" + pageAliases);
		$http({method: "GET", url: "/json/pages/", params: options}).
			success(function(data, status){
				console.log("JSON /pages/ data:"); console.log(data);
				var pagesData = data["pages"];
				for (var id in pagesData) {
					service.addPageToMap(pagesData[id], overwrite);
					delete loadingPageAliases[id];
					delete loadingPageAliases[pagesData[id].alias];
				}
				var usersData = data["users"];
				for (var id in usersData) {
					userService.userMap[id] = usersData[id];
				}
				if(success) success(pagesData, status);
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};
	
	// Load edit.
	// options {
	//   pageId: pageId to load
	//	 editLimit: only load edits lower than this number
	//	 createdAtLimit: only load edits that were created before this date
	//   overwrite: overwrite the existing pages with loaded data
	//   success: callback on success
	//   error: callback on error
	// }
	this.loadEdit = function(options) {
		var service = this;

		// Set up options.
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;
		var overwrite = options.overwrite; delete options.overwrite;

		console.log("Issuing a GET request to: /json/edit/?pageId=" + options.pageId);
		$http({method: "GET", url: "/json/edit/", params: options}).
			success(function(data, status){
				console.log("JSON /edit/ data:"); console.log(data);
				var pagesData = data["pages"];
				for (var id in pagesData) {
					data[id] = pagesData[id];
				}
				var usersData = data["users"];
				for (var id in usersData) {
					userService.userMap[id] = usersData[id];
				}
				if(success) success(pagesData, status);
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

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

	// Return true iff we should show that this page is public.
	this.showPublic = function(pageId) {
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
	console.log("Initial pageMap: "); console.log(this.pageMap);
	for (var id in this.pageMap) {
		setUpPage(this.pageMap[id], this.pageMap);
	}
});

// Autocomplete service provides data for autocompletion.
app.service("autocompleteService", function($http, $compile, pageService){
	var that = this;
	// Set how to render search results for the given autocomplete input.
	this.setAutocompleteRendering = function($input, scope) {
		$input.data("ui-autocomplete")._renderItem = function(ul, item) {
			var $el = $compile("<li class='search-result ui-menu-item' arb-likes-page-title page-id='" + item.value +
				"' show-clickbait='true' is-search-result='true'></li>")(scope);
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
			pageService.addPageToMap(data.pages[pageId], false);
		}
		// Create list of results we can give to autocomplete.
		var resultList = [];
		for (var n = 0; n < data.searchHits.hits.length; n++) {
			var source = data.searchHits.hits[n]._source;
			resultList.push({
				value: source.pageId,
				label: source.pageId,
				alias: source.alias,
				title: source.title,
				clickbait: source.clickbait,
				seeGroupId: source.seeGroupId,
			});
		}
		return resultList;
	};

	// Load data for autocompleting parents search.
	var parentsSource = function(request, callback) {
		$http({method: "GET", url: "/json/parentsSearch/", params: {term: request.term}})
		.success(function(data, status){
			callback(that.processAutocompleteResults(data));
		})
		.error(function(data, status){
			console.log("Error loading parentsSource autocomplete data:"); console.log(data); console.log(status);
			callback([]);
		});
	};

	// Set up autocompletion based on parents search for the given input field.
	this.setupParentsAutocomplete = function($input, selectCallback) {
	  $input.autocomplete({
			source: parentsSource,
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
			callback(that.processAutocompleteResults(data));
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

	// Refresh all the dates.
	var refreshDates = function() {
		$timeout(refreshDates, 30000);
	};
	refreshDates();

	// Process last visit url parameter
	var lastVisit = $location.search().lastVisit;
	if (lastVisit) {
		$("body").attr("last-visit", lastVisit);
		$location.search("lastVisit", null);
	}

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
				pageService.loadPages([pageAlias], {
					overwrite: true,
					includeAuxData: true,
					loadVotes: true,
					success: function(data, status) {
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
});

// PageTreeCtrl is controller for the PageTree.
app.controller("PageTreeCtrl", function ($scope, pageService) {
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
		if (node === $scope.rootNode) {
			var sortFunc = pageService.getChildSortFunc({sortChildrenBy: "alphabetical"});
		} else {
			var sortFunc = pageService.getChildSortFunc(pageService.pageMap[node.pageId]);
		}
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
});


// =============================== DIRECTIVES =================================

// navbar directive displays the navbar at the top of each page
app.directive("arbNavbar", function(pageService, userService, autocompleteService, $http) {
	return {
		templateUrl: "/static/html/navbar.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.user = userService.user;

			$("#logout").click(function() {
				$.removeCookie("zanaduu", {path: "/"});
			});

			// Function for getting search results from the server.
			var searchSource = function(request, callback) {
				$http({method: "GET", url: "/json/search/", params: {term: request.term}})
				.success(function(data, status){
					var resultMap = autocompleteService.processAutocompleteResults(data);
					callback(resultMap);
				})
				.error(function(data, status){
					console.log("Error loading parentsSource autocomplete data:"); console.log(data); console.log(status);
				});
			};

			// Setup search via navbar.
			var $navSearch = element.find("#nav-search");
			if ($navSearch.length > 0) {
				$navSearch.autocomplete({
					source: searchSource,
					minLength: 3,
					delay: 400,
					focus: function (event, ui) {
						return false;
					},
					select: function (event, ui) {
						window.location.href = "/pages/" + ui.item.value;
						return false;
					},
				});
				autocompleteService.setAutocompleteRendering($navSearch, scope);
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
		controller: "PageTreeCtrl",
		scope: {
			// Display options
			supersizeRoots: "@", // if defined, the root nodes are displayed bigger
			isParentTree: "@", // if defined, the nodes' children actually represent page's parents, not children
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
