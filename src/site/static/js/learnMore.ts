import app from './angular.ts';

// Directive to show the discussion section for a page
app.directive('arbLearnMore', function($compile, $location, $timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/learnMore.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			// Return true if there are any learn more suggestions to show
			$scope.hasLearnMore = function() {
				return Object.keys($scope.suggestedPageIds).length > 0;
			};

			// Map of suggested page id -> {
			//  teaches: pages it teaches
			//  reliesOn: pages it relies on,
			//  expandsOn: pages it expands on,
			// }
			$scope.suggestedPageIds = {};
			let addSuggestion = function(kind, subjectId, pageIds) {
				for (let pageId of pageIds) {
					if (!(pageId in $scope.suggestedPageIds)) {
						$scope.suggestedPageIds[pageId] = {
							teaches: '',
							reliesOn: '',
							expandsOn: '',
						};
					}
					if ($scope.suggestedPageIds[pageId][kind].length > 0) {
						$scope.suggestedPageIds[pageId][kind] += ', ';
					}
					$scope.suggestedPageIds[pageId][kind] += arb.stateService.pageMap[subjectId].title;
				}
			};
			for (let subjectId in $scope.page.learnMoreTaughtMap) {
				addSuggestion('teaches', subjectId, $scope.page.learnMoreTaughtMap[subjectId]);
			}
			for (let subjectId in $scope.page.learnMoreRequiredMap) {
				addSuggestion('reliesOn', subjectId, $scope.page.learnMoreRequiredMap[subjectId]);
			}
			for (let subjectId in $scope.page.learnMoreCoveredMap) {
				addSuggestion('expandsOn', subjectId, $scope.page.learnMoreCoveredMap[subjectId]);
			}
		},
		link: function(scope: any, element, attrs) {
			if (scope.hasLearnMore()) {
				arb.analyticsService.reportEventToHeapAndMixpanel('learn more shown');
			}
			$(element).on('click', '.intrasite-link', function(event) {
				arb.analyticsService.reportEventToHeapAndMixpanel('learn more link clicked');
			});
		}
	};
});
