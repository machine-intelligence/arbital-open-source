'use strict';

import app from './angular.ts';

// Directive for editing a path
app.directive('arbPathEditor', function($q, $timeout, $interval, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/pathEditor.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.editMap[$scope.pageId];

			// Sort listener (using ng-sortable library)
			$scope.sortListeners = {
				orderChanged: function(event) {
					if ($scope.page.pathPages.length <= 0) return;
					var params = {
						guideId: $scope.page.pageId,
						pageOrder: {},
					};
					for (var n = 0; n < $scope.page.pathPages.length; n++) {
						var pathPage = $scope.page.pathPages[n];
						pathPage.pathIndex = n;
						params.pageOrder[pathPage.id] = n;
					}
					arb.stateService.postData('/json/updatePathOrder/', params);
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
					guideId: $scope.page.pageId,
					pathPageId: result.pageId,
				};
				arb.stateService.postData('/json/newPathPage/', params, function(data) {
					$scope.page.pathPages.push(data.result.pathPage);
				});
			};

			// Called to remove a path page
			$scope.removePathPage = function(pathPage) {
				var params = {
					id: pathPage.id,
				};
				arb.stateService.postData('/json/deletePathPage/', params, function() {
					$scope.page.pathPages.splice($scope.page.pathPages.indexOf(pathPage), 1);
					$scope.sortListeners.orderChanged();
				});
			};
		},
	};
});

