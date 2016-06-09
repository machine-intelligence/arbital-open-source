'use strict';

// arb-continue-writing-mode-panel directive displays a list of things that prompt a user
// to continue writing, like their drafts or stubs
app.directive('arbContinueWritingModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
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
		templateUrl: 'static/html/writeNewPanel.html',
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

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbDraftRow', function(arb) {
	return {
		templateUrl: 'static/html/rows/draftRow.html',
		scope: {
			modeRow: '=',
		},
	};
});

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbTaggedForEditRow', function(arb) {
	return {
		templateUrl: 'static/html/rows/taggedForEditRow.html',
		scope: {
			modeRow: '=',
		},
	};
});

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbExplanationRequestRow', function(arb) {
	return {
		templateUrl: 'static/html/rows/explanationRequestRow.html',
		scope: {
			alias: '@',
			refCount: '@',
		},
		controller: function($scope) {
			var aliasWithSpaces = $scope.alias.replace(/_/g, ' ');
			$scope.prettyName = aliasWithSpaces.charAt(0).toUpperCase() + aliasWithSpaces.slice(1);

			$scope.editUrl = arb.urlService.getEditPageUrl($scope.alias);

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
		},
	};
});
