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

			$scope.getChangeLogCategory = function(changeLogType) {
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

					case "answerChange":
						return "relationship";

					case "newAlias":
					case "newSortChildrenBy":
					case "setVoteType":
					case "newEditGroup":
					case "lensOrderChanged":
					case "turnOnVote":
					case "turnOffVote":
					case "searchStringChange":
						return "settings";

					case "deletePage":
					case "undeletePage":
						return "deletedPage";
				}
				return false;
			};
		},
	};
});

// arb-update-row is the directive for showing an update
app.directive('arbUpdateRow', function(arb) {
	return {
		templateUrl: 'static/html/rows/updates/updateRow.html',
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

app.directive('arbAtMentionUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/atMentionUpdateRow.html'));
app.directive('arbDeletedPageUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/deletedPageUpdateRow.html'));
app.directive('arbPageEditUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/pageEditUpdateRow.html'));
app.directive('arbPageToDomainUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/pageToDomainUpdateRow.html'));
app.directive('arbRelationshipUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/relationshipUpdateRow.html'));
app.directive('arbResolvedThreadUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/resolvedThreadUpdateRow.html'));
app.directive('arbSettingsUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/settingsUpdateRow.html'));
app.directive('arbCommentUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/commentUpdateRow.html',
	function($scope) {
		$scope.comment = $scope.arb.stateService.pageMap[$scope.update.goToPageId];
	})
);
app.directive('arbMarkUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/markUpdateRow.html',
	function($scope) {
		$scope.markType = $scope.arb.markService.markMap[$scope.update.markId].type;
	})
);
app.directive('arbQuestionMergedUpdateRow', getUpdateRowDirectiveFunc('static/html/rows/updates/questionMergedUpdateRow.html',
	function($scope) {
		switch ($scope.update.type) {
			case 'questionMerged':
				$scope.acquireeId = update.subscribedToId;
				$scope.acquirerId = update.goToPageId;
			case 'questionMergedReverse':
				$scope.acquireeId = update.goToPageId;
				$scope.acquirerId = update.subscribedToId;
		}
	})
);

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
