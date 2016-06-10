'use strict';

// Directive for hidden text (usually for homework problems)
app.directive('arbHiddenText', function(arb) {
	return {
		templateUrl: 'static/html/hiddenText.html',
		transclude: true,
		scope: {
			buttonText: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.revealed = false;
		},
		link: function(scope, element, attrs) {
			scope.reveal = function() {
				scope.revealed = true;
				arb.markdownService.compileChildren(scope, element.find("[ng-transclude]"));
			};
		},
	};
});

