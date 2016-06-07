'use strict';

// arb-read-mode-page hosts the arb-read-mode-panel
app.directive('arbReadModePage', function($http, arb) {
	return {
		templateUrl: 'static/html/readModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});

// arb-read-mode-panel directive displays a list of things to read in a panel
app.directive('arbReadModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			type: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData('/json/readMode/', {
					type: $scope.type,
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
				});
		},
	};
});
