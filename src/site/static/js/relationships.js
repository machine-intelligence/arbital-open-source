"use strict";

// Directive for showing the parents, children, tags, or requirements.
app.directive("arbRelationships", function($q, $timeout, $interval, $http, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/relationships.html",
		scope: {
			pageId: "@",
			type: "@",
			isLensRequirements: "=",
			forceEditMode: "=",
			readonly: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			if ($scope.forceEditMode) {
				$scope.page = pageService.editMap[$scope.pageId];
			} else {
				$scope.page = pageService.pageMap[$scope.pageId];
			}
			$scope.inEditMode = $scope.forceEditMode;

			// Helper variables
			$scope.isParentType = $scope.type === "parent";
			$scope.isTagType = $scope.type === "tag";
			$scope.isRequirementType = $scope.type === "requirement";

			// Compute various variables based on the type
			if ($scope.isParentType) {
				$scope.title = "Parents";
				$scope.idsSource = $scope.page.parentIds;
			} else if ($scope.isTagType) {
				$scope.title = "Tags";
				$scope.idsSource = $scope.page.taggedAsIds;
			} else if ($scope.isRequirementType) {
				$scope.title = "Requirements";
				$scope.idsSource = $scope.page.requirementIds;
			}
			if ($scope.isLensRequirements) {
				$scope.title = "This version relies on:";
			}

			// Compute if we should show the panel
			$scope.showPanel = $scope.forceEditMode;

			// Check if the user has the given mastery.
			$scope.hasMastery = function(requirementId) {
				return pageService.masteryMap[requirementId].has;
			}

			// Check if the user meets all requirements
			$scope.meetsAllRequirements = function() {
				for (var n = 0; n < $scope.page.requirementIds.length; n++) {
					if (!$scope.hasMastery($scope.page.requirementIds[n])) {
						return false;
					}
				}
				return true;
			};

			// Do some custom stuff for requirements
			if ($scope.isRequirementType) {
				// Don't show the panel if the user has met all the requirements
				if (!$scope.forceEditMode) {
					$scope.showPanel = !$scope.meetsAllRequirements();
				}
	
				// Sort requirements
				$scope.page.requirementIds.sort(function(a, b) {
					return ($scope.hasMastery(a) ? 1 : 0) - ($scope.hasMastery(b) ? 1 : 0);
				});
			}

			// Toggle edit mode.
			$scope.inEditModeToggle = function() {
				$scope.inEditMode = !$scope.inEditMode;
			};

			// Set up search
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				autocompleteService.parentsSource({term: text}, function(results) {
					deferred.resolve(results);
				});
        return deferred.promise;
			};
			$scope.searchResultSelected = function(result) {
				if (!result) return;
				var data = {
					parentId: result.pageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				pageService.newPagePair(data);

				if ($scope.isRequirementType) {
					pageService.masteryMap[data.parentId] = {pageId: data.parentId, isMet: true, isManuallySet: true};
				}
				$scope.idsSource.push(data.parentId);
			}

			// Process deleting a relationship
			$scope.deleteRelationship = function(otherPageId) {
				var options = {
					parentId: otherPageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				pageService.deletePagePair(options);
				$scope.idsSource.splice($scope.idsSource.indexOf(options.parentId), 1);
			};

			// Toggle whether or not the user meets a requirement
			$scope.toggleRequirement = function(requirementId) {
				pageService.updateMastery($scope, requirementId, !$scope.hasMastery(requirementId));
			};
		},
	};
});

