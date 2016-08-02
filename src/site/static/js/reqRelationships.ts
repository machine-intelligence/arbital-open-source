'use strict';

import app from './angular.ts';

// Directive for editing the requirements or subjects.
app.directive('arbReqRelationships', function($q, $timeout, $interval, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/reqRelationships.html'),
		scope: {
			pageId: '@',
			type: '@',
			readonly: '=',
			// If set, will take the page from pageMap not editMap
			useNormalPageMap: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			if ($scope.useNormalPageMap) {
				$scope.page = arb.stateService.pageMap[$scope.pageId];
			} else {
				$scope.page = arb.stateService.editMap[$scope.pageId];
			}

			// Helper variables
			$scope.isRequirementType = $scope.type === 'requirement';
			$scope.isSubjectType = $scope.type === 'subject';

			// Compute various variables based on the type
			if ($scope.isRequirementType) {
				$scope.source = $scope.page.requirements;
			} else if ($scope.isSubjectType) {
				$scope.source = $scope.page.subjects;
			}
			$scope.source.sort(function(a, b) {
				var varsA = [a.isStrong ? 0 : 1, -a.level, a.createdAt];
				var varsB = [b.isStrong ? 0 : 1, -b.level, b.createdAt];
				for (var n = 0; n < varsA.length; n++) {
					if (varsA[n] == varsB[n]) continue;
					return varsA[n] < varsB[n] ? -1 : 1;
				}
				return 0;
			});

			// Set up search
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				arb.autocompleteService.parentsSource({term: text}, function(results) {
					deferred.resolve(results);
				});
				return deferred.promise;
			};
			$scope.searchResultSelected = function(result, isStrong = false) {
				if (!result) return;
				var params = {
					parentId: result.pageId,
					childId: $scope.page.pageId,
					type: $scope.type,
					level: 2,
					isStrong: isStrong,
				};
				arb.pageService.newPagePair(params, function success(data) {
					$scope.source.push(data.result.pagePair);
				});
			};

			// Called to update the given relationship
			$scope.updateRelationship = function(pagePair) {
				var params = {
					id: pagePair.id,
					level: +pagePair.level,
					isStrong: pagePair.isStrong === true || pagePair.isStrong === 'true',
				};
				arb.pageService.updatePagePair(params);
			};

			$scope.relatesToItself = $scope.source.some(function(pagePair) {
				return pagePair.parentId == $scope.pageId;
			});
			$scope.teachItself = function() {
				// Make the page teach itself. Used when the page isn't published yet.
				$scope.searchResultSelected({
					pageId: $scope.pageId,
				}, true);
				$scope.relatesToItself = true;
			};

			// Process deleting a relationship
			$scope.deleteRelationship = function(otherPageId) {
				var params = {
					parentId: otherPageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.deletePagePair(params, function success() {
					var indexToRemove;
					for (var i = 0; i < $scope.source.length; i++) {
						if ($scope.source[i].parentId == params.parentId) {
							indexToRemove = i;
							break;
						}
					}

					$scope.source.splice(indexToRemove, 1);

					$scope.relatesToItself = $scope.source.some(function(pagePair) {
						return pagePair.parentId == $scope.pageId;
					});
				});
			};
		},
	};
});

