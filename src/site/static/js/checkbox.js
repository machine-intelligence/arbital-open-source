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
			default: '@',
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

			// Called when a user toggles the choice (or sets it to the given value)
			$scope.toggleChoice = function(newChoice) {
				if (newChoice === undefined) {
					$scope.choice = !$scope.choice;
				} else {
					$scope.choice = newChoice;
				}
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
				arb.pathService.extendPath($scope.index, $scope.path[$scope.letterChoice]);
			};
		},
		link: function(scope, element, attrs) {
			var buttonHtml = '<md-button class=\'md-icon-button small-button\' ng-click=\'toggleChoice()\' aria-label=\'Toggle\'>' +
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

			// Add the element which shows what pages were added
			var addedPagesHtml = [
				'<div class="md-caption" ng-if="path[letterChoice].length > 0 && arb.pathService.isOnPath()">',
				'<span>Added to path:</span>',
				'<span ng-repeat="(index, pageId) in path[letterChoice]">',
				'	<span class="comma" ng-if="index > 0">,</span>',
				'	<arb-page-title page-id="{{::pageId}}" is-link="true"></arb-page-title>',
				'</span>',
				'</div>'].join('');
			element.find('ng-transclude').append($compile(addedPagesHtml)(scope));

			$timeout(function() {
				if (element.closest('arb-markdown').length > 0) {
					// Process all math.
					arb.markdownService.compileChildren(scope, element, {skipCompile: true});
				}

				// Restore the choice value set last time, or set the default
				var pageObject = arb.masteryService.getPageObject(scope.pageId, scope.objectAlias);
				if (pageObject) {
					scope.toggleChoice(pageObject.value);
				} else {
					scope.toggleChoice(scope.default == 'y');
				}
			});
		},
	};
});

