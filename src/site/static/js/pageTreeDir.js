"use strict";

// pageTree displays pageTreeNodes in a recursive tree structure.
app.directive("arbPageTree", function() {
	return {
		templateUrl: "/static/html/pageTree.html",
		scope: {
			supersizeRoots: "@", // if defined, the root nodes are displayed bigger
			isParentTree: "@", // if defined, the nodes' children actually represent page's parents, not children
			primaryPageId: "@", // if defined, we'll assume this page is the parent of the roots
			initMap: "=", // if defined, the pageId->page map will be used to seed the tree's roots
			additionalMap: "=", // if defined, the pageId->page map will be used to populate this tree
		},
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
	};
});

// pageTreeNode displays the corresponding page and it's node children
// recursively, allowing the user to recursively explore the page tree.
app.directive("arbPageTreeNode", function() {
	return {
		templateUrl: "/static/html/pageTreeNode.html",
		controller: function ($scope, pageService) {
			$scope.page = pageService.pageMap[$scope.node.pageId];
			$scope.node.showChildren = !!$scope.node.isTopLevel && $scope.additionalMap;
		
			// Toggle the node's children visibility.
			$scope.toggleNode = function(event) {
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
	};
});
