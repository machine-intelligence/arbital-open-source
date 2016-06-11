'use strict';

// arb-updates-panel directive displays a list of new maintenance updates
app.directive('arbUpdatesPanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			title: '@',
			moreLink: '@',
			postUrl: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData($scope.postUrl, {
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

			$scope.isRelationshipChangeLogType = function(changeLogType) {
				switch (changeLogType) {
					case "newParent":
					case "newChild":
					case "newLens":
					case "newTag":
					case "newUsedAsTag":
					case "newRequirement":
					case "newRequiredBy":
					case "newSubject":
					case "newTeacher":

					case "deleteParent":
					case "deleteChild":
					case "deleteTag":
					case "deleteUsedAsTag":
					case "deleteRequirement":
					case "deleteRequiredBy":
					case "deleteSubject":
					case "deleteTeacher":
						return true;
				}
				return false;
			};
		},
	};
});

// arb-update-row is the directive for showing an update
app.directive('arbUpdateRow', function(arb) {
	return {
		templateUrl: 'static/html/rows/updateRow.html',
		transclude: true,
		scope: {
			update: '=',
			onDismiss: '=',
			showLikeButton: '=',
		},
	};
});

var getUpdateRowDirectiveFunc = function(templateUrl, controllerInternal) {
	return function(arb) {
		return {
			templateUrl: templateUrl,
			scope: {
				update: '=',
				onDismiss: '&',
			},
			controller: function($scope) {
				$scope.arb = arb;
				$scope.goToPage = arb.stateService.pageMap[$scope.update.goToPageId];
				if (controllerInternal) controllerInternal($scope);
			},
		};
	};
};

app.directive('arbCommentUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/commentUpdateRow.html',
	function($scope) {
		$scope.comment = $scope.arb.stateService.pageMap[$scope.update.goToPageId];
	})
);
app.directive('arbPageToDomainUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/pageToDomainUpdateRow.html'));
app.directive('arbAtMentionUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/atMentionUpdateRow.html'));
app.directive('arbNewMarkUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/newMarkUpdateRow.html'));
app.directive('arbResolvedThreadUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/resolvedThreadUpdateRow.html'));
app.directive('arbResolvedMarkUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/resolvedMarkUpdateRow.html'));
app.directive('arbAnsweredMarkUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/answeredMarkUpdateRow.html'));
app.directive('arbPageEditUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/pageEditUpdateRow.html'));
app.directive('arbQuestionMergedUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/questionMergedUpdateRow.html'));
app.directive('arbQuestionMergedReverseUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/questionMergedReverseUpdateRow.html'));
app.directive('arbRelationshipUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/relationshipUpdateRow.html'));

// arb-maintenance-update-row is the directive for showing a maintenance update
app.directive('arbMaintenanceUpdateRow', function(arb) {
	return {
		templateUrl: 'static/html/rows/maintenanceUpdateRow.html',
		scope: {
			modeRow: '=',
			onDismiss: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
			$scope.showUserLink = $scope.update.subscribedToId != $scope.update.byUserId;
			$scope.showDismissIcon = true;
		},
	};
});

// arb-bell-updates-page is for displaying the entire /notifications page
app.directive('arbBellUpdatesPage', function($http, arb) {
	return {
		templateUrl: 'static/html/bellUpdatesPage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});

// arb-maintenance-mode-page is for displaying the entire /maintain page
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
