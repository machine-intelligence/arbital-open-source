/* angular.tmpl.js is a .tmpl file that is inserted as a <script> into the
	<header> portion of html pages that use angular. It defines the zanaduu module
	and ZanaduuCtrl, which are used on every page. */
{{define "angular"}}
<script>

// Set up angular module.
var app = angular.module("zanaduu", ["ngResource", "ui.bootstrap", "RecursionHelper"]);
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
				},
			{{end}}
		{{end}}
	};
	console.log("Initial userMap:");
	console.log(this.userMap);

	// Get maximum karma lock a user can set up.
	this.user.getMaxKarmaLock = function() {
		return Math.floor(this.karma * {{GetMaxKarmaLockFraction}});
	};
	this.getUserUrl = function(userId) {
		return "/filter?user=" + userId;
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
	this.primaryPage = "{{.PrimaryPageId}}" === "0" ? undefined : this.pageMap["{{.PrimaryPageId}}"];

	var pageFuncs = {
		// Check if the user has never visited this page before.
		isNewPage: function() {
			return this.creatorId != userService.user.id &&
				(this.lastVisit === "" || this.originalCreatedAt >= this.lastVisit);
		},
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
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
			var karmaReq = this.karmaLock;
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
			var karmaReq = this.karmaLock;
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
	};
	
	// Massage page's variables to be easier to deal with.
	var setUpPage = function(page) {
		if (page.children == null) page.children = [];
		if (page.parents == null) page.parents = [];
		page.url = "/pages/" + page.alias;
		page.editUrl = "/edit/" + page.pageId;
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		return page;
	};
	this.addPageToMap = function(page) {
		var existingPage = this.pageMap[page.pageId];
		if (existingPage !== undefined) {
			if (page === existingPage) return;
			console.log("existingPage:"); console.log(existingPage);
			// Merge.
			existingPage.children = existingPage.children.concat(page.children);
			existingPage.parents = existingPage.parents.concat(page.parents);
		} else {
			this.pageMap[page.pageId] = setUpPage(page);
		}
		return this.pageMap[page.pageId];
	};
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
				console.log("Unknown sort type: " + page.sortChildrenBy);
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
		console.log("Issuing a GET request to: /json/pages/?pageIds=" + pageIdsStr);
		$http({method: "GET", url: "/json/pages/", params: {pageIds: pageIdsStr, loadFullEdit: true}}).
			success(function(data, status){
				for (var id in data) {
					console.log("JSON /pages/ data:"); console.log(data[id]);
					data[id] = service.addPageToMap(data[id]);
					delete loadingPageIds[id];
				}
				if(success) success(data, status);
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
		$http({method: "POST", url: "/deletePage/", data: JSON.stringify(data)}).
			success(function(data, status){
				console.log("Successfully deleted " + pageId);
				if(success) success(data, status);
			}).error(function(data, status){
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
			}).error(function(data, status){
				console.log("Error abandoning " + pageId + ":"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

	// Return true iff we should show that this page is public.
	this.showPublic = function(pageId) {
		var page = this.pageMap[pageId];
		if (page.group.name) return false;
		if (!this.primaryPage) return false;
		return this.primaryPage.group.name !== page.group.name;
	};
	// Return true iff we should show that this page belongs to a group.
	this.showLockedGroup = function(pageId) {
		var page = this.pageMap[pageId];
		if (!page.group.name) return false;
		if (!this.primaryPage) return true;
		return this.primaryPage.group.name !== page.group.name;
	};

	// Setup all initial pages.
	console.log("Initial pageMap: "); console.log(this.pageMap);
	for (var id in this.pageMap) {
		setUpPage(this.pageMap[id]);
	}
});

// Autocomplete service provides data for autocompletion.
app.service("autocompleteService", function($http){
	// Map of all aliases: fullName -> {pageId, title}
	this.aliasMap = {};
	// This array stores all available aliases for link autocompletion.
	this.aliasSource = [];

	// Load the data for aliasSource.
	var aliasCallbacks = [];
	this.loadAliasSource = function(callback) {
		// Check if already loaded.
		if (this.aliasSource.length > 0) {
			if (callback) callback();
			return;
		}
		// Add this callback to our list.
		if (callback) {
			aliasCallbacks.push(callback);
		}
		// Load aliases.
		var that = this;
		console.log("Issuing a GET request to: /json/aliases/");
		$http({method: "GET", url: "/json/aliases/", params: {}})
		.success(function(data, status){
			that.aliasMap = data;
			// Convert data into the aliasSource.
			for (var fullName in that.aliasMap) {
				var val = that.aliasMap[fullName];
				that.aliasSource.push('"' + val.pageTitle + '" (' + fullName + ')');
			}
			// Execute all callbacks.
			for (var i = 0; i < aliasCallbacks.length; i++){
				aliasCallbacks[i]();
			}
			aliasCallbacks = [];
		})
		.error(function(data, status){
			console.log("Error loading alias autocomplete data:"); console.log(data); console.log(status);
			if(error) error(data, status);
		});
	}

	// Converts "title (alias)" string into "alias". Used to process the string
	// seleted by alias autocompletion.
	this.convertInputToAlias = function(input) {
		var openParenIndex = input.lastIndexOf("(");
		if (openParenIndex > 0) {
		  // Input is probably of the type: "title" (alias)
		  var closeParenIndex = input.lastIndexOf(")");
		  input = input.substr(openParenIndex + 1, closeParenIndex - openParenIndex - 1);
		}
		return input;
	};
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

// userName directive displayes a user's name.
app.directive("zndUserName", function(userService) {
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

// pageTitle displays page's title with optional meta info.
app.directive("zndPageTitle", function(pageService) {
	return {
		templateUrl: "/static/html/pageTitle.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.page = pageService.pageMap[scope.pageId];
		},
	};
});

// likesPageTitle displays likes span followed by page's title span.
app.directive("zndLikesPageTitle", function(pageService) {
	return {
		templateUrl: "/static/html/likesPageTitle.html",
		scope: {
			pageId: "@",
			showRedLinkCount: "@",
			showQuickEditLink: "@",
			showCreatedAt: "@",
			isSupersized: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
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
							console.log(recursiveExpand);
							if (recursiveExpand) {
								// Recursively expand children nodes
								window.setTimeout(function() {
									$(event.target).closest("znd-page-tree-node").find(".page-panel-body")
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
