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
			}

			// Prepare to show the diff
			var pageId = $scope.changeLog.pageId;
			var thisEditNum = $scope.changeLog.edit;
			var prevEditNum = thisEditNum - ($scope.numEdits || 1);

			var thisEditText;
			var prevEditText;

			// Makes the diffHtml once both thisEditText and prevEditText have been loaded.
			function makeDiffIfBothTextsLoaded() {
				if (thisEditText && prevEditText) {
					$scope.diffHtml = arb.diffService.getDiffHtml(thisEditText, prevEditText);

					// Default to showing short diffs
					$scope.showDiff = $scope.diffHtml.length < 500;
				}
			}

			// Load thisEditText.
			arb.pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: thisEditNum,
				skipProcessDataStep: true,
				convertPageIdsToAliases: true,
				success: function(data) {
					if (!data.edits) return;
					thisEditText = data.edits[pageId].text;
					makeDiffIfBothTextsLoaded();
				},
			});

			// Load prevEditText.
			arb.pageService.loadEdit({
				pageAlias: pageId,
				specificEdit: prevEditNum,
				skipProcessDataStep: true,
				convertPageIdsToAliases: true,
				success: function(data) {
					if (!data.edits) return;
					prevEditText = data.edits[pageId].text;
					makeDiffIfBothTextsLoaded();
				},
			});
		},
	};
});
