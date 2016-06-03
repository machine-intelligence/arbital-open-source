'use strict';

// arb-response-mode-panel directive displays a list of new notifications
app.directive('arbResponseModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			hideTitle: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.userService.user.newNotificationCount = 0;
			$scope.title = 'Notifications';
			$scope.moreLink = '/notifications';

			arb.stateService.postData('/json/notifications/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
				});

			$scope.dismissRow = function(allRows, index) {
				var update = allRows[index].update;
				$http({method: 'POST', url: '/dismissUpdate/', data: JSON.stringify({
					id: update.id
				})});

				// Remove this update from the list
				allRows.splice(index, 1);
			};
		},
	};
});

// arb-notification-row is the directive for showing an actionable notification
app.directive('arbNotificationRow', function(arb) {
	return {
		templateUrl: 'static/html/notificationRow.html',
		scope: {
			modeRow: '=',
			onDismiss: '&',
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
		},
	};
});

// arb-response-mode-page is for displaying the entire /notifications page
app.directive('arbResponseModePage', function($http, arb) {
	return {
		templateUrl: 'static/html/responseModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
