// panels.js contains directive for panels
'use strict';

// arb-list-panel directive displays a list of things in a panel
app.directive('arbListPanel', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/listPanel.html',
		transclude: true,
		scope: {
			title: '@',
			moreLink: '@',
			items: '=',
			numToDisplay: '=',
			isFullPage: '='
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});

// arb-new-hedons-panel directive displays a list of new hedonic updates
app.directive('arbNewHedonsPanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/newHedonsPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '='
		},
		controller: function($scope) {
			$http({method: 'POST', url: '/json/hedons/', data: JSON.stringify({})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.newLikes = Object.keys(data.result.newLikes).map(function(key) {
						var newLikeRow = data.result.newLikes[key];
						newLikeRow.createdAtDate = new Date(newLikeRow.createdAt);
						return newLikeRow;
					});
				});

			$scope.getNames = function(newLikeRow) {
				var names = newLikeRow.names;
				if (names.length == 1) {
					return names[0];
				}

				if (names.length == 2) {
					return names[0] + ' and ' + names[1]
				}

				return names[0] + ', ' + names[1] + ', and ' + (names.length - 2) + ' other people ';
			};
		},
	};
});

// arb-read-mode-panel directive displays a list of things to read in a panel
app.directive('arbReadModePanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/readModePanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '='
		},
		controller: function($scope) {
			$http({method: 'POST', url: '/json/readMode/', data: JSON.stringify({})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.hotPageIds = data.result.hotPageIds;
				});
		},
	};
});
