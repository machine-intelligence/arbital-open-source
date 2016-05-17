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
			isFullPage: '=',
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
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '='
		},
		controller: function($scope) {
			$scope.rowTemplate = 'hedons';
			$scope.title = 'Achievements';
			$scope.moreLink = "/achievements";
			$scope.listItemIsNew = function(listItem) {
				return listItem.createdAt > $scope.lastView;
			};

			$http({method: 'POST', url: '/json/hedons/', data: JSON.stringify({})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.items = Object.keys(data.result.newLikes).map(function(key) {
						return data.result.newLikes[key];
					});
					$scope.lastView = data.result.lastAchievementsView;
				});


		},
	};
});

// arb-hedons-row is the directive for a row of the arb-hedons-panel
app.directive('arbHedonsRow', function(pageService) {
	return {
		templateUrl: 'static/html/hedonsRow.html',
		replace: true,
		scope: {
			newLike: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;

			$scope.getNames = function(newLikeRow) {
				var names = newLikeRow.names;
				if (names.length == 1) {
					return names[0] + ' likes ';
				}

				if (names.length == 2) {
					return names[0] + ' and ' + names[1] + ' like '
				}

				var numExtraPeople = names.length - 2;
				var namesString = names[0] + ', ' + names[1] + ', and ' + numExtraPeople + ' other ';
				namesString = namesString + ((numExtraPeople == 1) ? 'person' : 'people') + ' like ';
				return namesString
			};
		},
	};
});

// arb-read-mode-panel directive displays a list of things to read in a panel
app.directive('arbReadModePanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.rowTemplate = 'page';
			$scope.title = 'New reading';
			$scope.moreLink = "/read";
			$scope.listItemIsNew = function(pageId) {
				return pageService.pageMap[pageId].createdAt > $scope.lastView;
			};

			$http({method: 'POST', url: '/json/readMode/', data: JSON.stringify({})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.items = data.result.hotPageIds;
					$scope.lastView = data.result.lastReadModeView;
				});
		},
	};
});

// Exists to share the template for a row in a md-list of pages
app.directive('arbPageRow', function() {
	return {
		templateUrl: 'static/html/pageRow.html',
		scope: {
			pageId: '=',
		},
		replace: true,
	};
});
