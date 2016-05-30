'use strict';

// Directive for the User page.
app.directive('arbUserPage', function(arb) {
	return {
		templateUrl: 'static/html/userPage.html',
		scope: {
			userId: '@',
			userPageData: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
