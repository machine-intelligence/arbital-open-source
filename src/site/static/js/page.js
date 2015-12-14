"use strict";

// Directive for showing a standard Arbital page.
app.directive("arbPage", function ($location, $compile, $timeout, $interval, $mdMedia, pageService, userService) {
	return {
		templateUrl: "/static/html/page.html",
		scope: {
			pageId: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.mastery = pageService.masteryMap[$scope.pageId];
			$scope.questionIds = $scope.page.questionIds || [];
			$scope.isTinyScreen = !$mdMedia("gt-sm");
			$scope.isSingleColumn = !$mdMedia("gt-md");

			// Add the primary page as the first lens.
			$scope.page.lensIds.unshift($scope.page.pageId);

			// Determine which lens is selected
			$scope.selectedLens = $scope.page;
			if ($location.search().lens) {
				$scope.selectedLens = pageService.pageMap[$location.search().lens];
			}
			$scope.selectedLensIndex = $scope.page.lensIds.indexOf($scope.selectedLens.pageId);
			$scope.originalLensId = $scope.selectedLens.pageId;
			$scope.getPageTitle = function() {
				var pageTitle = $scope.page.title;
				if ($scope.selectedLensIndex <= 0) {
					return pageTitle;
				}
				return pageTitle + ": " + $scope.selectedLens.title;
			}

			// Manage switching between lenses, including loading the necessary data.
			var switchToLens = function(lensId) {
				if (lensId === $scope.page.pageId) {
					$location.search("lens", undefined);
				} else {
					$location.search("lens", lensId);
				}
				$scope.selectedLens = pageService.pageMap[lensId];
				$scope.$broadcast("lensTabChanged", lensId);
			};
			$scope.tabSelect = function(lensId) {
				if ($scope.isLoaded(lensId)) {
					$timeout(function() {
						switchToLens(lensId);
					});
				} else {
					pageService.loadLens(lensId, {
						success: function(data, status) {
							switchToLens(lensId);
						},
					});
				}
			};
			$scope.isLoaded = function(lensId) {
				return pageService.pageMap[lensId].text.length > 0;
			};

			// Called when there is a click inside the tabs
			var currentSelectedLensIndex = -1;
			$scope.tabsClicked = function($event) {
				// Check if there was a CTRL+click on a tab
				if (!$event.ctrlKey) return;
				var $target = $(event.target);
				var $tab = $target.closest("md-tab-item");
				if ($tab.length != 1) return;
				var tabIndex = $tab.index();
				var lensId = $scope.page.lensIds[tabIndex];
				window.open(pageService.getPageUrl(lensId), "_blank");
			};

			// Check if this comment is selected via URL hash
			$scope.isSelected = function() {
				return $location.hash() === "subpage-" + $scope.page.pageId;
			};

			// Check if the user has all the masteries for the given lens
			$scope.hasMastery = function(lensId) {
				var reqs = pageService.pageMap[lensId].requirementIds;
				for (var n = 0; n < reqs.length; n++) {
					if (!pageService.masteryMap[reqs[n]].has) {
						return false;
					}
				}
				return true;
			};
		},
	};
});
