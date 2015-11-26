"use strict";

// Directive to show a lens' content
app.directive("arbLens", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
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
			window.setTimeout(function() {
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

// Directive for showing a the panel with requirements.
app.directive("arbRequirementsPanel", function($q, $timeout, $http, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/requirementsPanel.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			scope.inEditMode = false;

			// Check if the user has the given mastery.
			scope.hasMastery = function(requirementId) {
				return pageService.masteryMap[requirementId].has;
			}

			// Compute if we should show the panel
			scope.showPanel = false;
			for (var n = 0; n < scope.page.requirementIds.length; n++) {
				scope.showPanel |= !scope.hasMastery(scope.page.requirementIds[n]);
			}

			// Sort requirements
			scope.page.requirementIds.sort(function(a, b) {
				return (scope.hasMastery(a) ? 1 : 0) - (scope.hasMastery(b) ? 1 : 0);
			});

			// Toggle edit mode.
			scope.inEditModeToggle = function() {
				scope.inEditMode = !scope.inEditMode;
			};

			// Set up search
			scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				autocompleteService.parentsSource({term: text}, function(results) {
					deferred.resolve(results);
				});
        return deferred.promise;
			};
			scope.searchResultSelected = function(result) {
				if (!result) return;
				var data = {
					parentId: result.label,
					childId: scope.page.pageId,
					type: "requirement",
				};
				$http({method: "POST", url: "/newPagePair/", data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error creating a requirement:"); console.log(data); console.log(status);
				});

				pageService.masteryMap[data.parentId] = {pageId: data.parentId, isMet: true, isManuallySet: true};
				scope.page.requirementIds.push(data.parentId);
				scope.searchText = "";
			}

			// Process deleting requirements
			scope.deleteRequirement = function(requirementId) {
				var options = {
					parentId: requirementId,
					childId: scope.page.pageId,
					type: "requirement",
				};
				pageService.deletePagePair(options);
				scope.page.requirementIds.splice(scope.page.requirementIds.indexOf(options.parentId), 1);
			};

			// Toggle whether or not the user meets a requirement
			scope.toggleRequirement = function(requirementId) {
				pageService.updateMastery(scope, requirementId, !scope.hasMastery(requirementId));
			};
		},
	};
});

