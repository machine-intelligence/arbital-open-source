'use strict';

// Directive for a checkbox
app.directive('arbCheckbox', function($timeout, $http, $compile, pageService, userService) {
	return {
		templateUrl: 'static/html/checkbox.html',
		transclude: true,
		scope: {
			index: '@',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.choice = false;
			$scope.knows = [];
			$scope.wants = [];

			// Called when a user toggles the choice
			$scope.toggleChoice = function() {
				$scope.choice = !$scope.choice;
				pageService.setQuestionAnswer($scope.index, $scope.choice,
					$scope.choice ? $scope.knows : [], $scope.choice ? $scope.wants : []);
			};
		},
		link: function(scope, element, attrs) {
			var buttonHtml = '<md-button class=\'md-icon-button\' ng-click=\'toggleChoice()\'>' +
			'	<md-icon ng-if=\'choice\'>' +
			'		check_box' +
			'	</md-icon>' +
			'	<md-icon ng-if=\'!choice\'>' +
			'		check_box_outline_blank' +
			'	</md-icon>' +
			'</md-button>';
			element.find('ng-transclude > p').prepend($compile(buttonHtml)(scope));

			// Extract "knows" and "wants"
			element.find('ng-transclude > ul > li > p').each(function() {
				var text = $(this).text();
				if (text.indexOf('knows:') == 0) {
					$(this).children('a').each(function() {
						scope.knows.push($(this).attr('page-id'));
					});
				} else if (text.indexOf('wants:') == 0) {
					$(this).children('a').each(function() {
						scope.wants.push($(this).attr('page-id'));
					});
				}
			});
			element.find('ng-transclude > ul').remove();
		},
	};
});

