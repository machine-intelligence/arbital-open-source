'use strict';

// Directive for the Updates page.
app.directive('arbEditDiff', function($compile, $location, $rootScope, pageService, userService, diffService) {
	return {
		templateUrl: 'static/html/editDiff.html',
		scope: {
			changeLog: '=',
			numEdits: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.showDiff = false;
			$scope.thisEditText;
			$scope.prevEditText;

			$scope.numEdits = $scope.numEdits || 1;

			$scope.toggleDiff = function(update) {
				$scope.showDiff = !$scope.showDiff;
			};

			$scope.loadEditAndMaybeMakeDiff = function(changeLog, edit, successFn) {
				pageService.loadEdit({
					pageAlias: changeLog.pageId,
					specificEdit: edit,
					skipProcessDataStep: true,
					success: function(data, status) {
						successFn(data[changeLog.pageId]);
						if ($scope.thisEditText && $scope.prevEditText) {
							$scope.diffHtml = diffService.getDiffHtml($scope.thisEditText, $scope.prevEditText);
						}
					},
				});
			};

			$scope.loadEditAndMaybeMakeDiff($scope.changeLog, $scope.changeLog.edit, function(edit) {
				$scope.thisEditText = edit.text;
			});
			$scope.loadEditAndMaybeMakeDiff($scope.changeLog, $scope.changeLog.edit - $scope.numEdits, function(edit) {
				$scope.prevEditText = edit.text;
			});
		},
	};
});
