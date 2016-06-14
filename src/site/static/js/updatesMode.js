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

			// Given the type of the the changelog update, return the type of row we should use
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
				}
				return false;
			};
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
