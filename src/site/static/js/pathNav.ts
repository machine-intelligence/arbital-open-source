'use strict';

import app from './angular.ts';

// Directive for navigating the path
app.directive('arbPathNav', function($location, arb) {
	return {
		templateUrl: versionUrl('static/html/pathNav.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.path = arb.stateService.path;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.pathLength = 0;
			if (arb.pathService.isOnPath()) {
				$scope.pathLength = $scope.path.pages.length;
				if ($scope.path.pages[0].pageId == $scope.path.guideId) {
					// If the path starts with the path guide page, we don't count it
					$scope.pathLength--;
				}
			} else {
				$scope.pathLength = $scope.page.pathPages.length;
			}
			$scope.isHubPathActive = $location.search().pathPageId;

			$scope.getVisibleProgress = function() {
				if (!$scope.path) {
					if ($scope.page.pathPages[0].pathPageId != $scope.pageId) {
						// If the path starts with the path guide page, we don't count it
						return 0;
					}
					return 1;
				}
				if ($scope.path.pages[0].pageId == $scope.path.guideId) {
					// If the path starts with the path guide page, we don't count it
					return $scope.path.progress;
				}
				return $scope.path.progress + 1;
			};
		},
	};
});

