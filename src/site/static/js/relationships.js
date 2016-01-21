"use strict";

// Directive for editing the parents, tags, requirements, or subjects.
app.directive("arbRelationships", function($q, $timeout, $interval, $http, pageService, userService, autocompleteService) {
	return {
		templateUrl: "static/html/relationships.html",
		scope: {
			pageId: "@",
			type: "@",
			readonly: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.editMap[$scope.pageId];

			// Helper variables
			$scope.isParentType = $scope.type === "parent";
			$scope.isTagType = $scope.type === "tag";
			$scope.isRequirementType = $scope.type === "requirement";
			$scope.isSubjectType = $scope.type === "subject";

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
			} else if ($scope.isSubjectType) {
				$scope.title = "Subjects";
				$scope.idsSource = $scope.page.subjectIds;
			}
			if ($scope.isLensRequirements) {
				$scope.title = "This version relies on:";
			}

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
				$scope.idsSource.push(data.parentId);
			};

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
		},
	};
});

