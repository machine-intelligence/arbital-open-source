'use strict';

// arb-hedons-mode-panel directive displays a list of new hedonic updates
app.directive('arbHedonsModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			hideTitle: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.title = 'Achievements';
			$scope.moreLink = '/achievements';

			arb.stateService.postData('/json/hedons/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
				});
		},
	};
});

// arb-likes-mode-row is the directive for showing who liked current user's stuff
app.directive('arbLikesModeRow', function(arb) {
	return {
		templateUrl: 'static/html/likesModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.userNames = formatUsersForDisplay($scope.modeRow.userIds.map(function(userId) {
				return arb.userService.getFullName(userId);
			}));
		},
	};
});

// arb-reqs-taught-mode-row is the directive for showing who learned current user's reqs
app.directive('arbReqsTaughtModeRow', function(arb) {
	return {
		templateUrl: 'static/html/reqsTaughtModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.userNames = formatUsersForDisplay($scope.modeRow.userIds.map(function(userId) {
				return arb.userService.getFullName(userId);
			}));
			$scope.reqNames = formatReqsForDisplay($scope.modeRow.requisiteIds.map(function(pageMap) {
				return arb.stateService.pageMap[pageMap].title;
			}));
		},
	};
});

// arb-added-to-group-mode-row is the directive for showing that the user was added to a group
app.directive('arbAddedToGroupModeRow', function(arb) {
	return {
		templateUrl: 'static/html/addedToGroupModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
			if ($scope.update.goToPageId) {
				$scope.goToPage = arb.stateService.pageMap[$scope.update.goToPageId];
			}
		},
	};
});

// arb-removed-from-group-mode-row is the directive for showing that the user was removed from a group
app.directive('arbRemovedFromGroupModeRow', function(arb) {
	return {
		templateUrl: 'static/html/removedFromGroupModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
			if ($scope.update.goToPageId) {
				$scope.goToPage = arb.stateService.pageMap[$scope.update.goToPageId];
			}
		},
	};
});

// arb-hedons-mode-page is for displaying the entire /achievements page
app.directive('arbHedonsModePage', function($http, arb) {
	return {
		templateUrl: 'static/html/hedonsModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
