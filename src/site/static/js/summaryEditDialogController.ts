import app from './angular.ts';

// EditPageDialogController is used for editing a page in an mdDialog
app.controller('SummaryEditDialogController', function($scope, $mdDialog, $timeout, arb, page) {
	$scope.arb = arb;
	$scope.page = page;

	// Load the page edit
	arb.pageService.loadEdit({
		pageAlias: page.pageId,
		success: function() {
			$scope.pageId = page.pageId;
		},
	});

	// Called when the user is done editing the page
	$scope.doneFn = function(result) {
		arb.pageService.loadEdit({
			pageAlias: page.pageId,
			success: function() {
				$scope.pageId = page.pageId;
			},
		});
		$mdDialog.hide(result);
	};

	// Called when the user closed the dialog
	$scope.hide = function() {
		$mdDialog.hide({hidden: true, pageId: $scope.pageId});
	};

	$timeout(function() {
		// We need this class in the beginning to make sure dialog appears in the
		// correct position; but then we need to remove it, otherwise the body
		// is not visible.
		$('body').removeClass('md-dialog-is-showing');
	});
});
