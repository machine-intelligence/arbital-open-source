/* angular.tmpl.js is a .tmpl file that is inserted as a <script> into the
	<header> portion of html pages that use angular. It defines the zanaduu module
	and ZanaduuCtrl, which are used on every page. */
/* Stephanie:
 * 2) Avoid putting things on scope. Use 'this' instead.
 * 4) For PagePanelCtrl, set directive scope variables (needs to go with #1)
 * 5) $resource is for CRUD (also should have more obvious way to showing what params it takes)
 * 6) Rename "page" to something else
 * 7) Rename "panel" to something else
 * 8) ChildrenCtrl & ParentsCtrl refactor (via a directive)
 * 9) PageLoader service which encapsulates functions like loadChildren(parentId) and adds them to pageMap
 */
{{define "angular"}}
<script>

var app = angular.module("zanaduu", ["ngResource", "ui.bootstrap", "RecursionHelper"]);
app.config(function($interpolateProvider){
	$interpolateProvider.startSymbol("{[{").endSymbol("}]}");
});

// User service.
app.service("userService", function(){
	// Logged in user.
	this.user = {{GetUserJson}};

	// Get maximum karma lock a user can set up.
	this.user.getMaxKarmaLock = function() {
		return Math.floor(this.Karma * {{GetMaxKarmaLockFraction}});
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

	var pageFuncs = {
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
			return this.Author.Id != userService.user.Id && this.LastVisit != "" && this.CreatedAt >= this.LastVisit;
		},
		// Return empty string if the user can edit this page. Otherwise a reason for
		// why they can't.
		getEditLevel: function() {
			if (this.Type == "blog") {
				if (this.Author.Id == userService.user.Id) {
					return "";
				} else {
					return "blog";
				}
			}
			var karmaReq = this.KarmaLock;
			var editPageKarmaReq = 10; // TODO: fix this
			if (karmaReq < editPageKarmaReq && this.WasPublished) {
				karmaReq = editPageKarmaReq
			}
			if (userService.user.Karma < karmaReq) {
				if (userService.user.IsAdmin) {
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
			if (this.Type == "blog") {
				if (this.Author.Id == userService.user.Id) {
					return "";
				} else if (userService.user.IsAdmin) {
					return "admin";
				} else {
					return "blog";
				}
			}
			var karmaReq = this.KarmaLock;
			var deletePageKarmaReq = 200; // TODO: fix this
			if (karmaReq < deletePageKarmaReq) {
				karmaReq = deletePageKarmaReq;
			}
			if (userService.user.Karma < karmaReq) {
				if (userService.user.IsAdmin) {
					return "admin";
				}
				return "" + karmaReq;
			}
			return "";
		},
	};
	
	// Massage page's variables to be easier to deal with.
	var setUpPage = function(page) {
		if (page.Children == null) page.Children = [];
		if (page.Parents == null) page.Parents = [];
		page.Url = "/pages/" + page.Alias;
		page.EditUrl = "/edit/" + page.PageId;
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		return page;
	};
	this.addPageToMap = function(page) {
		var existingPage = this.pageMap[page.PageId];
		if (existingPage !== undefined) {
			if (page === existingPage) return;
			console.log("existingPage");
			console.log(existingPage);
			// Merge.
			existingPage.Children = existingPage.Children.concat(page.Children);
			existingPage.Parents = existingPage.Parents.concat(page.Parents);
		} else {
			this.pageMap[page.PageId] = setUpPage(page);
		}
		return this.pageMap[page.PageId];
	};
	this.removePageFromMap = function(pageId) {
		delete this.pageMap[pageId];
		console.log(this.pageMap);
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
		console.log("/json/children/?parentId=" + parent.PageId);
		$http({method: "GET", url: "/json/children/", params: {parentId: parent.PageId}}).
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
				console.log("error loading children");
				console.log(data);
				console.log(status);
				error(data, status);
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
		console.log("/json/parents/?childId=" + child.PageId);
		$http({method: "GET", url: "/json/parents/", params: {childId: child.PageId}}).
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
				console.log("error loading parents");
				console.log(data);
				console.log(status);
				error(data, status);
			});
	};

	// Load the page with the given pageIds. If it's empty, ask the server for
	// a new page id.
	var loadingPageIds = {};
	this.loadPages = function(pageIds, success, error) {
		var service = this;
		var pageIdsLen = pageIds.length;
		var pageIdsStr = "";
		// Add pages to the global map as necessary. Set pages as loading.
		// Compute pageIdsStr for page ids that are not being loaded already.
		for (var n = 0; n < pageIdsLen; n++) {
			var pageId = pageIds[n];
			if (!(pageId in loadingPageIds)) {
				loadingPageIds[pageId] = true;
				pageIdsStr += pageId + ",";
			}
		}
		if (pageIdsLen > 0 && pageIdsStr.length == 0) {
			return;  // we are loading all the pages already
		}
		console.log("/json/pages/?pageIds=" + pageIdsStr);
		$http({method: "GET", url: "/json/pages/", params: {pageIds: pageIdsStr, loadFullEdit: true}}).
			success(function(data, status){
				for (var id in data) {
					console.log("data");
					console.log(data[id]);
					data[id] = service.addPageToMap(data[id]);
					delete loadingPageIds[id];
				}
				success(data, status);
			}).error(function(data, status){
				console.log("error loading page"); console.log(data); console.log(status);
				error(data, status);
			});
	};

	// Setup all initial pages.
	console.log(this.pageMap);
	for (var id in this.pageMap) {
		setUpPage(this.pageMap[id]);
	}
});

// simpleDateTime filter converts our typical date&time string into local time.
app.filter("simpleDateTime", function() {
	return function(input) {
		var date = new Date(input + " UTC");
		return date.toLocaleString().format("dd-m-yy");
	};
});

// ZanaduuCtrl is used across all pages.
app.controller("ZanaduuCtrl", function ($scope, userService, pageService) {
	$scope.userService = userService;
});


// PageTreeCtrl is controller for the PageTree.
app.controller("PageTreeCtrl", function ($scope, pageService) {
	// Map of pageId -> array of nodes.
	var pageIdToNodesMap = {};
	// Return a new node object corresponding to the given pageId.
	// The pair will also be added to the pageIdToNodesMap.
	var createNode = function(pageId) {
		var node= {
			pageId: pageId,
			showChildren: false, // TODO: topLevel?
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
			var parents = page.Parents; // array of pagePairs used to populate children nodes
			if ($scope.isParentTree !== undefined) {
				parents = page.Children;
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
					var parentId = parents[i].ParentId;
					if ($scope.isParentTree !== undefined) {
						parentId = parents[i].ChildId;
					}
					var parentPage = pageService.pageMap[parentId];
					var parentNodes = pageIdToNodesMap[parentPage.PageId] || [];
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

	// Sorting function for node's children.
	$scope.sortChildren = function (node) {
		var page = pageService.pageMap[node.pageId];
		return page.Title;
	};
});

// PageTreeNodeCtrl is created for each node under the PageTreeCtrl.
app.controller("PageTreeNodeCtrl", function ($scope, pageService) {
	$scope.page = pageService.pageMap[$scope.node.pageId];
	$scope.node.showChildren = $scope.node.children.length > 0;

	// toggleNode gets called when the user clicks to show/hide the node.
	$scope.toggleNode = function() {
		$scope.node.showChildren = !$scope.node.showChildren;
		if ($scope.node.showChildren) {
			var loadFunc = pageService.loadChildren;
			if ($scope.isParentTree) {
				loadFunc = pageService.loadParents;
			}
			loadFunc.call(pageService, $scope.page,
				function(data, status) {
					$scope.processPages(data);
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
		return (!$scope.isParentTree && $scope.page.HasChildren) || ($scope.isParentTree && $scope.page.HasParents);
	};

	// Return true iff this node should be displayed larger.
	$scope.isSupersized = function() {
		return $scope.node.isTopLevel && $scope.supersizeRoots;
	};
});

// =============================== DIRECTIVES =================================

// pageTitle displays page's title with optional meta info.
app.directive("zndPageTitle", function(pageService) {
	return {
		templateUrl: "/static/html/pageTitle.html",
		scope: {
			pageId: "=",
		},
		// TODO: don't do link
		link: function(scope, element, attrs) {
			scope.page = pageService.pageMap[scope.pageId];
		},
	};
});

// likesPageTitle displays likes span followed by page's title span.
app.directive("zndLikesPageTitle", function(pageService) {
	return {
		templateUrl: "/static/html/likesPageTitle.html",
		scope: {
			pageId: "=",
			showRedLinkCount: "=",
			showQuickEditLink: "=",
			showCreatedAt: "=",
		},
		// TODO: don't do link
		link: function(scope, element, attrs) {
			scope.page = pageService.pageMap[scope.pageId];
		},
	};
});

// pageTree displays pageTreeNodes in a recursive tree structure.
app.directive("zndPageTree", function() {
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
app.directive("zndPageTreeNode", function(RecursionHelper) {
	return {
		templateUrl: "/static/html/pageTreeNode.html",
		controller: "PageTreeNodeCtrl",
		compile: function(element) {
			return RecursionHelper.compile(element);
		},
	};
});


</script>
{{end}}
