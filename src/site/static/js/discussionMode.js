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

			$scope.rowTemplate = 'discussion';
			$scope.title = 'Discussion';
			$scope.moreLink = '/discussion';

			$http({method: 'POST', url: '/json/discussionMode/', data: JSON.stringify({})})
				.success(function(data) {
					console.log('/json/discussionMode/ data:'); console.log(data);
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.items = data.result.commentIds.map(function(commentId) {
						return pageService.pageMap[commentId];
					});
					$scope.lastView = data.result.lastDiscussionModeView;
				});
		},
	};
});

// arb-discussion-mode-row is the directive for a row of the arb-discussion-mode-panel
app.directive('arbDiscussionModeRow', function($location, pageService, userService) {
	return {
		templateUrl: 'static/html/discussionModeRow.html',
		replace: true,
		scope: {
			comment: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.topLevelComment = $scope.comment.getTopLevelComment();
			$scope.threadExpanded = false;
			$scope.threadLoaded = false;

			$scope.toggleThread = function() {
				$scope.threadExpanded = !$scope.threadExpanded;
				if ($scope.threadExpanded && !$scope.threadLoaded) {
					pageService.loadCommentThread($scope.topLevelComment.pageId, {
						success: function() {
							$scope.threadLoaded = true;
						},
					});
				}
			};
		},
	};
});

