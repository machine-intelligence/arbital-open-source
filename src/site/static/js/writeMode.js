'use strict';

// arb-continue-writing-mode-panel directive displays a list of things that prompt a user
// to continue writing, like their drafts or stubs
app.directive('arbContinueWritingModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			hideTitle: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.title = 'Continue writing';

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
app.directive('arbWriteModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			hideTitle: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.title = 'Write new';

			arb.stateService.postData('/json/writeNew/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
				});
		},
	};
});

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbDraftRow', function(arb) {
	return {
		templateUrl: 'static/html/draftRow.html',
		scope: {
			modeRow: '=',
		},
	};
});

// arb-draft-mode-row is the directive for showing a user's draft
app.directive('arbTaggedForEditRow', function(arb) {
	return {
		templateUrl: 'static/html/taggedForEditRow.html',
		scope: {
			modeRow: '=',
		},
	};
});
