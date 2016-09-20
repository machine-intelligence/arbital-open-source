'use strict';

import app from './angular.ts';

// Directive for showing a standard Arbital page.
app.directive('arbPage', function($http, $location, $compile, $timeout, $interval, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/page.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.mastery = arb.masteryService.masteryMap[$scope.pageId];
			$scope.questionIds = $scope.page.questionIds || [];
			$scope.isTinyScreen = !$mdMedia('gt-sm');
			$scope.isSingleColumn = !$mdMedia('gt-md');
			$scope.isUser = !!arb.userService.userMap[$scope.pageId];
			$scope.selectedLens = undefined;
			$scope.page.authors = $scope.page.changeLogs
					.map(function(changeLog) {
						return changeLog.userId;
					})
					.reduce(function(accumulator, item) {
						if (!accumulator.includes(item)) {
							accumulator.push(item);
						}
						return accumulator;
					}, []);

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

			// Compute lenses
			$scope.lenses = $scope.page.lenses.slice();
			$scope.lenses.unshift({
				lensId: $scope.page.pageId,
				lensName: 'Main',
			});

			// Determine which lens is selected
			$scope.computeSelectedLensId = function() {
				var lensId = $location.search().l;
				// Check if lens is explicitly specified in the URL
				if (lensId) {
					// Check if this lens actually exists
					if ($scope.page.lenses.some(function(lens) { return lens.lensId == lensId; })) {
						return lensId;
					} else {
						$location.search('l', undefined);
					}
				}
				return $scope.page.pageId;
			};

			// Monitor URL to see if we need to switch lenses
			$scope.$watch(function() {
				return $location.absUrl();
			}, function() {
				// NOTE: this also gets called when the user clicks on a link to go to another page,
				// but in that case we don't want to do anything.
				// TODO: create a better workaround (we can broadcast an event)
				if ($location.path().indexOf($scope.pageId) >= 0 || $location.path().indexOf($scope.page.alias) >= 0) {
					$scope.tabSelect($scope.computeSelectedLensId());
				}
			});

			// Check if the given lens is loaded.
			$scope.isLoaded = function(lensId) {
				// Note that questions might have empty text.
				return lensId in arb.stateService.pageMap && (arb.stateService.pageMap[lensId].text.length > 0 || arb.stateService.pageMap[lensId].isQuestion());
			};

			// Called when there is a click inside the tabs
			$scope.tabsClicked = function($event, lensId) {
				// Check if there was a CTRL+click on a tab
				if ($event.ctrlKey || $event.metaKey) {
					window.open(arb.urlService.getPageUrl(lensId, {permalink: true}), '_blank');
				} else {
					$scope.tabSelect(lensId);
				}
			};

			// Show the panel to add tags
			$scope.isTagsPanelVisible = false;
			$scope.$on('showTagsPanel', function() {
				$scope.isTagsPanelVisible = true;
			});

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

			// Edit trust stuff
			$scope.updateUserTrust = function(domainId) {
				var data = {
					userId: $scope.page.pageId,
					domainId: domainId,
					editTrust: +$scope.page.trustMap[domainId].editTrust,
				};
				arb.stateService.postDataWithoutProcessing('/json/updateUserTrust/', data);
			};
		},
		link: function(scope: any, element, attrs) {
			// Manage switching between lenses, including loading the necessary data.
			var switchingLenses = false;
			var switchToLens = function(lensId) {
				if (scope.selectedLens && lensId === scope.selectedLens.pageId) { return; }
				if (switchingLenses) { return; }

				var $pageLensBody = $(element).find('.page-lens-body');
				scope.selectedLens = arb.stateService.pageMap[lensId];
				
				arb.analyticsService.reportPageIdView(scope.selectedLens.pageId);
				arb.stateService.setTitle(scope.selectedLens.title);
				
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
