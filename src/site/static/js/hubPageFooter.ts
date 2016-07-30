'use strict';

import app from './angular.ts';

// Directive to show the footer for HUB pages
app.directive('arbHubPageFooter', function($compile, $timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/hubPageFooter.html'),
		scope: {
			pageId: '@',
			hubPageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.hubPage = arb.stateService.pageMap[$scope.hubPageId];
			arb.masteryService.sortHubContent($scope.hubPage);

			// Compute which masteries will be leveled up as the result of this page,
			$scope.nextMasteryLevels = {}; // masteryId -> to which level this page would set it
			$scope.levelUpMasteries = {}; // masteryId -> whether to perform the level up
			$scope.page.subjects.forEach(function(subject) {
				if (!subject.isStrong) return;
				var currentLevel = -1;
				if (!(subject.parentId in arb.masteryService.masteryMap)) {
					currentLevel = arb.masteryService.masteryMap[subject.parentId].level;
				}
				if (currentLevel > subject.level) return;
				$scope.nextMasteryLevels[subject.parentId] = subject.level;
				$scope.levelUpMasteries[subject.parentId] = true;
			});
			$scope.levelUpMasteryCount = Object.keys($scope.levelUpMasteries).length;

			// Compute URLs for the HUB buttons
			var getNextHubMasteryLevel = function() {
				var nextHubMasteryLevel = arb.masteryService.masteryMap[$scope.hubPageId].level;
				if ($scope.levelUpMasteries[$scope.hubPageId]) {
					nextHubMasteryLevel = Math.max(nextHubMasteryLevel, $scope.nextMasteryLevels[$scope.hubPageId]);
				}
				return nextHubMasteryLevel;
			};
			$scope.getLearnUrl = function() {
				var level = getNextHubMasteryLevel();
				return $scope.hubPage.getBestLearnPageUrl(level, $scope.pageId);
			};
			$scope.getBoostUrl = function() {
				var level = getNextHubMasteryLevel();
				if (!arb.masteryService.hasUnreadBoostPages($scope.hubPage, level, $scope.pageId)) return undefined;
				return $scope.hubPage.getBestBoostPageUrl(level, $scope.pageId);
			};

			// Called to level up user's masteries
			$scope.doLevelUp = function() {
				var masteryLevels = {};
				for (var masteryId in $scope.levelUpMasteries) {
					if (!$scope.levelUpMasteries[masteryId]) return;
					masteryLevels[masteryId] = $scope.nextMasteryLevels[masteryId];
				}
				if (Object.keys(masteryLevels).length <= 0) return;
				var params = {
					masteryLevels: masteryLevels,
				};
				arb.stateService.postData('/updateMasteries/', params);
			};
		},
	};
});
