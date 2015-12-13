// EditPageDialogController is used for editing a page in an mdDialog
app.controller("EditPageDialogController", function ($scope, $mdDialog, $timeout, userService, pageService, parentIds, resumePageId) {
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

	$timeout(function() {
		// We need this class in the beginning to make sure dialog appears in the
		// correct position; but then we need to remove it, otherwise the body
		// is not visible.
		$("body").removeClass("md-dialog-is-showing");
	});
});
