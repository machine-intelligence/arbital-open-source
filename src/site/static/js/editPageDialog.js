// EditPageDialogController is used for editing a page in an mdDialog
app.controller("EditPageDialogController", function ($scope, $mdDialog, userService, pageService, parentIds, resumePageId) {
	$scope.pageService = pageService;
	$scope.userService = userService;

	// Load the page edit
	$scope.loadPageEdit = function(pageId) {
		pageService.loadEdit({
			pageAlias: pageId,
			success: function() {
				$scope.pageId = pageId;
			},
			error: function(error) {
				// TODO
			},
		});
	};

	// Create new comment
	if (!resumePageId) {
		pageService.getNewPage({
			type: "wiki",
			parentIds: parentIds, // injected from the caller
			success: function(newPageId) {
				$scope.loadPageEdit(newPageId);
			},
		});
	} else {
		$scope.loadPageEdit(resumePageId);
	}

	// Called when the user is done editing the page
	$scope.doneFn = function(result) {
		$mdDialog.hide(result);
	}

	// Called when the user closed the dialog
	$scope.hide = function() {
		$mdDialog.hide({hidden: true, pageId: $scope.pageId});
	};
});
