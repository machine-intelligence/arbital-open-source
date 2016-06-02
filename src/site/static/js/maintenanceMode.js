'use strict';

// arb-maintenance-mode-panel directive displays a list of new maintenance updates
app.directive('arbMaintenanceModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			hideTitle: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.userService.user.maintenanceUpdateCount = 0;
			$scope.title = 'Maintenance Updates';
			$scope.moreLink = '/maintain';

			arb.stateService.postData('/json/maintain/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
				});
		},
	};
});

// arb-maintenance-update-row is the directive for showing a maintenance update
app.directive('arbMaintenanceUpdateRow', function(arb) {
	return {
		templateUrl: 'static/html/maintenanceUpdateRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.changeLog = $scope.modeRow.update.changeLog;
			$scope.byUserId = $scope.modeRow.update.byUserId;
			$scope.showUserLink = $scope.modeRow.update.subscribedToId != $scope.modeRow.update.byUserId;
			$scope.type = $scope.modeRow.update.type;
			$scope.markId = $scope.modeRow.update.markId;
			$scope.subscribedToId = $scope.modeRow.update.subscribedToId;
			$scope.goToPageId = $scope.modeRow.update.goToPageId;
			$scope.isRelatedPageAlive = $scope.modeRow.update.isGoToPageAlive;
			$scope.createdAt = $scope.modeRow.update.createdAt;
			$scope.repeated = $scope.modeRow.update.repeated;
			$scope.showDismissIcon = true;
			// ROGTODO: handle dismiss
			// $scope.onDismiss = dismissUpdate(update, group.updates, index)
		},
	};
});


// arb-maintenance-mode-page is for displaying the entire /achievements page
app.directive('arbMaintenanceModePage', function($http, arb) {
	return {
		templateUrl: 'static/html/maintenanceModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
