'use strict';

import app from './angular.ts';

// arb-recent-changes displays a list of recent changes
app.directive('arbRecentChanges', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/listPanel.html'),
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			modeRows: '=',
			type: '@',
			moreLink: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;

			if (!$scope.modeRows) {
				$scope.fetchMore = function() {
					var postUrl = '/json/recentChanges/';
					if ($scope.type == 'relationships') {
						postUrl = '/json/recentRelationshipChanges/';
					}

					var createdBefore = $scope.modeRows ?
							$scope.modeRows[$scope.modeRows.length - 1].changeLog.createdAt : '';

					$scope.fetchingMore = true;
					arb.stateService.postData(postUrl, {
							numToLoad: $scope.numToDisplay,
							createdBefore: createdBefore,
						},
						function(data) {
							if ($scope.modeRows) {
								var allModeRows = $scope.modeRows.concat(data.result.modeRows);

								// Remove duplicates
								$scope.modeRows = allModeRows.filter(function(i, index) {
								    return index == allModeRows.findIndex(function(j) {
											if (i.changeLog && j.changeLog) {
												return i.changeLog.id == j.changeLog.id;
											}
										return false;
								    });
								});
							} else {
								$scope.modeRows = data.result.modeRows;
								$scope.lastView = data.result.lastView;
							}
							$scope.fetchingMore = false;
						});
				};
				$scope.fetchMore();
			}
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
			$scope.changesTab = 0;
			$scope.todoTab = 0;

			$scope.selectChangesTab = function(tab) {
				$scope.changesTab = tab;
			};

			$scope.selectTodoTab = function(tab) {
				$scope.todoTab = tab;
			};
		},
	};
});
