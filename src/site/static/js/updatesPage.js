'use strict';

// Directive for the Updates page.
app.directive('arbUpdates', function($http, arb) {
	return {
		templateUrl: 'static/html/updatesPage.html',
		scope: {
			updateGroups: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.updateGroups.sort(function(a, b) {
				if (b.key.seen !== a.key.seen) {
					return (b.key.seen ? 0 : 1) - (a.key.seen ? 0 : 1);
				}
				if (b.mostRecentDate === a.mostRecentDate) return 0;
				return b.mostRecentDate < a.mostRecentDate ? -1 : 1;
			});

			$scope.dismissUpdate = function(update, updateGroup, index) {
				$http({method: 'POST', url: '/dismissUpdate/', data: JSON.stringify({
					id: update.id
				})});

				// Remove this update from the list
				updateGroup.splice(index, 1);
			};
		},
	};
});
