'use strict';

import app from './angular.ts';

import {
	formatUsersForDisplay,
	formatReqsForDisplay,
} from './util.ts';

// directive for an update expand button
app.directive('arbUpdateRowExpandButton', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/updates/updateRowExpandButton.html'),
		scope: false,
		require: '^updateRow'
	};
});

// directive for an update like button
app.directive('arbChangeLogRowLikeButton', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/updates/changeLogRowLikeButton.html'),
		scope: false,
		require: '^updateRow'
	};
});

// directive for an update dismiss button
app.directive('arbUpdateRowDismissButton', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/updates/updateRowDismissButton.html'),
		scope: false,
		require: '^updateRow'
	};
});

// directive for an update timestamp
app.directive('arbUpdateTimestamp', function(arb) {
	return {
		template: '<span class="md-caption nowrap" ng-bind="::(update.createdAt || changeLog.createdAt || createdAt | smartDateTime)"></span>',
		scope: false,
		require: '^updateRow'
	};
});

var getUpdateRowDirectiveFunc = function(templateUrl, controllerInternal = null) {
	return function(arb) {
		return {
			templateUrl: templateUrl,
			scope: {
				// One of the following three things must be provided
				changeLog: '=',
				submission: '=',
				update: '=',
				// Optional, only shown if update is provided
				onDismiss: '&',
			},
			controller: function($scope) {
				$scope.arb = arb;

				// Fill these in to make the following code easier.
				$scope.changeLog = $scope.changeLog || {};
				if ($scope.update && $scope.update.changeLog) {
					$scope.changeLog = $scope.update.changeLog;
				}
				$scope.submission = $scope.submission || {};

				$scope.goToPageId = $scope.changeLog.pageId || $scope.submission.pageId || $scope.update.goToPageId;
				$scope.byUserId = $scope.changeLog.userId || $scope.submission.submitterId || $scope.update.byUserId;
				$scope.createdAt = $scope.changeLog.createdAt || $scope.submission.createdAt || $scope.update.createdAt;
				$scope.domainId = $scope.update ? $scope.update.subscribedToId : $scope.submission.domainId;
				$scope.goToPage = arb.stateService.pageMap[$scope.goToPageId];

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
app.directive('arbPageEditUpdateRow', getUpdateRowDirectiveFunc(versionUrl('static/html/rows/updates/pageEditUpdateRow.html')));

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
				$scope.acquireeId = $scope.update.subscribedToId;
				$scope.acquirerId = $scope.update.goToPageId;
				break;
			case 'questionMergedReverse':
				$scope.acquireeId = $scope.update.goToPageId;
				$scope.acquirerId = $scope.update.subscribedToId;
				break;
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

// arb-user-trust-mode-row is the directive for showing that the user trust has changed
app.directive('arbUserTrustModeRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/userTrustModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.update = $scope.modeRow.update;
		},
	};
});
