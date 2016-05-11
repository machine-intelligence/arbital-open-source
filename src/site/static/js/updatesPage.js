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

			// Load diffs for the pageEdit updates.
			$scope.updateGroups.forEach(function(updateGroup) {
				updateGroup.updates.forEach(function(update) {
					if (update.type == 'pageEdit' && update.changeLog && update.changeLog.edit) {
						var thisEditText;
						var prevEditText;

						function loadEditAndMaybeMakeDiff(edit, successFn) {
							pageService.loadEdit({
								pageAlias: updateGroup.key.groupByPageId,
								specificEdit: edit,
								skipProcessDataStep: true,
								success: function(data, status) {
									successFn(data[updateGroup.key.groupByPageId]);
									if (thisEditText && prevEditText) {
										update.diffHtml = diffService.getDiffHtml(thisEditText, prevEditText);
									}
								},
							});
						}

						loadEditAndMaybeMakeDiff(update.changeLog.edit, function(edit) {
							thisEditText = edit.text;
						});
						loadEditAndMaybeMakeDiff(update.changeLog.edit - update.repeated, function(edit) {
							prevEditText = edit.text;
						});
					}
				});
			});
		},
	};
});
