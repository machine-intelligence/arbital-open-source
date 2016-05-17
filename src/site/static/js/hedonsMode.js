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
			$scope.rowTemplate = 'hedons';
			$scope.title = 'Achievements';
			$scope.moreLink = '/achievements';

			$http({method: 'POST', url: '/json/hedons/', data: JSON.stringify({})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.items = data.result.hedons.sort(function(a, b) {
						a.createdAt > b.createdAt
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
					return names[0] + ' and ' + names[1] + ' like ';
				}

				var numExtraPeople = names.length - 2;
				var namesString = names[0] + ', ' + names[1] + ', and ' + numExtraPeople + ' other ';
				namesString = namesString + ((numExtraPeople == 1) ? 'person' : 'people') + ' like ';
				return namesString;
			};
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
