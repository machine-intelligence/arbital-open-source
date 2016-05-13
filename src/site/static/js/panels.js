// panels.js contains directive for panels
'use strict';

// arb-list-panel directive displays a list of things in a panel
app.directive('arbListPanel', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/listPanel.html',
		transclude: true,
		scope: {
			title: '@',
			moreLink: '@',
			items: '=',
			numToDisplay: '='
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
		},
	};
});

// arb-read-mode-panel directive displays a list of things to read in a panel
app.directive('arbReadModePanel', function($http, userService, pageService) {
	return {
		templateUrl: 'static/html/readModePanel.html',
		scope: {
			numToDisplay: '='
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
