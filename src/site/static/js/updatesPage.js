'use strict';

// Directive for the Updates page.
app.directive('arbUpdates', function($compile, $location, $rootScope, pageService, userService) {
	return {
		templateUrl: 'static/html/updatesPage.html',
		scope: {
			updateGroups: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.updateGroups.sort(function(a, b) {
				if (b.key.unseen !== a.key.unseen) {
					return (b.key.unseen ? 1 : 0) - (a.key.unseen ? 1 : 0);
				}
				if (b.mostRecentDate === a.mostRecentDate) return 0;
				return b.mostRecentDate < a.mostRecentDate ? -1 : 1;
			});
		},
	};
});
