'use strict';

// Directive for showing a standard Arbital page.
app.directive('arbPage', function($http, $location, $compile, $timeout, $interval, $mdMedia, arb) {
	return {
		templateUrl: 'static/html/page.html',
		scope: {
			pageId: '@',
			isSimpleEmbed: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.mastery = arb.masteryService.masteryMap[$scope.pageId];
			$scope.questionIds = $scope.page.questionIds || [];
			$scope.isTinyScreen = !$mdMedia('gt-sm');
			$scope.isSingleColumn = !$mdMedia('gt-md');
			$scope.isUser = !!arb.userService.userMap[$scope.pageId];

			// Check if the user has all the requisites for the given lens
			$scope.hasAllReqs = function(lensId) {
				var reqs = arb.stateService.pageMap[lensId].requirementIds;
				for (var n = 0; n < reqs.length; n++) {
					if (!arb.masteryService.hasMastery(reqs[n])) {
						return false;
					}
				}
				return true;
			};

			// Sort lenses (from most technical to least)
			$scope.page.lensIds.sort(function(a, b) {
				return arb.stateService.pageMap[a].lensIndex - arb.stateService.pageMap[b].lensIndex;
			});
			$scope.page.lensIds.unshift($scope.page.pageId);

			// Determine which lens is selected
			$scope.computeSelectedLensId = function() {
				if ($location.search().l) {
					// Lens is explicitly specified in the URL
					return $location.search().l;
				} else if (arb.pathService.path && arb.pathService.path.onPath) {
					// The learning list specified this page specifically
					return $scope.page.pageId;
				}
				// Select the hardest lens for which the user has met all requirements
				var lastIndex = $scope.page.lensIds.length - 1;
				var selectedLensId = $scope.page.lensIds[lastIndex];
				for (var n = lastIndex - 1; n >= 0; n--) {
					var lensId = $scope.page.lensIds[n];
					if ($scope.hasAllReqs(lensId)) {
						selectedLensId = lensId;
					}
				}
				return selectedLensId;
			};

			// Monitor URL to see if we need to switch lenses
			$scope.$watch(function() {
				return $location.absUrl();
			}, function() {
				// NOTE: this also gets called when the user clicks on a link to go to another page,
				// but in that case we don't want to do anything.
				// TODO: create a better workaround
				if ($location.path().indexOf($scope.pageId) >= 0 || $location.path().indexOf($scope.page.alias) >= 0) {
					$scope.tabSelect($scope.computeSelectedLensId());
				}
			});

			// Check if the given lens is loaded.
			$scope.isLoaded = function(lensId) {
				// Note that questions might have empty text.
				return lensId in arb.stateService.pageMap && (arb.stateService.pageMap[lensId].text.length > 0 || arb.stateService.pageMap[lensId].isQuestion());
			};
console.log('bleh"')
			// Called when there is a click inside the tabs
			$scope.tabsClicked = function($event, lensId) {
				// Check if there was a CTRL+click on a tab
				if ($event.ctrlKey) {
					console.log(arb.urlService.getPageUrl(lensId));
					window.open(arb.urlService.getPageUrl(lensId, {permalink: true}), '_blank');
				} else {
					$scope.tabSelect(lensId);
				}
			};

			// Toggle between show delete answer buttons
			$scope.showDeleteAnswer = false;
			$scope.toggleDeleteAnswers = function() {
				$scope.showDeleteAnswer = !$scope.showDeleteAnswer;
			};

			// Submit this page to a domain (currently just math)
			$scope.submitToDomain = function() {
				var data = {
					pageId: $scope.pageId,
					domainId: '1lw',
				};
				arb.stateService.postData('/json/newPageToDomainSubmission/', data, function successFn(data) {
					var submission = data.result.submission;
					$scope.page.domainSubmissions[submission.domainId] = submission;
					arb.analyticsService.reportPageToDomainSubmission();
				});
			};
		},
		link: function(scope, element, attrs) {
			// Manage switching between lenses, including loading the necessary data.
			var switchingLenses = false;
			var switchToLens = function(lensId) {
				if (scope.selectedLens && lensId === scope.selectedLens.pageId) { return; }
				if (switchingLenses) { return; }

				var $pageLensBody = $(element).find('.page-lens-body');
				scope.selectedLens = arb.stateService.pageMap[lensId];
				$pageLensBody.animate({opacity: 0}, 400, 'swing', function() {
					switchingLenses = true;
					$timeout(function() {
						$pageLensBody.animate({opacity: 1}, 400, 'swing', function() {
							$pageLensBody.css('opacity', '');
						});
					});
					if (scope.selectedLens || lensId !== scope.pageId) {
						$location.search('l', lensId);
					}
					// A new lens became visible. Sometimes this happens when the user is going through
					// a path and clicks "Next" at the bottom of the page. In this case we need to
					// scroll upwards to have them start reading this lens
					if ($('body').scrollTop() > $pageLensBody.offset().top) {
						$('body').scrollTop($pageLensBody.offset().top - 100);
					}
					switchingLenses = false;
				});
			};
			scope.tabSelect = function(lensId) {
				if (scope.isLoaded(lensId)) {
					$timeout(function() {
						switchToLens(lensId);
					});
				} else {
					arb.pageService.loadLens(lensId);
					// Switch to the loaded lens. If we couldn't load the specified lens (e.g. if it doesn't exist),
					// then just go to the main lens.
					switchToLens(lensId in arb.stateService.pageMap ? lensId : scope.pageId);
				}
			};
			scope.tabSelect(scope.computeSelectedLensId());
		},
	};
});
