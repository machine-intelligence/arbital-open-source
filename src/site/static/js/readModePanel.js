'use strict';

// arb-read-mode-panel displays a list of hot pages, recommended for reading
app.directive('arbReadModePanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/readModePanel.html',
		scope: {
			numPagesToShow: '=',
			includeSeeMoreLink: '=',
		},
		controller: function($scope) {
			$http({method: 'POST', url: '/json/readMode/', data: JSON.stringify({numPagesToLoad: $scope.numPagesToShow})})
				.success(function(data) {
					userService.processServerData(data);
					pageService.processServerData(data);
					$scope.hotPageIds = data.result.hotPageIds;
				})
				.error(function(data) {
					$scope.addMessage('hotPageIds', 'Error loading hot page ids: ' + data, 'error');
				});
		},
	};
});
