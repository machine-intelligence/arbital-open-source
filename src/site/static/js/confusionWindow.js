'use strict';

// Directive for showing a window after a user said they were confused.
app.directive('arbConfusionWindow', function($interval, pageService, userService, autocompleteService) {
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

			// Search for similar questions / pages
			$scope.similarPages = [];
			var lastTerm = "";
			var findSimilarFunc = function() {
				if ($scope.confusionText === lastTerm) return;
				lastTerm = $scope.confusionText;
				var data = {
					term: $scope.confusionText,
				};
				autocompleteService.performSearch(data, function(data) {
					console.log(data);
					$scope.similarPages.length = 0;
					for (var n = 0; n < data.length; n++) {
						var pageId = data[n].pageId;
						$scope.similarPages.push({pageId: pageId, score: data[n].score});
					}
				});
			};
			var similarInterval = $interval(findSimilarFunc, 1000);
			$scope.$on('$destroy', function() {
				$interval.cancel(similarInterval);
			});

			// Called if the user clicks one of the answer links.
			$scope.answerClicked = function(pageId) {
				pageService.updateMark({
					markId: $scope.markId,
					resolvedPageId: pageId,
				});
				$scope.dismissWindow();
			};
		},
	};
});
