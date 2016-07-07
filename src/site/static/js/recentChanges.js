'use strict';

// arb-recent-changes displays a list of recent changes
app.directive('arbRecentChanges', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/listPanel.html'),
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			type: '@'
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData('/json/recentChanges/', {
					numToLoad: $scope.numToDisplay,
					changeType: $scope.type || 'edits',
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
				});
		},
	};
});


// arb-recent-changes displays a list of recent changes
app.directive('arbRecentChangesPage', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/recentChangesPage.html'),
		scope: {},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
