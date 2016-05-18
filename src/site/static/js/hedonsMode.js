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

					$scope.items = data.result.hedons.map(function(hedonsRow) {
						hedonsRow.names = Object.keys(hedonsRow.userIdsMap).map(function(userId) {
							return userService.getFullName(userId);
						});
						if (hedonsRow.requisiteIdsMap) {
							hedonsRow.requisites = Object.keys(hedonsRow.requisiteIdsMap).map(function(reqId) {
								return pageService.pageMap[reqId].title;
							});
						}
						return hedonsRow;
					});

					$scope.items = $scope.items.sort(function(a, b) {
						return b.newActivityAt > a.newActivityAt ? 1 : -1;
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
			hedonsRow: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;

			$scope.getNames = function(hedonsRow) {
				return $scope.formatListForDisplay(hedonsRow.names, 'person', 'people');
			};

			$scope.getReqs = function(hedonsRow) {
				return $scope.formatListForDisplay(hedonsRow.requisites, 'requisite', 'requisites');
			};

			$scope.formatListForDisplay = function(list, singularThing, pluralThing) {
				if (list.length == 1) {
					return list[0];
				}

				if (list.length == 2) {
					return list[0] + ' and ' + list[1];
				}

				var numExtra = list.length - 2;
				return list[0] + ', ' + list[1] + ', and ' + numExtra + ' other ' +
					((numExtra == 1) ? singularThing : pluralThing);
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
