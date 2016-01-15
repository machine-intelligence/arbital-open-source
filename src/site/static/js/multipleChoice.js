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

			// Check if the user has the given mastery
			$scope.hasMastery = function(masteryId) {
				return pageService.masteryMap[masteryId].has;
			};

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function(masteryId) {
				pageService.updateMastery($scope, masteryId, !$scope.hasMastery(masteryId));
			};

			$scope.choiceChanged = function() {
				userService.setQuestionAnswer($scope.index, $scope.choice,
						$scope.knows[$scope.choice], $scope.wants[$scope.choice]);
			};
		},
		link: function(scope, element, attrs) {
			return;
			element.find("ng-transclude > p").prepend($compile("<md-icon>help_outline</md-icon>")(scope));
			var answerValue = "a";
			element.find("ng-transclude > ul > li").each(function () {
				$(this).find("ul > li").each(function() {
					var text = $(this).text();
					if (text.startsWith("knows:")) {
						$(this).children("a").each(function() {
							//if (!(answerValue:q
							$scope.knows[answerValue].push($(this).attr("page-id"));
						});
					} else if (text.startsWith("wants:")) {
						$(this).children("a").each(function() {
							$scope.wants[answerValue].push($(this).attr("page-id"));
						});
					}
				});
				//$(this).children("ul").remove();
				$(this).changeElementType("md-radio-button")
					.addClass("md-primary")
					.attr("value", answerValue);
				answerValue = String.fromCharCode(answerValue.charCodeAt() + 1);
			});
			console.log($scope.knows);
			console.log($scope.wants);
			var $ul = element.find("ng-transclude > ul")
				.changeElementType("md-radio-group")
				.attr("ng-model", "choice")
				.attr("ng-change", "choiceChanged()");
			//$compile($ul)(scope);
		},
	};
});

