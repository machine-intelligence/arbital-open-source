'use strict';

import app from './angular.ts';

// arb-continue-writing-mode-panel directive displays a list of things that prompt a user
// to continue writing, like their drafts or stubs
app.directive('arbContinueWritingModePanel', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/listPanel.html'),
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData('/json/continueWriting/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
				});
		},
	};
});

// arb-write-new-mode-panel displays a list of things that prompt a user
// to contribute new content, like redlinks and requests
app.directive('arbWriteNewModePanel', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/writeNewPanel.html'),
		scope: {
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData('/json/writeNew/', {
					isFullPage: $scope.isFullPage,
				},
				function(data) {
					$scope.redLinkRows = data.result.redLinks.map(function(redLink) {
						var aliasWithSpaces = redLink.alias.replace(/_/g, ' ');
						var prettyName = aliasWithSpaces.charAt(0).toUpperCase() + aliasWithSpaces.slice(1);

						redLink.requestType = 'redLink';
						redLink.prettifiedAlias = prettyName;
						redLink.originalTotalLikeCount = redLink.likeCount + redLink.myLikeValue;
						return redLink;
					});

					$scope.contentRequestRows = data.result.contentRequests.map(function(contentRequest) {
						contentRequest.originalTotalLikeCount = contentRequest.likeCount + contentRequest.myLikeValue;
						return contentRequest;
					});

					$scope.requests = $scope.redLinkRows.concat($scope.contentRequestRows).filter(function(request) {
						return request.originalTotalLikeCount > 0;
					});
				}
			);
		},
	};
});

// arb-pending-mode-panel displays a list of edits and pages that are pending approval.
app.directive('arbPendingModePanel', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/pendingPanel.html'),
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData('/json/pending/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.pageToDomainSubmissionRows = data.result.pageToDomainSubmissionRows;
					$scope.editProposalRows = data.result.editProposalRows;
				});
		},
	};
});

// arb-page-to-domain-submission-row is the directive for showing a page a user submitted to a domain
app.directive('arbPageToDomainSubmissionRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/pageToDomainSubmissionRow.html'),
		scope: {
			submission: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.submission.pageId];
		},
	};
});

// arb-edit-proposal-row is the directive for showing a page a user submitted to a domain
app.directive('arbEditProposalRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/editProposalRow.html'),
		scope: {
			changeLog: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			// Check if the edit was proposed for a version that's no longer live
			$scope.page = arb.stateService.pageMap[$scope.changeLog.pageId];
			$scope.isObsolete = $scope.page.currentEdit != $scope.page.editHistory[$scope.changeLog.edit].prevEdit;

			$scope.showApprove = function() {
				return $scope.changeLog.type == 'newEditProposal' && !$scope.isObsolete && $scope.page.permissions.edit.has;
			};
		},
	};
});

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbDraftRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/draftRow.html'),
		scope: {
			modeRow: '=',
		},
	};
});

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbTaggedForEditRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/taggedForEditRow.html'),
		scope: {
			modeRow: '=',
		},
	};
});

// arb-explanation-request-row shows an explanation request
app.directive('arbExplanationRequestRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/explanationRequestRow.html'),
		scope: {
			request: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.editUrl = arb.urlService.getEditPageUrl($scope.alias);

			$scope.stopSuggesting = function() {
				arb.signupService.processLikeClick($scope.row, $scope.row.alias, -1);
			};

			$scope.toggleExpand = function() {
 				$scope.expanded = !$scope.expanded;
 			};
		},
	};
});
