'use strict';

// Directive for a checkbox
app.directive('arbCheckbox', function($timeout, $http, $compile, arb) {
	return {
		templateUrl: versionUrl('static/html/checkbox.html'),
		transclude: true,
		scope: {
			pageId: '@',
			index: '@',
			objectAlias: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.choice = false;
			$scope.letterChoice = 'n';
			$scope.knows = {};
			$scope.wants = {};
			$scope.delKnows = {};
			$scope.delWants = {};
			$scope.path = {};

			// Called when a user toggles the choice
			$scope.toggleChoice = function() {
				$scope.choice = !$scope.choice;
				$scope.letterChoice = $scope.choice ? 'y' : 'n';
				arb.masteryService.setQuestionAnswer($scope.index,
						$scope.knows[$scope.letterChoice], $scope.wants[$scope.letterChoice],
						$scope.delKnows[$scope.letterChoice], $scope.delWants[$scope.letterChoice],
						{
							pageId: $scope.pageId,
							edit: arb.stateService.pageMap[$scope.pageId].edit,
							object: $scope.objectAlias,
							value: $scope.letterChoice,
						});
				arb.pathService.extendPath($scope.index, $scope.path[$scope.letterChoice], true);
			};
		},
		link: function(scope, element, attrs) {
			var buttonHtml = '<md-button class=\'md-icon-button\' ng-click=\'toggleChoice()\' aria-label=\'Toggle\'>' +
			'	<md-icon ng-if=\'choice\'>' +
			'		check_box' +
			'	</md-icon>' +
			'	<md-icon ng-if=\'!choice\'>' +
			'		check_box_outline_blank' +
			'	</md-icon>' +
			'</md-button>';
			element.find('ng-transclude > p').prepend($compile(buttonHtml)(scope));
			var answerValue = 'y';

			// Go through all answers
			element.find('ng-transclude > ul > li').each(function() {
				// For each answer, extract "knows" and "wants"
				scope.knows[answerValue] = [];
				scope.wants[answerValue] = [];
				scope.delKnows[answerValue] = [];
				scope.delWants[answerValue] = [];
				scope.path[answerValue] = [];
				$(this).find('ul > li').each(function() {
					var text = $(this).text();
					if (text.indexOf('knows:') == 0) {
						$(this).children('a').each(function() {
							scope.knows[answerValue].push($(this).attr('page-id'));
						});
					} else if (text.indexOf('wants:') == 0) {
						$(this).children('a').each(function() {
							scope.wants[answerValue].push($(this).attr('page-id'));
						});
					} else if (text.indexOf('-knows:') == 0) {
						$(this).children('a').each(function() {
							scope.delKnows[answerValue].push($(this).attr('page-id'));
						});
					} else if (text.indexOf('-wants:') == 0) {
						$(this).children('a').each(function() {
							scope.delWants[answerValue].push($(this).attr('page-id'));
						});
					} else if (text.indexOf('path:') == 0) {
						$(this).children('a').each(function() {
							scope.path[answerValue].push($(this).attr('page-id'));
						});
					}
				});
				answerValue = 'n';
			});
			element.find('ng-transclude > ul').remove();

			$timeout(function() {
				// Process all math.
				arb.markdownService.compileChildren(scope, element, true);
			});
		},
	};
});

