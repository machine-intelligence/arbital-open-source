"use strict";

// Directive for multiple choice
app.directive("arbMultipleChoice", function($timeout, $http, $compile, pageService, userService) {
	return {
		templateUrl: "static/html/multipleChoice.html",
		transclude: true,
		scope: {
			pageId: "@",
			index: "@",
			objectAlias: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.choice = "";
			$scope.knows = {};
			$scope.wants = {};
			$scope.delKnows = {};
			$scope.delWants = {};

			// Called when a user makes a choice
			$scope.choiceChanged = function() {
				pageService.setQuestionAnswer($scope.index,
						$scope.knows[$scope.choice], $scope.wants[$scope.choice],
						$scope.delKnows[$scope.choice], $scope.delWants[$scope.choice],
						{
							pageId: $scope.pageId,
							edit: pageService.pageMap[$scope.pageId].edit,
							object: $scope.objectAlias,
							value: $scope.choice,
						});
			};
		},
		link: function(scope, element, attrs) {
			element.find("ng-transclude > p").prepend($compile("<md-icon class='question-icon'>help_outline</md-icon>")(scope));
			var answerValue = "a";
			// Go through all answers
			element.find("ng-transclude > ul > li").each(function () {
				// For each answer, extract "knows" and "wants"
				scope.knows[answerValue] = [];
				scope.wants[answerValue] = [];
				scope.delKnows[answerValue] = [];
				scope.delWants[answerValue] = [];
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
					} else if (text.startsWith("-knows:")) {
						$(this).children("a").each(function() {
							scope.delKnows[answerValue].push($(this).attr("page-id"));
						});
					} else if (text.startsWith("-wants:")) {
						$(this).children("a").each(function() {
							scope.delWants[answerValue].push($(this).attr("page-id"));
						});
					}
				});
				$(this).children("ul").remove();
				$(this).changeElementType("div")
				.prepend("<md-radio-button class='md-primary' aria-label='Answer " + answerValue +
					"' value='" + answerValue + "'></md-radio-button>");
				answerValue = String.fromCharCode(answerValue.charCodeAt() + 1);
			});
			var $ul = element.find("ng-transclude > ul")
				.changeElementType("md-radio-group")
				.attr("ng-model", "choice")
				.attr("ng-change", "choiceChanged()");
			$compile($ul)(scope);

			// Since user's requisites might have changed since they answered this question,
			// we'll see if we can find an answer that matches their current state.
			var possibleAnswers = Object.keys(scope.knows).concat(Object.keys(scope.wants))
					.concat(Object.keys(scope.delKnows)).concat(Object.keys(scope.delWants));
			var processedAnswers = {}; // used to prevent processing the same answer twice
			for (var a = 0; a < possibleAnswers.length; a++) {
				var possibleAnswer = possibleAnswers[a];
				if (possibleAnswer in processedAnswers) continue;
				processedAnswers[possibleAnswer] = true;
				var isAnswerValid = true;
				var knowsList = scope.knows[possibleAnswer];
				if (knowsList) {
					for (var n = 0; n < knowsList.length; n++) {
						if (!pageService.hasMastery(knowsList[n])) {
							isAnswerValid = false;
							break;
						}
					}
					if (!isAnswerValid) continue;
				}
				var wantsList = scope.wants[possibleAnswer];
				if (wantsList) {
					for (var n = 0; n < wantsList.length; n++) {
						if (!pageService.getMasteryStatus(wantsList[n])) {
							isAnswerValid = false;
							break;
						}
					}
					if (!isAnswerValid) continue;
				}
				var delKnowsList = scope.delKnows[possibleAnswer];
				if (delKnowsList) {
					for (var n = 0; n < delKnowsList.length; n++) {
						if (pageService.hasMastery(delKnowsList[n])) {
							isAnswerValid = false;
							break;
						}
					}
					if (!isAnswerValid) continue;
				}
				var delWantsList = scope.delWants[possibleAnswer];
				if (delWantsList) {
					for (var n = 0; n < delWantsList.length; n++) {
						if (pageService.wantsMastery(delWantsList[n])) {
							isAnswerValid = false;
							break;
						}
					}
					if (!isAnswerValid) continue;
				}
				if (!scope.choice) {
					scope.choice = possibleAnswer;
					console.log("Found a good choice for " + scope.objectAlias + ":" + scope.choice);
				} else {
					// We already have a possible answer and just found another one, so give up.
					scope.choice = "";
					console.log("Found another good choice:" + possibleAnswer);
					break;
				}
			}

			if (!scope.choice) {
				// Our fallback plan: recall the answer user gave last time.
				var pageObject = pageService.getPageObject(scope.pageId, scope.objectAlias);
				if (pageObject) {
					scope.choice = pageObject.value;
					console.log("Restored saved choice for " + scope.objectAlias + ":" + scope.choice);
				}
			}
		},
	};
});

