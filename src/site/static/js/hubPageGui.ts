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
			$scope.levels = [1, 2, 3, 4];
			$scope.level = '0';
			if (arb.masteryService.masteryMap[$scope.pageId].has) {
				$scope.level = '' + arb.masteryService.masteryMap[$scope.pageId].level;
			}
			$scope.getIntLevel = function() {
				return +$scope.level;
			};
			$scope.getLevelName = function(level) {
				switch (+level) {
					case 1:
						return 'LooseUnderstanding';
					case 2:
						return 'BasicUnderstanding';
					case 3:
						return 'TechnicalUnderstanding';
					case 4:
						return 'ResearchLevelUnderstanding';
					default:
						return 'NoUnderstanding';
				}
			};
			$scope.getLevelTitle = function(level) {
				switch (+level) {
					case 1:
						return 'Loose (you are familiar with terminology)';
					case 2:
						return 'Basic (you are familiar with the basic formula)';
					case 3:
						return 'Technical (you can do tricks with the formula)';
					case 4:
						return 'Research level (you do this for a living)';
				}
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

			// Called when user clicks on "quick learn" button
			$scope.goLearn = function(event) {
				var pageId = $scope.page.getBestLearnPageId($scope.getIntLevel());
				var url = arb.urlService.getHubSuggestionPageUrl(pageId, {hubId: $scope.pageId});
				arb.urlService.goToUrl(url, {event: event});
			};

			// Called when user clicks on "quick boost" button
			$scope.goBoost = function(event) {
				var pageId = $scope.page.getBestBoostPageId($scope.getIntLevel());
				var url = arb.urlService.getHubSuggestionPageUrl(pageId, {hubId: $scope.pageId});
				arb.urlService.goToUrl(url, {event: event});
			};
		},
	};
});
