"use strict";

// Directive to show a lens' content
app.directive("arbLens", function($compile, $location, $timeout, $interval, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/lens.html",
		scope: {
			pageId: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];

			$scope.mastery = pageService.masteryMap[$scope.pageId];
			if (!$scope.mastery) {
				$scope.mastery = {has: false};
			}

			// Process mastery events.
			$scope.toggleMastery = function() {
				pageService.updateMastery($scope, $scope.page.pageId, !$scope.mastery.has);
				$scope.mastery = pageService.masteryMap[$scope.pageId];
			};
		},
		link: function(scope, element, attrs) {
			// Process all embedded votes.
			$timeout(function() {
				element.find("[embed-vote-id]").each(function(index) {
					var $link = $(this);
					var pageAlias = $link.attr("embed-vote-id");
					pageService.loadIntrasitePopover(pageAlias, {
						success: function(data, status) {
							var pageId = pageService.pageMap[pageAlias].pageId;
							var divId = "embed-vote-" + pageId;
							var $embedDiv = $compile("<div id='" + divId + "' class='embedded-vote'><arb-vote-bar page-id='" + pageId + "'></arb-vote-bar></div>")(scope);
							$link.replaceWith($embedDiv);
						},
						error: function(data, status) {
							console.error("Couldn't load embedded votes: " + pageAlias);
						}
					});
				});
			});
		},
	};
});

