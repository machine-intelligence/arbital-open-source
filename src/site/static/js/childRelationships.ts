'use strict';

import app from './angular.ts';

// Directive for editing the children and lenses
app.directive('arbChildRelationships', function($q, $timeout, $interval, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/childRelationships.html'),
		scope: {
			pageId: '@',
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

			// Compute children which are not lenses
			// TODO: this is bad code because we are modifying the page object
			$scope.page.childIds = $scope.page.childIds.filter(function(childId) {
				return !$scope.page.lenses.some(function(lens) {
					return lens.lensId == childId;
				});
			});
			$scope.idsSource = $scope.page.childIds;

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
				var params = {
					parentId: $scope.page.pageId,
					childId: result.pageId,
					type: 'parent',
				};
				arb.pageService.newPagePair(params, function success() {
					$scope.idsSource.push(params.childId);
				});
			};

			// Process deleting a relationship
			$scope.deleteRelationship = function(otherPageId) {
				var params = {
					parentId: $scope.page.pageId,
					childId: otherPageId,
					type: 'parent',
				};
				arb.pageService.deletePagePair(params, function success() {
					$scope.idsSource.splice($scope.idsSource.indexOf(params.childId), 1);
				});
			};

			// Called to submit a new name for a lens
			$scope.changeLensName = function(lens) {
				var params = {
					id: lens.id,
					name: lens.lensName,
					subtitle: lens.lensSubtitle,
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

