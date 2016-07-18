// arb-slow-down-button
app.directive('arbSlowDownButton', function(arb, $window, $timeout) {
	return {
		templateUrl: versionUrl('static/html/slowDown.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			arb.stateService.postData('/json/alternatePages/', {pageId: $scope.pageId},
				function(data) {
					$scope.altTeachers = data.result.alternateTeachers.map(function(altTeacherId) {
						return arb.stateService.getPage(altTeacherId);
					});
				});

			if (!$scope.page.slowPagePairs) {
				arb.stateService.postData('/json/slowDown/', {pageId: $scope.pageId});
			}

			// Return true if there is at least one page that's suggested
			$scope.hasSomeSuggestions = function() {
				var hasSlowDown = $scope.slowDownMap && Object.keys($scope.slowDownMap).length > 0;
				return $scope.page.requirements.length > 0 || hasSlowDown;
			};

			// Allow the user to request an easier explanation
			$scope.request = {
				freeformText: '',
			};
			$scope.submitExplanationRequest = function() {
				// Register the +1 to request
				var erData = {
					pageId: $scope.page.pageId,
					type: 'slowDown',
				};
				arb.stateService.postData('/json/explanationRequest/', erData);

				// Submit feedback if there is any text
				if ($scope.request.freeformText.length > 0) {
					arb.stateService.postData(
						'/feedback/',
						{text: 'Explanation request for page ' + $scope.page.pageId + ':\n' + $scope.request.freeformText}
					)
					$scope.request.freeformText = '';
				}
			};
		},
		link: function(scope, element, attrs) {
			var parent = element.parent();
			var slowDownContainer = angular.element(element.find('.slow-down-container'));

			var topOfParent = parent[0].getBoundingClientRect().top + 10;
			slowDownContainer.css('top', topOfParent);

			angular.element($window).bind('scroll', function() {
				scope.haveScrolled = true;

				// Make the button not go past the bottom of the parent
				var bottomOfParent = parent[0].getBoundingClientRect().bottom + 20;
				slowDownContainer.css('top', Math.min(bottomOfParent, topOfParent));
			});
		},
	}
});
