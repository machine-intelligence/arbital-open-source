'use strict';

// Directive for rhsButtons that appear in $mdBottomSheet
app.controller('RhsButtonsController', function($scope, $mdMedia, $mdBottomSheet, arb) {
	$scope.pageService = pageService;
	$scope.userService = userService;
	$scope.isTinyScreen = !$mdMedia('gt-xs');

	$scope.newInlineComment = function(isEditorComment) {
		$mdBottomSheet.hide({func: 'newInlineComment'});
	};
	$scope.newEditorMark = function(markType) {
		$mdBottomSheet.hide({func: 'newEditorMark', params: [markType]});
	};
	$scope.newQueryMark = function() {
		$mdBottomSheet.hide({func: 'newQueryMark'});
	};
});

