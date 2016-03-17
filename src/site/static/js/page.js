"use strict";

// Directive for showing a standard Arbital page.
app.directive("arbPage", function ($location, $compile, $timeout, $interval, $mdMedia, pageService, userService) {
	return {
		templateUrl: "static/html/page.html",
		scope: {
			pageId: "@",
			isSimpleEmbed: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.mastery = pageService.masteryMap[$scope.pageId];
			$scope.questionIds = $scope.page.questionIds || [];
			$scope.isTinyScreen = !$mdMedia("gt-sm");
			$scope.isSingleColumn = !$mdMedia("gt-md");

			// Check if the user has all the requisites for the given lens
			$scope.hasAllReqs = function(lensId) {
				var reqs = pageService.pageMap[lensId].requirementIds;
				for (var n = 0; n < reqs.length; n++) {
					if (!pageService.hasMastery(reqs[n])) {
						return false;
					}
				}
				return true;
			};

			// Sort lenses (from most technical to least)
			$scope.page.lensIds.sort(function(a, b) {
				return pageService.pageMap[a].lensIndex - pageService.pageMap[b].lensIndex;
			});
			$scope.page.lensIds.unshift($scope.page.pageId);

			// Determine which lens is selected
			var computeSelectedLens = function() {
				if ($location.search().l) {
					// Lens is explicitly specified in the URL
					$scope.selectedLens = pageService.pageMap[$location.search().l];
				} else if (pageService.path && pageService.path.onPath) {
					// The learning list specified this page specifically
					$scope.selectedLens = pageService.pageMap[$scope.page.pageId];
				} else {
					// Select the hardest lens for which the user has met all requirements
					var lastIndex = $scope.page.lensIds.length - 1;
					$scope.selectedLens = pageService.pageMap[$scope.page.lensIds[lastIndex]];
					for (var n = lastIndex - 1; n >= 0; n--) {
						var lensId = $scope.page.lensIds[n];
						if ($scope.hasAllReqs(lensId)) {
							$scope.selectedLens = pageService.pageMap[lensId];
						}
					}
				}
			};
			computeSelectedLens();

			// Monitor URL to see if we need to switch lenses
			$scope.$watch(function() {
				return $location.absUrl();
			}, function() {
				// NOTE: this also gets called when the user clicks on a link to go to another page,
				// but in that case we don't want to do anything.
				// TODO: create a better workaround
				if ($location.path().indexOf($scope.pageId) >= 0 || $location.path().indexOf($scope.page.alias) >= 0) {
					computeSelectedLens();
				}
			});

			// Check if the given lens is loaded.
			$scope.isLoaded = function(lensId) {
				return pageService.pageMap[lensId].text.length > 0;
			};

			// Called when there is a click inside the tabs
			$scope.tabsClicked = function($event, lensId) {
				// Check if there was a CTRL+click on a tab
				if ($event.ctrlKey) {
					console.log(pageService.getPageUrl(lensId));
					window.open(pageService.getPageUrl(lensId, {permalink: true}), "_blank");
				} else {
					$scope.tabSelect(lensId);
				}
			};

			// Check if this page is selected via URL hash
			$scope.isSelected = function() {
				return $location.hash() === "subpage-" + $scope.page.pageId;
			};
		},
		link: function(scope, element, attrs) {
			// Manage switching between lenses, including loading the necessary data.
			var switchToLens = function(lensId) {
				if (lensId === scope.selectedLens.pageId) return;

				var $pageLensBody = $(element).find(".page-lens-body");
				$pageLensBody.animate({opacity:0}, 400, "swing", function() {
					$timeout(function() {
						$pageLensBody.animate({opacity:1}, 400, "swing", function() {
							$pageLensBody.css("opacity", "");
						});
					});
					scope.selectedLens = pageService.pageMap[lensId];
					$location.search("l", lensId);
					scope.$broadcast("lensTabChanged", lensId);
				});
			};
			scope.tabSelect = function(lensId) {
				if (scope.isLoaded(lensId)) {
					$timeout(function() {
						switchToLens(lensId);
					});
				} else {
					pageService.loadLens(lensId);
					switchToLens(lensId);
				}
			};
			scope.tabSelect(scope.selectedLens.pageId);
		},
	};
});
