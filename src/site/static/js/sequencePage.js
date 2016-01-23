"use strict";

// Directive for the sequence page.
app.directive("arbSequencePage", function($timeout, $http, pageService, userService) {
	return {
		templateUrl: "static/html/sequencePage.html",
		scope: {
			sequence: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});

// Directive for a recursive part of a sequence.
app.directive("arbSequencePart", function(pageService, userService, RecursionHelper) {
	return {
		templateUrl: "static/html/sequencePart.html",
		scope: {
			part: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
		compile: function(element) {
			return RecursionHelper.compile(element);
		}
	};
});

