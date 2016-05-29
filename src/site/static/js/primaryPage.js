'use strict';

// Directive for the entire primary page.
app.directive('arbPrimaryPage', function($compile, $location, $timeout, arb, autocompleteService) {
	return {
		templateUrl: 'static/html/primaryPage.html',
		scope: {
			noFooter: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			
			$scope.page = pageService.primaryPage;
			$scope.page.childIds.sort(pageService.getChildSortFunc($scope.page.sortChildrenBy));
			$scope.page.relatedIds.sort(pageService.getChildSortFunc('likes'));
		},
	};
});
