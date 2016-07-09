'use strict';

// arb-recent-changes displays a list of recent changes
app.directive('arbRecentChanges', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/listPanel.html'),
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			type: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.fetchMore = function() {
				var postUrl = '/json/recentChanges/';
				if ($scope.type == 'relationships') {
					postUrl = '/json/recentRelationshipChanges/';
				}

				var createdBefore = $scope.modeRows ?
						$scope.modeRows[$scope.modeRows.length - 1].changeLog.createdAt : 'a';

				$scope.fetchingMore = true;
				arb.stateService.postData(postUrl, {
						numToLoad: $scope.numToDisplay,
						createdBefore: createdBefore,
					},
					function(data) {
						if ($scope.modeRows) {
							$scope.modeRows = $scope.modeRows.concat(data.result.modeRows);
						} else {
							$scope.modeRows = data.result.modeRows;
							$scope.lastView = data.result.lastView;
						}
						$scope.fetchingMore = false;
					});
			};
			$scope.fetchMore();
		},
	};
});


// arb-recent-changes displays a list of recent changes
app.directive('arbRecentChangesPage', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/recentChangesPage.html'),
		scope: {},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
