'use strict';

// Directive for the Updates page.
app.directive('arbUpdates', function($compile, $location, $rootScope, pageService, userService, diffService) {
	return {
		templateUrl: 'static/html/updatesPage.html',
		scope: {
			updateGroups: '=',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.updateGroups.sort(function(a, b) {
				if (b.key.unseen !== a.key.unseen) {
					return (b.key.unseen ? 1 : 0) - (a.key.unseen ? 1 : 0);
				}
				if (b.mostRecentDate === a.mostRecentDate) return 0;
				return b.mostRecentDate < a.mostRecentDate ? -1 : 1;
			});

			$scope.toggleDiff = function(update) {
				update.showDiff = !update.showDiff;
			};

			$scope.loadEditAndMaybeMakeDiff = function(updateGroup, update, edit, successFn) {
				pageService.loadEdit({
					pageAlias: updateGroup.key.groupByPageId,
					specificEdit: edit,
					skipProcessDataStep: true,
					success: function(data, status) {
						successFn(data[updateGroup.key.groupByPageId]);
						if (update.thisEditText && update.prevEditText) {
							update.diffHtml = diffService.getDiffHtml(update.thisEditText, update.prevEditText);
						}
					},
				});
			};

			// Load diffs for the pageEdit updates.
			$scope.updateGroups.forEach(function(updateGroup) {
				updateGroup.updates.forEach(function(update) {
					if (update.type == 'pageEdit' && update.changeLog && update.changeLog.edit) {
						var thisEditText;
						var prevEditText;

						$scope.loadEditAndMaybeMakeDiff(updateGroup, update, update.changeLog.edit, function(edit) {
							update.thisEditText = edit.text;
						});
						$scope.loadEditAndMaybeMakeDiff(updateGroup, update, update.changeLog.edit - update.repeated, function(edit) {
							update.prevEditText = edit.text;
						});
					}
				});
			});
		},
	};
});
