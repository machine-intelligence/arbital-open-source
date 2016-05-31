'use strict';

// exploreTreeNode displays the corresponding page and its node children
// recursively, allowing the user to recursively explore the page tree.
app.directive('arbExploreTreeNode', function(RecursionHelper, arb) {
	return {
		templateUrl: '/static/html/exploreTreeNode.html',
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.pageIds = $scope.page.childIds.concat($scope.page.lensIds);
			$scope.showChildren = true;

			// Sort children.
			$scope.pageIds.sort(function(aId, bId) {
				var pageA = arb.stateService.pageMap[aId];
				var pageB = arb.stateService.pageMap[bId];
				var varsA = [pageA.isLens() ? 0 : 1, pageA.hasChildren ? 0 : 1, pageA.title];
				var varsB = [pageB.isLens() ? 0 : 1, pageB.hasChildren ? 0 : 1, pageB.title];
				for (var n = 0; n < varsA.length; n++) {
					if (varsA[n] == varsB[n]) continue;
					return varsA[n] < varsB[n] ? -1 : 1;
				}
				return 0;
			});

			// Toggle the node's children visibility.
			$scope.toggleNode = function(event) {
				$scope.showChildren = !$scope.showChildren;
			};
			// Return true iff the corresponding page is loading children.
			$scope.isLoadingChildren = function() {
				return $scope.page.isLoadingChildren;
			};

			// Return true if we should show the collapse arrow button for this page.
			$scope.showCollapseArrow = function() {
				return $scope.page.hasChildren;
			};

			$scope.getExploreIcon = function() {
				if ($scope.page.isLens()) return 'lens';
				if ($scope.page.isQuestion()) return 'help_outline';
				return '';
			};

			$scope.getExploreSvgIcon = function() {
				return 'file_outline';
			};
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		},
	};
});
