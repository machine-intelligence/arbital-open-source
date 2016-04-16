'use strict';

// Directive for showing a window after a user said they were confused.
app.directive('arbConfusionWindow', function($interval, pageService, userService, autocompleteService) {
	return {
		templateUrl: 'static/html/confusionWindow.html',
		scope: {
			// Id of the confusion mark that was created.
			markId: '@',
			// Set to true if the user just created this mark.
			isNew: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.mark = pageService.markMap[$scope.markId];
			$scope.isOnPage = $scope.mark.pageId == pageService.getCurrentPageId();

			// Update mark's text.
			$scope.updateMarkText = function() {
				pageService.updateMark({
						markId: $scope.markId,
						text: $scope.mark.text,
					},
					function(data) {
						$scope.hideEventWindow();
					}
				);
				$scope.mark.resolvedPageId = '';
				$scope.mark.resolvedBy = '';
			};

			// Search for similar questions / pages
			$scope.similarPages = [];
			var lastTerm = $scope.mark.text;
			var findSimilarFunc = function() {
				if ($scope.mark.text === lastTerm) return;
				lastTerm = $scope.mark.text;
				var data = {
					term: $scope.mark.text,
				};
				autocompleteService.performSearch(data, function(data) {
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

			// Call to resolve the mark with the given page.
			$scope.resolveWith = function(pageId, hideWindow) {
				pageService.updateMark({
					markId: $scope.markId,
					resolvedPageId: pageId,
				});
				$scope.mark.resolvedPageId = pageId;
				$scope.mark.resolvedBy = userService.user.id;
				if (hideWindow) {
					$scope.hideEventWindow();
				}
			};

			$scope.suggestedLinkClicked = function(pageId, event) {
				// Ctrl + click opens a new tab, so we shouldn't close the window
				var hideWindow = !(event && event.ctrlKey && event.type === 'click');
				$scope.resolveWith(pageId, hideWindow);
			};

			// Show the input to connect the mark to a question.
			var showQuestionInput = false;
			$scope.showConnectToQuestion = function() {
				$scope.showQuestionInput = true;
			};

			// Called when a user selects a question to match to this mark.
			$scope.questionLinked = false;
			$scope.questionResultSelected = function(result) {
				$scope.resolveWith(result.pageId, false);
				$scope.questionLinked = true;
				$scope.showQuestionInput = false;
			};

			// Called when an author wants to resolve the mark.
			$scope.dismissMark = function() {
				pageService.updateMark({
					markId: $scope.markId,
					dismiss: true,
				});
				$scope.mark.resolvedPageId = '';
				$scope.mark.resolvedBy = userService.user.id;
				$scope.hideEventWindow();
			};
		},
		link: function(scope, element, attrs) {
			// Hide current event window, if it makes sense.
			var isInsideEventWindow = element.closest('#events-info-div').length > 0;
			scope.hideEventWindow = function() {
				if (isInsideEventWindow) {
					pageService.hideEvent();
				}
			};
		},
	};
});
