'use strict';

// arb-discussion-mode-page hosts the arb-discussion-mode-panel
app.directive('arbDiscussionModePage', function($http, arb) {
	return {
		templateUrl: 'static/html/discussionModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});

// arb-discussion-mode-panel directive displays a list of things to discussion in a panel
app.directive('arbDiscussionModePanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.title = 'Discussion';

			arb.stateService.postData('/json/discussionMode/', {
					numPagesToLoad: $scope.numToDisplay,
				},
				function(data) {
					$scope.modeRows = data.result.modeRows;
					$scope.lastView = data.result.lastView;
				});
		},
	};
});

// arb-comment-mode-row is the directive for a row of the arb-discussion-mode-panel
app.directive('arbCommentModeRow', function($location, arb) {
	return {
		templateUrl: 'static/html/rows/commentModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.comment = arb.stateService.pageMap[$scope.modeRow.commentId];
		},
	};
});

// arb-comment-activity is the directive for a comment activity control (to be used in an activity row)
app.directive('arbCommentActivity', function($location, arb) {
	return {
		templateUrl: 'static/html/rows/commentActivity.html',
		scope: {
			comment: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.topLevelComment = $scope.comment.getTopLevelComment();
		},
	};
});
