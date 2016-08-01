'use strict';

import app from './angular.ts';

// Footer that appears at the bottom of the page if the user is exploring from a HUB page
app.directive('arbHubPageFooter', function($location, $compile, $timeout, arb) {
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

			$scope.pathPage = undefined;
			if ($location.search().pathPageId) {
				$scope.pathPage = arb.stateService.pageMap[$location.search().pathPageId];
				$scope.currentPathIndex = 0;
				for (let n = 0; n < $scope.pathPage.pathPages.length; n++) {
					if ($scope.pathPage.pathPages[n].pathPageId == $scope.pageId) {
						$scope.currentPathIndex = n;
					}
				}
				if ($scope.currentPathIndex > 0) {
					let prevPageId = $scope.pathPage.pathPages[$scope.currentPathIndex - 1].pathPageId;
					$scope.prevPageUrl = arb.urlService.getPageUrl(prevPageId, {
						pathPageId: $scope.pathPage.pageId,
						hubId: $scope.hubPageId,
					});
				}
				if ($scope.currentPathIndex < $scope.pathPage.pathPages.length - 1) {
					let nextPageId = $scope.pathPage.pathPages[$scope.currentPathIndex + 1].pathPageId;
					$scope.nextPageUrl = arb.urlService.getPageUrl(nextPageId, {
						pathPageId: $scope.pathPage.pageId,
						hubId: $scope.hubPageId,
					});
				}
			}

			$scope.isOnLastPathPage = function() {
				if (!$scope.pathPage) return false;
			};

			// Compute which masteries will be leveled up as the result of this page,
			$scope.nextMasteryLevels = {}; // masteryId -> to which level this page would set it
			$scope.levelUpMasteries = {}; // masteryId -> whether to perform the level up
			// For all page's masteries, if they will level up the current user, add them to maps
			var extractMasteries = function(page) {
				page.subjects.forEach(function(subject) {
					if (!subject.isStrong) return;
					var currentLevel = 0;
					if (!(subject.parentId in arb.masteryService.masteryMap)) {
						currentLevel = arb.masteryService.masteryMap[subject.parentId].level;
					}
					if (currentLevel > subject.level) return;
					$scope.nextMasteryLevels[subject.parentId] = subject.level;
					$scope.levelUpMasteries[subject.parentId] = true;
				});
			};
			extractMasteries($scope.page);
			// If this is the last page of an arc, count arc page's masteries as well
			if ($scope.pathPage && !$scope.nextPageUrl) {
				extractMasteries($scope.pathPage);
			}
			$scope.levelUpMasteryCount = Object.keys($scope.levelUpMasteries).length;

			var getNextHubMasteryLevel = function() {
				var nextHubMasteryLevel = arb.masteryService.masteryMap[$scope.hubPageId].level;
				if ($scope.levelUpMasteries[$scope.hubPageId]) {
					nextHubMasteryLevel = Math.max(nextHubMasteryLevel, $scope.nextMasteryLevels[$scope.hubPageId]);
				}
				return nextHubMasteryLevel;
			};
			// Compute url for "learn" button
			$scope.getLearnUrl = function() : string {
				var level = getNextHubMasteryLevel();
				var pageId = $scope.hubPage.getBestLearnPageId(level, $scope.pageId);
				if (!pageId) return '';
				return arb.urlService.getHubSuggestionPageUrl(pageId, {hubId: $scope.hubPageId});
			};
			// Compute url for "boost" button
			$scope.getBoostUrl = function() : string {
				var level = getNextHubMasteryLevel();
				if (!arb.masteryService.hasUnreadBoostPages($scope.hubPage, level, $scope.pageId)) return undefined;
				var pageId = $scope.hubPage.getBestBoostPageId(level, $scope.pageId);
				if (!pageId) return '';
				return arb.urlService.getHubSuggestionPageUrl(pageId, {hubId: $scope.hubPageId});
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
