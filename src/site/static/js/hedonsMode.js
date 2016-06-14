'use strict';

// arb-hedons-mode-panel directive displays a list of new hedonic updates
app.directive('arbHedonsModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
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
