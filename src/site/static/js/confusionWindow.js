'use strict';

// Directive for showing a window after a user said they were confused.
app.directive('arbConfusionWindow', function(pageService, userService) {
	return {
		templateUrl: 'static/html/confusionWindow.html',
		scope: {
			// Id of the confusion mark that was created.
			markId: '@',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.confusionText = "";

			$scope.updateMarkText = function() {
				pageService.updateMark({
						markId: $scope.markId,
						text: $scope.confusionText,
					},
					function(data) {
						$scope.dismissWindow();
					}
				);
			};

			$scope.dismissWindow = function() {
				pageService.hideEvent();
			};
		},
	};
});
