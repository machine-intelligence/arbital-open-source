'use strict';

import app from './angular.ts';

// arb-discussion-mode-page hosts the arb-discussion-mode-panel
app.directive('arbDiscussionModePage', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/discussionModePage.html'),
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
		templateUrl: versionUrl('static/html/listPanel.html'),
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
		templateUrl: versionUrl('static/html/rows/commentModeRow.html'),
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.comment = arb.stateService.pageMap[$scope.modeRow.commentId];
		},
	};
});

// arb-comment-row-internal is the directive for the guts of a comment row
app.directive('arbCommentRowInternal', function($location, arb) {
	return {
		templateUrl: versionUrl('static/html/rows/commentRowInternal.html'),
		scope: {
			comment: '=',
			onDismiss: '&',
			update: '=', // optional, used to display the time of the comment
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.topLevelComment = $scope.comment.getTopLevelComment();

			$scope.threadLoaded = false;
			arb.pageService.loadCommentThread($scope.topLevelComment.pageId, {
				success: function() {
					$scope.threadLoaded = true;
				},
			});
		},
	};
});
