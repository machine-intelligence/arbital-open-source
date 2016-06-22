'use strict';

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

// arb-write-mode-panel displays a list of things that prompt a user
// to contribute new content, like redlinks and requests
app.directive('arbWriteNewModePanel', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/writeNewPanel.html'),
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			arb.stateService.postData('/json/writeNew/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.redLinkRows = data.result.redLinks;
				});
		},
	};
});

// arb-pending-mode-panel displays a list of edits and pages that are pending approval.
app.directive('arbPendingModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/pendingPanel.html',
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
		templateUrl: 'static/html/rows/pageToDomainSubmissionRow.html',
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
		templateUrl: 'static/html/rows/editProposalRow.html',
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

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbExplanationRequestRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/explanationRequestRow.html'),
		scope: {
			alias: '@',
			row: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.alias = $scope.alias || $scope.row.alias;

			var aliasWithSpaces = $scope.alias.replace(/_/g, ' ');
			$scope.prettyName = aliasWithSpaces.charAt(0).toUpperCase() + aliasWithSpaces.slice(1);
			$scope.editUrl = arb.urlService.getEditPageUrl($scope.alias);
			$scope.wrapper = {};

			$scope.editLinkClicked = function(event) {
				arb.analyticsService.reportEditLinkClick(event);
			};

			$scope.toggleExpand = function() {
				$scope.expanded = !$scope.expanded;

				if ($scope.linkedByPageIds) return;

				arb.stateService.postData('/json/moreRelationships/',
					{
						pageAlias: $scope.alias,
						restrictToMathDomain: true,
					},
					function success(data) {
						$scope.linkedByPageIds = data.result.moreRelationshipIds;
					}
				);
			};

			$scope.stopSuggesting = function() {
				arb.stateService.processLikeClick($scope.row, $scope.row.alias, -1);
			};
		},
	};
});
