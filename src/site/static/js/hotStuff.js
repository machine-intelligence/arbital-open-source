'use strict';

// arb-hot-stuff directive displays a list of hot pages, recommended for reading
app.directive('arbHotStuff', function() {
	return {
		templateUrl: 'static/html/hotStuff.html',
		controller: function($scope) {
			$scope.pageIds = ['11x', '1y'];
		},
	};
});
