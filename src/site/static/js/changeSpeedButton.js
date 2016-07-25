// arb-change-speed-button
app.directive('arbChangeSpeedButton', function(arb, $window, $timeout) {
	return {
		templateUrl: versionUrl('static/html/changeSpeedButton.html'),
		scope: {
			pageId: '@',
			// If true, this is a 'slow down' button, otherwise 'speed up'
			goSlow: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			// Fetch data
			if (!$scope.page.slowDownMap || !$scope.page.speedUpMap) {
 				arb.stateService.postData('/json/changeSpeed/', {pageId: $scope.pageId});
 			}

			// Return true if there is at least one page that's suggested
			$scope.hasSomeSuggestions = function() {
				if ($scope.goSlow) {
					var hasMap = $scope.slowDownMap && Object.keys($scope.slowDownMap).length > 0;
					return $scope.page.requirements.length > 0 || hasMap;
				}
				var hasMap = $scope.speedUpMap && Object.keys($scope.speedUpMap).length > 0;
				return $scope.page.subjects.length > 0 || hasMap;
			};

			// Allow the user to request an easier explanation
			$scope.request = {
				freeformText: '',
			};
			$scope.submitExplanationRequest = function() {
				// Register the +1 to request
				var erData = {
					pageId: $scope.page.pageId,
					type: $scope.goSlow ? 'slowDown' : 'speedUp',
				};
				arb.stateService.postData('/json/explanationRequest/', erData);

				// Submit feedback if there is any text
				if ($scope.request.freeformText.length > 0) {
					var text = $scope.goSlow ? 'Slower' : 'Faster';
					text += ' explanation request for page ' + $scope.page.pageId + ':\n' + $scope.request.freeformText;
					arb.stateService.postData(
						'/feedback/',
						{text: text}
					)
					$scope.request.freeformText = '';
				}
			};
		},
		link: function(scope, element, attrs) {
			var parent = element.parent();
			var container = angular.element(element.find('.change-speed-container'));
			var topOfParent = parent[0].getBoundingClientRect().top + 10;
			container.css('top', topOfParent);

			angular.element($window).bind('scroll', function() {
				scope.haveScrolled = true;

				// Make the button not go past the bottom of the parent
				var bottomOfParent = parent[0].getBoundingClientRect().bottom + 20;
				container.css('top', Math.min(bottomOfParent, topOfParent));
			});
		},
	}
});
