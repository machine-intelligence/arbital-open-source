'use strict';

import app from './angular.ts';

// arb-explore-page directive displays a set of featured domains
app.directive('arbExplorePage', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/explorePage.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
	};
});
