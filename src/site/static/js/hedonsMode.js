'use strict';

// arb-hedons-mode-panel directive displays a list of new hedonic updates
app.directive('arbHedonsModePanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '='
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.title = 'Achievements';
			$scope.moreLink = '/achievements';

			pageService.loadModeData('/json/hedons/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
			});
		},
	};
});

// arb-likes-row is the directive for showing who liked current user's stuff
app.directive('arbLikesRow', function(pageService, userService) {
	return {
		templateUrl: 'static/html/likesModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.userNames = formatUserNamesForDisplay($scope.modeRow.userIds.map(function(userId) {
				return userService.getFullName(userId);
			}));
		},
	};
});

// arb-reqs-taught-row is the directive for showing who learned current user's reqs
app.directive('arbReqsTaughtRow', function(pageService, userService) {
	return {
		templateUrl: 'static/html/reqsTaughtModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.userNames = formatUserNamesForDisplay($scope.modeRow.userIds.map(function(userId) {
				return userService.getFullName(userId);
			}));
			$scope.reqNames = formatReqsNamesForDisplay($scope.modeRow.requisiteIds.map(function(pageMap) {
				return pageService.pageMap[pageMap].title;
			}));
		},
	};
});

// arb-hedons-mode-page is for displaying the entire /achievements page
app.directive('arbHedonsModePage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/hedonsModePage.html',
		scope: {
		},
	};
});
