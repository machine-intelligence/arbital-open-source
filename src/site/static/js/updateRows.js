'use strict';

// arb-update-row is the directive for showing an update
app.directive('arbUpdateRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/updates/updateRow.html'),
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

app.directive('arbAtMentionUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/atMentionUpdateRow.html')));
app.directive('arbDeletedPageUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/deletedPageUpdateRow.html')));
app.directive('arbPageToDomainSubmissionUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/pageToDomainSubmissionUpdateRow.html')));
app.directive('arbPageToDomainAcceptedUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/pageToDomainAcceptedUpdateRow.html')));
app.directive('arbEditProposalAcceptedUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/editProposalAcceptedUpdateRow.html')));
app.directive('arbRelationshipUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/relationshipUpdateRow.html')));
app.directive('arbResolvedThreadUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/resolvedThreadUpdateRow.html')));
app.directive('arbSettingsUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/settingsUpdateRow.html')));

app.directive('arbPageEditUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/pageEditUpdateRow.html'),
	function($scope) {
		$scope.approveProposal = function() {
			$scope.arb.stateService.postDataWithoutProcessing('/json/approvePageEditProposal/', {
				changeLogId: $scope.update.changeLog.id,
			}, function(data) {
				$scope.update.changeLog.type = 'newEdit';
			});
		};
	})
);

app.directive('arbCommentUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/commentUpdateRow.html'),
	function($scope) {
		$scope.comment = $scope.arb.stateService.pageMap[$scope.update.goToPageId];
	})
);

app.directive('arbMarkUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/markUpdateRow.html'),
	function($scope) {
		$scope.markType = $scope.arb.markService.markMap[$scope.update.markId].type;
	})
);

app.directive('arbQuestionMergedUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/questionMergedUpdateRow.html'),
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

// arb-maintenance-update-row is the directive for showing a maintenance update
app.directive('arbMaintenanceUpdateRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/updates/maintenanceUpdateRow.html'),
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

// arb-likes-mode-row is the directive for showing who liked current user's stuff
app.directive('arbLikesModeRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/likesModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.userNames = formatUsersForDisplay($scope.modeRow.userIds.map(function(userId) {
				return arb.userService.getFullName(userId);
			}));
		},
	};
});

// arb-reqs-taught-mode-row is the directive for showing who learned current user's reqs
app.directive('arbReqsTaughtModeRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/reqsTaughtModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.userNames = formatUsersForDisplay($scope.modeRow.userIds.map(function(userId) {
				return arb.userService.getFullName(userId);
			}));
			$scope.reqNames = formatReqsForDisplay($scope.modeRow.requisiteIds.map(function(pageMap) {
				return arb.stateService.pageMap[pageMap].title;
			}));
		},
	};
});

// arb-added-to-group-mode-row is the directive for showing that the user was added to a group
app.directive('arbAddedToGroupModeRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/addedToGroupModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
			if ($scope.update.goToPageId) {
				$scope.goToPage = arb.stateService.pageMap[$scope.update.goToPageId];
			}
		},
	};
});

// arb-removed-from-group-mode-row is the directive for showing that the user was removed from a group
app.directive('arbRemovedFromGroupModeRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/removedFromGroupModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
			if ($scope.update.goToPageId) {
				$scope.goToPage = arb.stateService.pageMap[$scope.update.goToPageId];
			}
		},
	};
});

// arb-invite-received-mode-row is the directive for showing that the user was invited to a domain
app.directive('arbInviteReceivedModeRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/inviteReceivedModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
		},
	};
});

