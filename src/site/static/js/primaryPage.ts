'use strict';

import app from './angular.ts';

// Directive for the entire primary page.
app.directive('arbPrimaryPage', function($compile, $location, $timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/primaryPage.html'),
		scope: {
			noFooter: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.primaryPage;
			$scope.page.childIds.sort(arb.pageService.getChildSortFunc($scope.page.sortChildrenBy));
			$scope.page.relatedIds.sort(arb.pageService.getChildSortFunc('likes'));
			$scope.page.explanations.sort(function(a,b) {
				return arb.stateService.pageMap[b.childId].likeScore() - arb.stateService.pageMap[a.childId].likeScore();
			});

			// Explanation request
			$scope.explanationRequest = {
				speed: '0',
				level: 2,
			};
			$scope.speedOptions = {
				'0': 'normal speed',
				'1': 'fast speed',
				'-1': 'low speed',
			};
			$scope.levelOptions = {
				1: arb.stateService.getLevelName(1) + ' level',
				2: arb.stateService.getLevelName(2) + ' level',
				3: arb.stateService.getLevelName(3) + ' level',
				4: arb.stateService.getLevelName(4) + ' level',
			};
			$scope.getRequestName = function(level) {
				switch (+level) {
					case 0:
						return 'NoUnderstanding';
					case 1:
						return 'LooseUnderstanding';
					case 2:
						return 'BasicUnderstanding';
					case 3:
						return 'TechnicalUnderstanding';
					case 4:
						return 'ResearchLevelUnderstanding';
				}
			};
			$scope.requestSubmitted = false;
			$scope.submitRequest = function() {
				var requestKey = 'teach' + $scope.getRequestName($scope.explanationRequest.level + 1);
				arb.signupService.submitContentRequest(requestKey, $scope.page);
				$scope.requestSubmitted = true;
			};
		},
		link: function(scope: any, element, attrs) {
			if (scope.page.domainIds.indexOf('1lw') >= 0) {
				element.addClass('math-background');
			}
		},
	};
});
