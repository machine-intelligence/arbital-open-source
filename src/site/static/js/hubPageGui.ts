'use strict';

import app from './angular.ts';

// Directive to show the GUI at the top of a HUB page
app.directive('arbHubPageGui', function($compile, $timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/hubPageGui.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			arb.masteryService.sortHubContent($scope.page);

			// Track current user's the mastery level for this requisite
			$scope.level = '0';
			if (arb.masteryService.masteryMap[$scope.pageId].has) {
				$scope.level = '' + arb.masteryService.masteryMap[$scope.pageId].level;
			}
			$scope.getIntLevel = function() {
				return +$scope.level;
			};

			// Update user's mastery level
			$scope.updateLevel = function() {
				var params = {
					masteryLevels: {
					},
				};
				params.masteryLevels[$scope.pageId] = $scope.getIntLevel();
				arb.stateService.postData('/updateMasteries/', params);
				arb.masteryService.masteryMap[$scope.pageId].level = $scope.getIntLevel();
			};

			$scope.goLearn = function(event) {
				arb.urlService.goToUrl($scope.page.getBestLearnPageUrl($scope.getIntLevel()), {event: event});
			};

			$scope.goBoost = function(event) {
				arb.urlService.goToUrl($scope.page.getBestBoostPageUrl($scope.getIntLevel()), {event: event});
			};
		},
	};
});
