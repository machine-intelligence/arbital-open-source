'use strict';

// Directive for showing a diff for a newEdit changeLog.
app.directive('arbEditDiff', function($compile, $location, $rootScope, arb) {
	return {
		templateUrl: 'static/html/editDiff.html',
		scope: {
			changeLog: '=',
			numEdits: '=', // Optional number of edits to group together in this diff. Defaults to 1.
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
						var prevEditNum = $scope.numEdits != 1 ? (thisEditNum - $scope.numEdits) : thisEdit.prevEdit;
						arb.pageService.loadEdit({
							pageAlias: pageId,
							specificEdit: prevEditNum,
							skipProcessDataStep: true,
							convertPageIdsToAliases: true,
							success: function(data) {
								var prevEdit = data.edits[pageId];

								// Make the diff
								$scope.diffHtml = arb.diffService.getDiffHtml(thisEdit, prevEdit);
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
