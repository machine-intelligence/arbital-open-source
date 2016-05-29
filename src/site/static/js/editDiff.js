'use strict';

// Directive for showing a diff for a newEdit changeLog.
app.directive('arbEditDiff', function($compile, $location, $rootScope, arb, diffService) {
	return {
		templateUrl: 'static/html/editDiff.html',
		scope: {
			changeLog: '=',
			numEdits: '=', // Optional number of edits to group together in this diff. Defaults to 1.
		},
		controller: function($scope) {
			$scope.arb = arb;
			

			$scope.showDiff = false;

			$scope.toggleDiff = function(update) {
				$scope.showDiff = !$scope.showDiff;

				if (!$scope.showDiff || $scope.diffHtml) {
					return;
				}

				var pageId = $scope.changeLog.pageId;
				var thisEditNum = $scope.changeLog.edit;
				var prevEditNum = thisEditNum - ($scope.numEdits || 1);

				var thisEditText;
				var prevEditText;

				// Makes the diffHtml once both thisEditText and prevEditText have been loaded.
				function makeDiffIfBothTextsLoaded() {
					if (thisEditText && prevEditText) {
						$scope.diffHtml = diffService.getDiffHtml(thisEditText, prevEditText);
					}
				}

				// Load thisEditText.
				pageService.loadEdit({
					pageAlias: pageId,
					specificEdit: thisEditNum,
					skipProcessDataStep: true,
					success: function(data) {
						thisEditText = data[pageId].text;
						makeDiffIfBothTextsLoaded();
					},
				});

				// Load prevEditText.
				pageService.loadEdit({
					pageAlias: pageId,
					specificEdit: prevEditNum,
					skipProcessDataStep: true,
					success: function(data) {
						prevEditText = data[pageId].text;
						makeDiffIfBothTextsLoaded();
					},
				});
			};
		},
	};
});
