'use strict';

// Directive for hidden text (usually for homework problems)
app.directive('arbHiddenText', function($compile, $timeout, arb) {
	return {
		scope: {
			buttonText: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
		link: function(scope, element, attrs) {
			if (!scope.buttonText) return;

			$timeout(function() {
				$(element).prepend($compile('<md-button class="md-primary md-hue-1 md-raised"' +
					'ng-bind="buttonText"' +
					'ng-click="reveal()"' +
					'aria-label="{{buttonText}}"' +
					'ng-if="!revealed">' +
					'</md-button>')(scope));
			});
			scope.reveal = function() {
				scope.revealed = true;
				$(element).find("md-button").remove();
				$(element).find(".display-none").removeClass("display-none");
			};
		},
	};
});

