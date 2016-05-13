'use strict';

// arb-read-mode-page hosts the arb-read-mode-panel
app.directive('arbReadModePage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/readModePage.html',
		controller: function($scope) {
			$http({method: 'POST', url: '/json/readMode/', data: JSON.stringify({})})
				.success(function(data) {
					pageService.processServerData(data);
					userService.processServerData(data);
					$scope.hotPageIds = data.result.hotPageIds;
				});
		},
	};
});
