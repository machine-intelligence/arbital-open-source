'use strict';

// Directive for rhsButtons that appear in $mdBottomSheet
app.controller('RhsButtonsController', function($scope, $mdMedia, $mdBottomSheet, pageService, userService) {
	$scope.pageService = pageService;
	$scope.userService = userService;
	$scope.isTinyScreen = !$mdMedia('gt-xs');

	$scope.newInlineComment = function(isEditorComment) {
		$mdBottomSheet.hide({func: 'newInlineComment', params: [isEditorComment]});
	};
	$scope.newEditorMark = function(markType) {
		$mdBottomSheet.hide({func: 'newEditorMark', params: [markType]});
	};
	$scope.newQueryMark = function() {
		$mdBottomSheet.hide({func: 'newQueryMark'});
	};

	$scope.isSubmenu = false;
	$scope.showEditorFeedbackSubmenu = function() {
		$scope.isSubmenu = true;
	}
});

