'use strict';

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
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.editMap[$scope.pageId];

			// Helper variables
			$scope.isParentType = $scope.type === 'parent';
			$scope.isTagType = $scope.type === 'tag';
			$scope.isRequirementType = $scope.type === 'requirement';
			$scope.isSubjectType = $scope.type === 'subject';
			$scope.isChildType = $scope.type === 'child';

			// Compute various variables based on the type
			if ($scope.isParentType) {
				$scope.idsSource = $scope.page.parentIds;
			} else if ($scope.isTagType) {
				$scope.idsSource = $scope.page.taggedAsIds;
			} else if ($scope.isRequirementType) {
				$scope.idsSource = $scope.page.requirementIds;
			} else if ($scope.isSubjectType) {
				$scope.idsSource = $scope.page.subjectIds;
			} else if ($scope.isChildType) {
				// Compute children which are not lenses
				$scope.page.childIds = $scope.page.childIds.filter(function(childId) {
					return !$scope.page.lenses.some(function(lens) {
						return lens.lensId == childId;
					});
				});
				$scope.idsSource = $scope.page.childIds;
			}

			// Lens sort listener (using ng-sortable library)
			$scope.lensSortListeners = {
				orderChanged: function(event) {
					if ($scope.page.lenses.length <= 0) return;
					var params = {
						pageId: $scope.page.pageId,
						lensOrder: {},
					};
					for (var n = 0; n < $scope.page.lenses.length; n++) {
						var lens = $scope.page.lenses[n];
						lens.lensIndex = n;
						params.lensOrder[lens.id] = n;
					}
					arb.stateService.postData('/json/updateLensOrder/', params);
				},
			};

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
				if ($scope.type === 'child') {
					var params = {
						parentId: $scope.page.pageId,
						childId: result.pageId,
						type: 'parent',
					};
					arb.pageService.newPagePair(params, function success() {
						$scope.idsSource.push(params.childId);
					});
					return;
				}
				var params = {
					parentId: result.pageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.newPagePair(params, function success() {
					$scope.idsSource.push(params.parentId);
				});
			};

			$scope.relatesToItself = $scope.idsSource.indexOf($scope.pageId) >= 0;
			$scope.teachItself = function() {
				// Make the page teach itself. Used when the page isn't published yet.
				$scope.searchResultSelected({
					pageId: $scope.pageId,
				});
				$scope.relatesToItself = true;
			};

			// Process deleting a relationship
			$scope.deleteRelationship = function(otherPageId) {
				if ($scope.type === 'child') {
					var params = {
						parentId: $scope.page.pageId,
						childId: otherPageId,
						type: 'parent',
					};
					arb.pageService.deletePagePair(params, function success() {
						$scope.idsSource.splice($scope.idsSource.indexOf(params.childId), 1);
					});
					return;
				}
				var params = {
					parentId: otherPageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.deletePagePair(params, function success() {
					$scope.idsSource.splice($scope.idsSource.indexOf(params.parentId), 1);
					$scope.relatesToItself = $scope.idsSource.indexOf($scope.pageId) >= 0;
				});
			};

			$scope.addQuickParent = function() {
				$scope.searchResultSelected({pageId: $scope.quickParentId});
				$scope.quickParentId = undefined;
			};

			// Called to submit a new name for a lens
			$scope.changeLensName = function(lens) {
				var params = {
					id: lens.id,
					name: lens.lensName,
				};
				arb.stateService.postData('/json/updateLensName/', params, function() {
					arb.popupService.showToast({text: 'Lens name updated'});
				});
			};

			// Called to create a new lens
			$scope.newLens = function(lensId) {
				var params = {
					pageId: $scope.pageId,
					lensId: lensId,
				};
				arb.stateService.postData('/json/newLens/', params, function(data) {
					var lens = data.result.lens;
					$scope.page.lenses.push(lens);
					$scope.page.childIds.splice($scope.page.childIds.indexOf(lensId), 1);
				});
			};

			// Called to remove a lens (but still keep it as a child page)
			$scope.removeLens = function(lens) {
				var params = {
					id: lens.id,
				};
				arb.stateService.postData('/json/deleteLens/', params, function() {
					$scope.page.lenses.splice($scope.page.lenses.indexOf(lens), 1);
					$scope.page.childIds.push(lens.lensId);
					$scope.lensSortListeners.orderChanged();
				});
			};
		},
	};
});

