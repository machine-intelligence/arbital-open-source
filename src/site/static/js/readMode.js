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
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.title = 'New';
			$scope.moreLink = '/read';

			pageService.loadModeData('/json/readMode/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
			});
		},
	};
});
