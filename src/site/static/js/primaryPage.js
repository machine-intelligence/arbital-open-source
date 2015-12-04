"use strict";

// Directive for the entire primary page.
app.directive("arbPrimaryPage", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/primaryPage.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.primaryPage;
			scope.page.answerIds.sort(pageService.getChildSortFunc("likes"));

			// Create the edit section for a new answer
			var createNewAnswer = function() {
				scope.newAnswerId = undefined;
				pageService.getNewPage({
					type: "answer",
					parentIds: [scope.page.pageId],
					success: function(newAnswerId) {
						pageService.loadEdit({
							pageAlias: newAnswerId,
							success: function() {
								scope.newAnswerId = newAnswerId;
							},
							error: function(error) {
								// TODO
							},
						});
					},
				});
			};
			createNewAnswer();

			// Called when the user is done editing the new answer
			scope.answerDone = function(result) {
				if (result.discard) {
					createNewAnswer();
				} else {
					window.location.href = pageService.getPageUrl(scope.page.pageId) + "#subpage-" + scope.newAnswerId;
					window.location.reload();
				}
			};

			// Called when the user selects an answer to suggest
			scope.suggestedAnswer = function(result) {
				if (!result) return;
				var data = {
					parentId: scope.page.pageId,
					childId: result.label,
					type: "parent",
				};
				pageService.newPagePair(data, function() {
					window.location.href = pageService.getPageUrl(scope.page.pageId) + "#subpage-" + result.label;
					window.location.reload();
				});
			};
		},
	};
});
