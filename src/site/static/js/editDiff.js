'use strict';

// Directive for showing a diff for a newEdit changeLog.
app.directive('arbEditDiff', function($compile, $location, $rootScope, arb) {
	return {
		templateUrl: versionUrl('static/html/editDiff.html'),
		scope: {
			changeLog: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.showDiff = false;

			// Fetch the necessary edits from server to do the diff
			var computeDiffHtml = function() {
				// Prepare to show the diff
				var pageId = $scope.changeLog.pageId;

				// Load thisEdit.
				var thisEditNum = $scope.changeLog.edit;
				arb.pageService.loadEdit({
					pageAlias: pageId,
					specificEdit: thisEditNum,
					skipProcessDataStep: true,
					convertPageIdsToAliases: true,
					success: function(data) {
						var thisEdit = data.edits[pageId];

						// Load prevEdit.
						arb.pageService.loadEdit({
							pageAlias: pageId,
							specificEdit: thisEdit.prevEdit,
							skipProcessDataStep: true,
							convertPageIdsToAliases: true,
							success: function(data) {
								var prevEdit = data.edits[pageId];

								// Make the diff
								$scope.diffHtml = arb.diffService.getDiffHtml(prevEdit, thisEdit);
							},
						});
					},
				});
			};

			$scope.toggleDiff = function(update) {
				$scope.showDiff = !$scope.showDiff;

				if ($scope.showDiff && !$scope.diffHtml) {
					computeDiffHtml();
				}
			};
		},
	};
});
