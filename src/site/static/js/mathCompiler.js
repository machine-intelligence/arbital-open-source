'use strict';

// Directive for a mathjax
app.directive('arbMathCompiler', function($timeout, $http, $compile, pageService, userService) {
	return {
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
		link: function(scope, element, attrs) {
			var mathjaxText = element.text();
			var cacheMathjaxDom = function() {
				pageService.mathjaxCache[mathjaxText] = element.html();
			};
			$timeout(function() {
				if (mathjaxText in pageService.mathjaxCache) {
					element.html(pageService.mathjaxCache[mathjaxText]);
				} else {
					MathJax.Hub.Queue(['Typeset', MathJax.Hub, element.get(0)]);
					MathJax.Hub.Queue(cacheMathjaxDom);
				}
			});
		},
	};
});

