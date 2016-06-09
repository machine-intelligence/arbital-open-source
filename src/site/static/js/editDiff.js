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

			if (!$scope.changeLog.edit) return;

			$scope.toggleDiff = function(update) {
				$scope.showDiff = !$scope.showDiff;
			};

			// Prepare to show the diff
			var pageId = $scope.changeLog.pageId;

			var thisEditText;
			var prevEditText;

			// Load thisEditText.
			var thisEditNum = $scope.changeLog.edit;
			arb.pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: thisEditNum,
				skipProcessDataStep: true,
				convertPageIdsToAliases: true,
				success: function(data) {
					var edit = data.edits[pageId];

					thisEditText = edit.text;


					// Load prevEditText.
					var prevEditNum = $scope.numEdits != 1 ? (thisEditNum - $scope.numEdits) : edit.prevEdit;
					arb.pageService.loadEdit({
						pageAlias: pageId,
						specificEdit: prevEditNum,
						skipProcessDataStep: true,
						convertPageIdsToAliases: true,
						success: function(data) {
							prevEditText = data.edits[pageId].text;

							// Make the diff
							$scope.diffHtml = arb.diffService.getDiffHtml(thisEditText, prevEditText);
						},
					});
				},
			});
		},
	};
});
