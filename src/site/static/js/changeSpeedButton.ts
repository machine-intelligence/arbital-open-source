import app from './angular.ts';

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
				$scope.page.slowDownMap = {};
				$scope.page.speedUpMap = {};
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

			$scope.submitExplanationRequest = function(requestType, event) {
				arb.signupService.wrapInSignupFlow(requestType, function() {

					// Register the +1 to request
					arb.signupService.submitContentRequest(requestType || ($scope.goSlow ? 'slowDown' : 'speedUp'), $scope.page);

					// Submit feedback if there is any text
					if ($scope.request.freeformText.length > 0) {
						var text = $scope.goSlow ? 'Slower' : 'Faster';
						text += ' explanation request for page ' + $scope.page.pageId + ':\n' + $scope.request.freeformText;
						arb.stateService.postData(
							'/feedback/',
							{text: text}
						)
						$scope.request.freeformText = '';
						$scope.submittedFreeform = true;
					}
				});
			};

			$scope.hoverStart = function() {
				if (arb.isTouchDevice) {
					return;
				}

				$scope.timer = $timeout(function() {
					$scope.isHovered = true;
				}, 100);
			};

			$scope.hoverEnd = function() {
				$timeout.cancel($scope.timer);
			};
		},
		link: function(scope: any, element, attrs) {
			var parent = element.parent();
			var container = angular.element(element.find('.change-speed-container'));

			var positionButton = function() {
				$timeout(function() {
					// Stick to the top of the parent unless the top of the parent is off the page
					var topOfParent = parent[0].getBoundingClientRect().top;
					scope.stickToTop = topOfParent > 0;

					// Stick to the bottom of the parent if the bottom is near the top of the page
					var bottomOfParent = parent[0].getBoundingClientRect().bottom;
					scope.stickToBottom = bottomOfParent < 0;
					if (bottomOfParent < 0) {
						container.css('top', parent.offset().top + parent.height());
					} else {
						container.css('top', '');
					}
				});
			}

			angular.element($window).bind('scroll', function() {
				scope.haveScrolled = true;
				// positionButton();
			});

			positionButton();
		},
	}
});
