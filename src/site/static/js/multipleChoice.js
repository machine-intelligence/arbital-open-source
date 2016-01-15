"use strict";

// Directive for multiple choice
app.directive("arbMultipleChoice", function($timeout, $http, $compile, pageService, userService) {
	return {
		templateUrl: "/static/html/multipleChoice.html",
		transclude: true,
		scope: {
			index: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.choice = "";
			$scope.knows = {};
			$scope.wants = {};

			// Called when a user makes a choice
			$scope.choiceChanged = function() {
				pageService.setQuestionAnswer($scope.index, $scope.choice,
						$scope.knows[$scope.choice], $scope.wants[$scope.choice]);
			};
		},
		link: function(scope, element, attrs) {
			element.find("ng-transclude > p").prepend($compile("<md-icon>help_outline</md-icon>")(scope));
			var answerValue = "a";
			// Go through all answers
			element.find("ng-transclude > ul > li").each(function () {
				// For each answer, extract "knows" and "wants"
				scope.knows[answerValue] = [];
				scope.wants[answerValue] = [];
				$(this).find("ul > li").each(function() {
					var text = $(this).text();
					if (text.startsWith("knows:")) {
						$(this).children("a").each(function() {
							scope.knows[answerValue].push($(this).attr("page-id"));
						});
					} else if (text.startsWith("wants:")) {
						$(this).children("a").each(function() {
							scope.wants[answerValue].push($(this).attr("page-id"));
						});
					}
				});
				$(this).children("ul").remove();
				$(this).changeElementType("md-radio-button")
					.addClass("md-primary")
					.attr("value", answerValue);
				answerValue = String.fromCharCode(answerValue.charCodeAt() + 1);
			});
			var $ul = element.find("ng-transclude > ul")
				.changeElementType("md-radio-group")
				.attr("ng-model", "choice")
				.attr("ng-change", "choiceChanged()");
			$compile($ul)(scope);
		},
	};
});

