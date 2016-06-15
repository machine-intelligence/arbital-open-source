'use strict';

// Directive for editing the parents, tags, requirements, or subjects.
app.directive('arbRelationships', function($q, $timeout, $interval, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/relationships.html'),
		scope: {
			pageId: '@',
			type: '@',
			readonly: '=',
			// Optional. Id of the parent page for quick add
			quickParentId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.editMap[$scope.pageId];

			// Helper variables
			$scope.isParentType = $scope.type === 'parent';
			$scope.isTagType = $scope.type === 'tag';
			$scope.isRequirementType = $scope.type === 'requirement';
			$scope.isSubjectType = $scope.type === 'subject';

			// Compute various variables based on the type
			if ($scope.isParentType) {
				$scope.title = 'Parents';
				$scope.idsSource = $scope.page.parentIds;
			} else if ($scope.isTagType) {
				$scope.title = 'Tags';
				$scope.idsSource = $scope.page.taggedAsIds;
			} else if ($scope.isRequirementType) {
				$scope.title = 'Requirements';
				$scope.idsSource = $scope.page.requirementIds;
			} else if ($scope.isSubjectType) {
				$scope.title = 'Subjects';
				$scope.idsSource = $scope.page.subjectIds;
			}

			// Set up search
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				arb.autocompleteService.parentsSource({term: text}, function(results) {
					deferred.resolve(results);
				});
				return deferred.promise;
			};
			$scope.searchResultSelected = function(result) {
				if (!result) return;
				var params = {
					parentId: result.pageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.newPagePair(params, function success() {
					$scope.idsSource.push(params.parentId);
				});
			};

			$scope.relatesToItself = $scope.idsSource.indexOf($scope.pageId) >= 0;
			$scope.teachItself = function() {
				// Make the page teach itself. Used when the page isn't published yet.
				$scope.searchResultSelected({
					pageId: $scope.pageId,
				});
				$scope.relatesToItself = true;
			};

			// Process deleting a relationship
			$scope.deleteRelationship = function(otherPageId) {
				var params = {
					parentId: otherPageId,
					childId: $scope.page.pageId,
					type: $scope.type,
				};
				arb.pageService.deletePagePair(params, function success() {
					$scope.idsSource.splice($scope.idsSource.indexOf(params.parentId), 1);
					$scope.relatesToItself = $scope.idsSource.indexOf($scope.pageId) >= 0;
				});
			};

			$scope.addQuickParent = function() {
				$scope.searchResultSelected({pageId: $scope.quickParentId});
				$scope.quickParentId = undefined;
			};
		},
	};
});

