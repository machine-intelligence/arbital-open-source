'use strict';

// Directive for showing a window for creating/editing a query
app.directive('arbQueryInfo', function($interval, pageService, userService, autocompleteService) {
	return {
		templateUrl: 'static/html/queryInfo.html',
		scope: {
			// Id of the query mark that was created.
			markId: '@',
			// Set to true if the user just created this mark.
			isNew: '=',
			// Set to true if this is in a popup
			inPopup: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.mark = pageService.markMap[$scope.markId];
			$scope.potentialResolvedPageId = undefined;

			// Return true if the user is on the page where the mark was created.
			$scope.isOnPage = function() {
				return $scope.mark.pageId == pageService.getCurrentPageId();
			};

			// Update mark's text.
			$scope.updateMarkText = function(submit) {
				pageService.updateMark({
						markId: $scope.markId,
						text: $scope.mark.text,
						submit: submit,
					},
					function(data) {
						if (submit) {
							$scope.hidePopup();
						}
					}
				);
				if (submit) {
					$scope.mark.resolvedPageId = '';
					$scope.mark.resolvedBy = '';
				}
			};

			// Search for similar questions / pages
			$scope.responses = [];
			var lastTerm = $scope.mark.resolvedBy ? $scope.mark.text : '';
			var findSimilarFunc = function() {
				if (!$scope.mark.isCurrentUserOwned) return;
				if ($scope.mark.text === lastTerm || $scope.mark.text.length <= 0) return;
				lastTerm = $scope.mark.text;
				var data = {
					term: $scope.mark.text,
				};
				autocompleteService.performSearch(data, function(data) {
					$scope.responses.length = 0;
					for (var n = 0; n < data.length; n++) {
						var pageId = data[n].pageId;
						$scope.responses.push({pageId: pageId, score: data[n].score});
					}
				});
			};
			var similarInterval = $interval(findSimilarFunc, 1000);
			$scope.$on('$destroy', function() {
				$interval.cancel(similarInterval);
			});

			// Call to resolve the mark with the given page.
			$scope.resolveWith = function(pageId) {
				pageService.updateMark({
					markId: $scope.markId,
					resolvedPageId: pageId,
					text: $scope.mark.text,
					submit: true,
				});
				$scope.mark.resolvedPageId = pageId;
				$scope.mark.resolvedBy = userService.user.id;
			};

			// Called when the user clicks one of the suggested links
			$scope.suggestedLinkClicked = function(pageId, event) {
				$scope.updateMarkText(false);
				$scope.potentialResolvedPageId = pageId;
				$scope.isNew = false;
			};

			// Called when the user accepts/rejects the suggested page id.
			$scope.resolveSuggestion = function(accept) {
				if (accept) {
					$scope.resolveWith($scope.potentialResolvedPageId);
					$scope.hidePopup();
				} else {
					$scope.potentialResolvedPageId = undefined;
					$scope.responses = [];
					lastTerm = '';
					findSimilarFunc();
				}
			};

			// Show the input to link the query to a question.
			var showQuestionInput = false;
			$scope.showLinkToQuestion = function() {
				$scope.showQuestionInput = true;
			};

			// Called when a user selects a question to match to this mark.
			$scope.questionLinked = false;
			$scope.questionResultSelected = function(result) {
				$scope.resolveWith(result.pageId);
				$scope.questionLinked = true;
				$scope.showQuestionInput = false;
			};

			// Called when an author wants to resolve the mark.
			$scope.dismissMark = function() {
				pageService.updateMark({
					markId: $scope.markId,
					text: $scope.mark.text,
					dismiss: true,
				});
				$scope.mark.resolvedPageId = '';
				$scope.mark.resolvedBy = userService.user.id;
				$scope.hidePopup({dismiss: true});
			};
		},
		link: function(scope, element, attrs) {
			// Hide current popup, if it makes sense.
			scope.hidePopup = function(result) {
				if (scope.inPopup) {
					pageService.hidePopup(result);
				}
			};
		},
	};
});
