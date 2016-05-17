'use strict';

// arb-read-mode-page hosts the arb-read-mode-panel
app.directive('arbReadModePage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/readModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
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

			$http({method: 'POST', url: '/json/readMode/', data: JSON.stringify({})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.items = data.result.hotPageIds.map(function(pageId) {
						return pageService.pageMap[pageId];
					});
					$scope.lastView = data.result.lastReadModeView;
				});
		},
	};
});

