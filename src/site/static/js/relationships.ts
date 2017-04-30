'use strict';

import app from './angular.ts';

// Directive for editing the parents, tags, requirements, or subjects.
app.directive('arbRelationships', function($q, $timeout, $interval, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/relationships.html'),
		scope: {
			pageId: '@',
			type: '@',
			readonly: '=',
			// Optional. Id of the parent page for quick add
			quickParentId: '@',
			// If set, will take the page from pageMap not editMap
			useNormalPageMap: '=',
			// Function to call when relationships are updated
			onRelationshipChange: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;
			if ($scope.useNormalPageMap) {
				$scope.page = arb.stateService.pageMap[$scope.pageId];
			} else {
				$scope.page = arb.stateService.editMap[$scope.pageId];
			}

			// Helper variables
			$scope.isParentType = $scope.type === 'parent';
			$scope.isTagType = $scope.type === 'tag';

			// Compute various variables based on the type
			if ($scope.isParentType) {
				$scope.idsSource = $scope.page.parentIds;
			} else if ($scope.isTagType) {
				$scope.idsSource = $scope.page.tagIds;
			}

			// Set up search
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				arb.autocompleteService.parentsSource({term: text}, function(results) {
					deferred.resolve(results);
				});
				return deferred.promise;
			};
			$scope.searchResultSelected = function(result) {
				if (!result) return;
				var params = {
					parentId: result.pageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.newPagePair(params, function success() {
					$scope.idsSource.push(params.parentId);
					if ($scope.onRelationshipChange) {
						$scope.onRelationshipChange();
					}
				});
			};

			// Process deleting a relationship
			$scope.deleteRelationship = function(otherPageId) {
				var params = {
					parentId: otherPageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.deletePagePair(params, function success() {
					$scope.idsSource.splice($scope.idsSource.indexOf(params.parentId), 1);
					if ($scope.onRelationshipChange) {
						$scope.onRelationshipChange();
					}
				});
			};

			$scope.addQuickParent = function() {
				$scope.searchResultSelected({pageId: $scope.quickParentId});
				$scope.quickParentId = undefined;
			};
		},
	};
});

