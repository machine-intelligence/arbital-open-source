'use strict';

// arb-discussion-mode-page hosts the arb-discussion-mode-panel
app.directive('arbDiscussionModePage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/discussionModePage.html',
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});

// arb-discussion-mode-panel directive displays a list of things to discussion in a panel
app.directive('arbDiscussionModePanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.title = 'Discussion';
			$scope.moreLink = '/discussion';

			pageService.loadModeData('/json/discussionMode/', {
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
app.directive('arbCommentModeRow', function($location, pageService, userService) {
	return {
		templateUrl: 'static/html/commentModeRow.html',
		scope: {
			modeRow: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.comment = pageService.pageMap[$scope.modeRow.commentId];
			$scope.topLevelComment = $scope.comment.getTopLevelComment();
		},
	};
});

