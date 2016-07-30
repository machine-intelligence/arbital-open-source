'use strict';

import app from './angular.ts';

// Directive for hidden text (usually for homework problems)
app.directive('arbHiddenText', function($compile, $timeout, arb) {
	return {
		scope: {
			buttonText: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
		link: function(scope: any, element, attrs) {
			if (!scope.buttonText) return;

			$timeout(function() {
				$(element).prepend($compile('<md-button class="md-primary md-hue-1 md-raised"' +
					'ng-bind="buttonText"' +
					'ng-click="toggle()"' +
					'aria-label="{{buttonText}}">' +
					'</md-button>')(scope));
			});
			scope.toggle = function() {
				$(element).find('.hidden-text').toggleClass('display-none');
			};
		},
	};
});

