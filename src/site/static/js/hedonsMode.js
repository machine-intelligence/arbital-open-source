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
			
			arb.userService.user.newAchievementCount = 0;
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
