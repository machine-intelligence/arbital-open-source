import app from './angular.ts';

declare var heap: any;

// arb-change-speed-button
app.directive('arbChangeSpeedButton', function(arb, $window, $timeout, analyticsService) {
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
			$scope.pageRequirements = $scope.page.requirements.filter(function(requirement) {
				// Don't show if the page requires itself
				if (requirement.parentId == $scope.pageId) return false;
				return !arb.stateService.pageMap[requirement.parentId].indirectTeacher;
			});

			$scope.pageSubjectsExceptItself = $scope.page.subjects.filter(function(subject) {
				return subject.parentId != $scope.pageId;
			});

			// Fetch data
			if (!$scope.page.slowDownMap || !$scope.page.speedUpMap) {
				$scope.page.slowDownMap = {};
				$scope.page.speedUpMap = {};
				arb.stateService.postData('/json/changeSpeed/', {pageId: $scope.pageId}, function() {

					// These maps contain objects: suggested page id -> list of page ids that teach similar subjects
					$scope.page.slowerLevelMap = {};
					$scope.page.lowerLevelMap = {};
					$scope.page.fasterLevelMap = {};
					$scope.page.higherLevelMap = {};

					let currentPageSpeed = arb.pageService.getPageSpeed($scope.page.tagIds);

					for (let i = 0; i < $scope.page.subjects.length; ++i) {
						let subjectPair = $scope.page.subjects[i];

						// Find pages that teach the same subject at the same level, but slower, or at a lower level.
						let slowerOrLowerPages = $scope.page.slowDownMap[subjectPair.parentId];
						if (slowerOrLowerPages) {
							for (let j = 0; j < slowerOrLowerPages.length; ++j) {
								let otherPair = $scope.page.slowDownMap[subjectPair.parentId][j];

								if (otherPair.level == subjectPair.level) {
									let otherTeacher = arb.stateService.pageMap[otherPair.childId];
									let otherTeacherSpeed = arb.pageService.getPageSpeed(otherTeacher.tagIds);

									if (otherTeacherSpeed < currentPageSpeed) {
										if (!(otherPair.childId in $scope.page.slowerLevelMap)) {
											$scope.page.slowerLevelMap[otherPair.childId] = [];
										} 
										$scope.page.slowerLevelMap[otherPair.childId].push(subjectPair.parentId);
									}
								} else if (otherPair.level < subjectPair.level) {
									if (!(otherPair.childId in $scope.page.lowerLevelMap)) {
										$scope.page.lowerLevelMap[otherPair.childId] = [];
									} 
									$scope.page.lowerLevelMap[otherPair.childId].push(subjectPair.parentId);
								}
							}
						}

						// Find pages that teach the same subject at the same level, but faster, or at a higher level.
						let fasterOrHigherPages = $scope.page.speedUpMap[subjectPair.parentId];
						if (fasterOrHigherPages) {
							for (let j = 0; j < fasterOrHigherPages.length; ++j) {
								let otherPair = $scope.page.speedUpMap[subjectPair.parentId][j];

								if (otherPair.level == subjectPair.level) {
									let otherTeacher = arb.stateService.pageMap[otherPair.childId];
									let otherTeacherSpeed = arb.pageService.getPageSpeed(otherTeacher.tagIds);

									if (otherTeacherSpeed > currentPageSpeed) {
										if (!(otherPair.childId in $scope.page.fasterLevelMap)) {
											$scope.page.fasterLevelMap[otherPair.childId] = [];
										} 
										$scope.page.fasterLevelMap[otherPair.childId].push(subjectPair.parentId);
									}
								} else if (otherPair.level > subjectPair.level) {
									if (!(otherPair.childId in $scope.page.higherLevelMap)) {
										$scope.page.higherLevelMap[otherPair.childId] = [];
									} 
									$scope.page.higherLevelMap[otherPair.childId].push(subjectPair.parentId);
								}
							}
						}
					}

				});
			}

			$scope.hasLowerLevelMap = function() {
				return $scope.page.lowerLevelMap && Object.keys($scope.page.lowerLevelMap).length > 0;
			};

			$scope.hasHigherLevelMap = function() {
				return $scope.page.higherLevelMap && Object.keys($scope.page.higherLevelMap).length > 0;
			};

			$scope.hasSlowDownMap = function() {
				return $scope.page.slowerLevelMap && Object.keys($scope.page.slowerLevelMap).length > 0;
			};

			$scope.hasSpeedUpMap = function() {
				return $scope.page.fasterLevelMap && Object.keys($scope.page.fasterLevelMap).length > 0;
			};

			// Return true if there is at least one page that's suggested
			$scope.hasSomeSuggestions = function() {
				if ($scope.goSlow) {
					if ($scope.page.arcPageIds && $scope.page.arcPageIds.length > 0 && !arb.pathService.isOnPath()) return true;
					return $scope.pageRequirements.length > 0 || $scope.hasSlowDownMap();
				}
				return $scope.hasSpeedUpMap();
			};

			// Allow the user to request an easier explanation
			$scope.request = {
			};

			$scope.submitExplanationRequest = function(requestType, event) {
				// Register the +1 to request
				arb.signupService.submitContentRequest(requestType, $scope.page);

				analyticsService.reportEventToHeapAndMixpanel('lateral nav: submit a +1 request', {
					type: $scope.goSlow ? 'say-what' : 'go-faster',
					wasBlue: $scope.hasSomeSuggestions(),
					pageId: $scope.pageId,
					requestType: requestType,
				});
			};

			$scope.hoverStart = function() {
				if (arb.isTouchDevice) {
					return;
				}

				$scope.timer = $timeout(function() {
					analyticsService.reportEventToHeapAndMixpanel('lateral nav: hover', {
						type: $scope.goSlow ? 'say-what' : 'go-faster',
						wasBlue: $scope.hasSomeSuggestions(),
						pageId: $scope.pageId,
					});
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
				//positionButton();
			});

			positionButton();
		},
	}
});
