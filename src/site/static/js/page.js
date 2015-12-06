"use strict";

// Directive for showing a standard Arbital page.
app.directive("arbPage", function ($location, $compile, $timeout, $interval, pageService, userService) {
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

			// Add the primary page as the first lens.
			$scope.page.lensIds.unshift($scope.page.pageId);

			// Determine which lens is selected
			$scope.selectedLens = $scope.page;
			if ($location.search().lens) {
				$scope.selectedLens = pageService.pageMap[$location.search().lens];
			}
			$scope.selectedLensIndex = $scope.page.lensIds.indexOf($scope.selectedLens.pageId);
			$scope.originalLensId = $scope.selectedLens.pageId;

			// Manage switching between lenses, including loading the necessary data.
			var switchToLens = function(lensId) {
				if (lensId === $scope.page.pageId) {
					$location.search("lens", undefined);
				} else {
					$location.search("lens", lensId);
				}
				$scope.selectedLens = pageService.pageMap[lensId];
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

			// Process click on showing the page diff button.
			$scope.showingDiff = false;
			$scope.toggleDiff = function() {
				$scope.showingDiff = !$scope.showingDiff;
				$scope.$broadcast("toggleDiff", $scope.page.pageId);
			};
		},
	};
});
